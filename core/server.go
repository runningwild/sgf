package core

import (
	"encoding/gob"
	"github.com/runningwild/sluice"
	"io"
)

type Game interface{}

type Update interface {
	Apply(game Game)
}
type Request interface {
	Apply(game Game) []Update
}

type Common struct {
	Game Game
}

type Host struct {
	Common
	Comm            *sluice.Host
	Requests        chan interface{}
	MajorUpdatesEnc *gob.Encoder
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
		Comm: comm,
	}
	go host.primaryRoutine()
	go collector(host.Comm.GetReadersChan("Requests"), host.Requests)
	return host, nil
}

func (host *Host) primaryRoutine() {
	for {
		select {
		case request := <-host.Requests:
			updates := request.(Request).Apply(host.Game)
			for _, update := range updates {
				update.Apply(host.Game)
				host.MajorUpdatesEnc.Encode(update)
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
	go client.primaryRoutine()

	// This will launch go routines to collect all remote events that we can
	// receive and send them all along the same channel.  All of the events
	// will be contained in the same interface so this works fine.
	for _, name := range []string{"MinorUpdates", "MajorUpdates"} {
		go collector(client.Comm.GetReadersChan(name), client.AllRemoteUpdates)
	}

	return client, err
}

func collector(readers <-chan sluice.StreamReader, objs chan<- interface{}) {
	for reader := range readers {
		go func(reader io.Reader) {
			dec := gob.NewDecoder(reader)
			for {
				var obj interface{}
				err := dec.Decode(&obj)
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
			update.(Update).Apply(c.Game)

		case update := <-c.MinorUpdatesChan:
			err := c.MinorUpdatesEnc.Encode(update)
			if err != nil {
				// TODO: Handle this more gracefully
				panic(err)
			}
			update.Apply(c.Game)

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
