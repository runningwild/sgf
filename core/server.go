package core

import (
	"encoding/gob"
	"fmt"
	"github.com/runningwild/sluice"
	"io"
	"sync"
)

func init() {
	fmt.Printf("")
}

type Game interface{}

type Update interface {
	ApplyUpdate(game Game)
}
type Request interface {
	ApplyRequest(game Game) []Update
}

type completeGameState struct {
	Game Game
}

func (cgs completeGameState) Apply(game *Game) {
	*game = cgs.Game
}

type joinGameRequest struct{}

func (joinGameRequest) Apply(game *Game) []Update { return nil }

type Common struct {
	Game      Game
	GameMutex sync.RWMutex

	GameRegistry    TypeRegistry
	RequestRegistry TypeRegistry
	UpdateRegistry  TypeRegistry
}

func (c *Common) RegisterGame(t interface{}) {
	if _, ok := t.(Game); !ok {
		panic("RegisterGame() can only be called with values of type Game.")
	}
	c.GameRegistry.Register(t)
}
func (c *Common) RegisterRequest(t interface{}) {
	if _, ok := t.(Request); !ok {
		panic("RegisterRequest() can only be called with values of type Request.")
	}
	c.RequestRegistry.Register(t)
}
func (c *Common) RegisterUpdate(t interface{}) {
	if _, ok := t.(Update); !ok {
		panic("RegisterUpdate() can only be called with values of type Update.")
	}
	c.UpdateRegistry.Register(t)
}

type Host struct {
	Common
	Comm *sluice.Host

	majorUpdates          chan Update
	majorUpdatesCollector chan Update
}

func MakeHost(addr string, port int) (*Host, error) {
	configs := []sluice.StreamConfig{
		sluice.StreamConfig{
			Name:      "MinorUpdates",
			Broadcast: true,
			Mode:      sluice.ModeUnreliableOrdered,
		},
		sluice.StreamConfig{
			Name:      "Requests",
			Broadcast: false,
			Mode:      sluice.ModeReliableOrdered,
		},

		// Major updates are targeted and sent from the host to the clients.
		// The first update that a client receives will contain the complete
		// game state.
		sluice.StreamConfig{
			Name:      "MajorUpdates",
			Broadcast: false,
			Mode:      sluice.ModeReliableOrdered,
		},
	}
	comm, err := sluice.MakeHost(addr, port, configs)
	if err != nil {
		return nil, err
	}
	host := &Host{
		Comm:                  comm,
		majorUpdates:          make(chan Update),
		majorUpdatesCollector: make(chan Update),
	}
	return host, nil
}

func (common *Common) Start() {
	common.GameRegistry.Register(completeGameState{})
	common.GameRegistry.Complete()
	common.RequestRegistry.Register(joinGameRequest{})
	common.RequestRegistry.Complete()
	common.UpdateRegistry.Complete()
}

func (host *Host) Start(game Game) {
	host.Game = game
	host.Common.Start()
	go host.handleRequestsAndUpdates()
}

func InfinitelyBufferUpdates(in <-chan Update, out chan<- Update) {
	var updates []Update
	var nextUpdate Update
	var send chan<- Update
	for {
		select {
		case update, ok := <-in:
			if !ok {
				if send != nil {
					send <- nextUpdate
					for _, update := range updates {
						send <- update
					}
				}
				close(out)
				return
			}
			if send == nil {
				nextUpdate = update
				send = out
			} else {
				updates = append(updates, update)
			}

		case send <- nextUpdate:
			if len(updates) > 0 {
				nextUpdate = updates[0]
				updates = updates[1:]
			} else {
				send = nil
			}
		}
	}
}

func (host *Host) handleRequestsAndUpdates() {
	newReaders := host.Comm.GetReadersChan("Requests")
	requestCollector := make(chan Request)
	var nodes []int
	go InfinitelyBufferUpdates(host.majorUpdates, host.majorUpdatesCollector)
	for {
		select {
		case reader := <-newReaders:
			// Given a new reader, make sure that the first thing it sent was a
			// joinGameRequest, then sent it the game state,
			val, err := host.RequestRegistry.Decode(reader)
			if err != nil {
				// TODO: Log this, don't panic
				panic(err)
			}
			_, ok := val.(joinGameRequest)
			if !ok {
				// TODO: don't panic
				panic("Not a join request.")
			}
			nodes = append(nodes, reader.NodeId())
			go func(reader sluice.StreamReader) {
				for {
					val, err := host.RequestRegistry.Decode(reader)
					if err != nil {
						// TODO: Obviously don't panic
						panic(err)
					}
					requestCollector <- val.(Request)
				}
			}(reader)

			writer := host.Comm.GetDirectedWriter("MajorUpdates", reader.NodeId())
			host.GameMutex.RLock()
			err = host.GameRegistry.Encode(completeGameState{host.Game}, writer)
			if err != nil {
				panic(err)
			}
			host.GameMutex.RUnlock()
			nodes = append(nodes, reader.NodeId())

		case request := <-requestCollector:
			host.GameMutex.Lock()
			updates := request.ApplyRequest(&host.Game)
			host.GameMutex.Unlock()
			for _, update := range updates {
				host.GameMutex.Lock()
				update.ApplyUpdate(&host.Game)
				host.GameMutex.Unlock()
				for _, node := range nodes {
					writer := host.Comm.GetDirectedWriter("MajorUpdates", node)
					err := host.UpdateRegistry.Encode(update, writer)
					if err != nil {
						// TODO: obviously don't panic
						panic(err)
					}
				}
			}

		case update := <-host.majorUpdates:
			host.GameMutex.Lock()
			update.ApplyUpdate(&host.Game)
			host.GameMutex.Unlock()
			for _, node := range nodes {
				writer := host.Comm.GetDirectedWriter("MajorUpdates", node)
				err := host.UpdateRegistry.Encode(update, writer)
				if err != nil {
					// TODO: obviously don't panic
					panic(err)
				}
			}
		}
	}
}

type Client struct {
	Common
	Comm *sluice.Client

	AllRemoteUpdates chan interface{}

	MinorUpdatesChan chan Update
	MinorUpdatesEnc  *gob.Encoder
	RequestsChan     chan Request
	RequestsEnc      *gob.Encoder
}

func MakeClient(addr string, port int) (*Client, error) {
	comm, err := sluice.MakeClient(addr, port)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Comm:             comm,
		AllRemoteUpdates: make(chan interface{}),
		MinorUpdatesChan: make(chan Update),
		MinorUpdatesEnc:  gob.NewEncoder(comm.GetWriter("MinorUpdates")),
		RequestsChan:     make(chan Request),
		RequestsEnc:      gob.NewEncoder(comm.GetWriter("Requests")),
	}
	return client, err
}

func (client *Client) Start() {
	client.Common.Start()
	err := client.RequestRegistry.Encode(joinGameRequest{}, client.Comm.GetWriter("Requests"))
	if err != nil {
		// TODO: Obviously don't panic
		panic(err)
	}
	majorUpdatesChan := client.Comm.GetReadersChan("MajorUpdates")
	majorUpdates := <-majorUpdatesChan
	update, err := client.GameRegistry.Decode(majorUpdates)
	if err != nil {
		// TODO: Obviously don't panic
		panic(err)
	}
	cgs, ok := update.(completeGameState)
	if !ok {
		// TODO: Obviously don't panic
		panic("Not a completeGameState")
	}
	client.Game = cgs.Game

	go client.primaryRoutine()

	// This will launch go routines to collect all remote events that we can
	// receive and send them all along the same channel.  All of the events
	// will be contained in the same interface so this works fine.
	for _, name := range []string{"MinorUpdates", "MajorUpdates"} {
		go collector(&client.UpdateRegistry, client.Comm.GetReadersChan(name), client.AllRemoteUpdates)
	}

}

func collector(registry *TypeRegistry, readers <-chan sluice.StreamReader, objs chan<- interface{}) {
	for reader := range readers {
		go func(reader io.Reader) {
			for {
				obj, err := registry.Decode(reader)
				if err != nil {
					// TODO: Grace
					panic(err)
				}
				objs <- obj
			}
		}(reader)
	}
}

func (c *Client) primaryRoutine() {
	for {
		select {
		case update := <-c.AllRemoteUpdates:
			update.(Update).ApplyUpdate(&c.Game)

		case update := <-c.MinorUpdatesChan:
			err := c.MinorUpdatesEnc.Encode(update)
			if err != nil {
				// TODO: Handle this more gracefully
				panic(err)
			}
			update.ApplyUpdate(&c.Game)

		case request := <-c.RequestsChan:
			err := c.RequestsEnc.Encode(request)
			if err != nil {
				// TODO: Handle this more gracefully
				panic(err)
			}
		}
	}
}

func (c *Client) MakeMinorUpdate(update Update) {
	c.MinorUpdatesChan <- update
}
func (c *Client) MakeRequest(request Request) {
	c.RequestsChan <- request
}
