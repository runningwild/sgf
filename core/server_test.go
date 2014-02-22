package core_test

import (
	"fmt"
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/sgf/core"
	"time"
)

type Game struct {
	Players []Player
	myName  string
	myIndex int
	me      *Player
}
type Player struct {
	Name string
	Node int
	Pos  int
}
type Advance struct {
	Index  int
	NewPos int
}

func (g *Game) getIndex(node int, index int) int {
	if node == -1 || node == 0 {
		return index
	}
	for i := range g.Players {
		if g.Players[i].Node == node {
			return i
		}
	}
	return -1
}

func (a Advance) ApplyRequest(node int, _game core.Game) []core.Update {
	fmt.Printf("Advance: ApplyRequest\n")
	game := _game.(*Game)
	a.Index = game.getIndex(node, a.Index)
	if a.Index == -1 {
		return nil
	}
	a.NewPos = game.Players[a.Index].Pos + 1
	return []core.Update{a}
}
func (a Advance) ApplyUpdate(node int, _game core.Game) {
	game := _game.(*Game)
	index := game.getIndex(node, a.Index)
	if index == -1 {
		return
	}
	game.Players[index].Pos = a.NewPos
}

type Join struct {
	Name string
	Node int
}

func (j Join) ApplyRequest(node int, _game core.Game) []core.Update {
	game := _game.(*Game)
	if j.Name == "" {
		return nil
	}
	j.Node = node
	for _, player := range game.Players {
		if player.Name == j.Name {
			return nil
		}
	}
	return []core.Update{j}
}
func (j Join) ApplyUpdate(node int, _game core.Game) {
	game := _game.(*Game)
	if j.Name == "" {
		return
	}
	fmt.Printf("Join(%d): %v\n", node, j)
	if node != 0 {
		return
	}
	game.Players = append(game.Players, Player{j.Name, j.Node, 0})
	if game.myName == j.Name {
		game.myIndex = len(game.Players) - 1
		game.me = &game.Players[game.myIndex]
	}
}

// type Update interface {
// 	Apply(game *Game)
// }
// type Request interface {
// 	Apply(game *Game) []Update
// }

type registerer interface {
	RegisterGame(interface{})
	RegisterRequest(interface{})
	RegisterUpdate(interface{})
}

func registerGameForAll(t interface{}, rs ...registerer) {
	for _, r := range rs {
		r.RegisterGame(t)
	}
}
func registerRequestForAll(t interface{}, rs ...registerer) {
	for _, r := range rs {
		r.RegisterRequest(t)
	}
}
func registerUpdateForAll(t interface{}, rs ...registerer) {
	for _, r := range rs {
		r.RegisterUpdate(t)
	}
}

func SimpleServerSpec(c gospec.Context) {
	c.Specify("Hook up all of the basic parts and make them talk.", func() {
		host, err := core.MakeHost("127.0.0.1", 1234)
		c.Assume(err, gospec.Equals, error(nil))
		client0, err := core.MakeClient("127.0.0.1", 1234)
		c.Assume(err, gospec.Equals, error(nil))
		client1, err := core.MakeClient("127.0.0.1", 1234)
		c.Assume(err, gospec.Equals, error(nil))
		fmt.Printf("%v %v %v\n", host, client0)
		registerGameForAll(&Game{}, host, client0, client1)
		registerRequestForAll(Advance{}, host, client0, client1)
		registerUpdateForAll(Advance{}, host, client0, client1)
		registerRequestForAll(Join{}, host, client0, client1)
		registerUpdateForAll(Join{}, host, client0, client1)
		game := &Game{Players: []Player{}}
		host.Start(game)
		client0.Start()
		client1.Start()

		client0.GameMutex.Lock()
		game0 := client0.Game.(*Game)
		game0.myName = "foo"
		client0.GameMutex.Unlock()
		client0.MakeRequest(Join{Name: "foo"})

		client1.GameMutex.Lock()
		game1 := client1.Game.(*Game)
		game1.myName = "bar"
		client1.GameMutex.Unlock()
		client1.MakeRequest(Join{Name: "bar"})

		time.Sleep(time.Millisecond * 100)

		client0.GameMutex.RLock()
		fmt.Printf("game0: %v\n", client0.Game)
		client0.GameMutex.RUnlock()
		client1.GameMutex.RLock()
		fmt.Printf("game1: %v\n", client1.Game)
		client1.GameMutex.RUnlock()

		incs := 1000
		done := make(chan struct{})
		for _, client := range []*core.Client{client0, client1} {
			go func(client *core.Client) {
				for i := 0; i < incs; i++ {
					time.Sleep(time.Millisecond)
					client.GameMutex.RLock()
					g := client.Game.(*Game)
					adv := Advance{g.myIndex, g.me.Pos + 1}
					fmt.Printf("Request(%d): %d\n", i, adv.NewPos)
					client.GameMutex.RUnlock()
					client.MakeRequest(adv)
					client.MakeMinorUpdate(adv)
				}
				done <- struct{}{}
			}(client)
		}

		<-done
		<-done
		time.Sleep(time.Millisecond * 100)

		client0.GameMutex.RLock()
		fmt.Printf("game0: %v\n", client0.Game)
		client0.GameMutex.RUnlock()
		client1.GameMutex.RLock()
		fmt.Printf("game1: %v\n", client1.Game)
		client1.GameMutex.RUnlock()

		// // for _, client := range []*core.Client{client0, client1} {

		// // }

		// client1.GameMutex.RLock()
		// fmt.Printf("game: %v\n", client1.Game)
		// client1.GameMutex.RUnlock()

		// adv := Advance{0, 2}
		// client0.MakeRequest(adv)
		// // client0.MakeMinorUpdate(adv)
		// time.Sleep(time.Millisecond * 100)

		// client1.GameMutex.RLock()
		// fmt.Printf("game: %v\n", client1.Game)
		// client1.GameMutex.RUnlock()

		c.Expect(true, gospec.Equals, false)
	})
}
