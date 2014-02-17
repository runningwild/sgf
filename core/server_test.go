package core_test

import (
	"fmt"
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/sgf/core"
	"time"
)

type Game struct {
	Players []Player
}
type Player struct {
	Pos int
}
type Advance struct {
	Index  int
	NewPos int
}

func (a Advance) ApplyRequest(_game core.Game) []core.Update {
	fmt.Printf("Advance: ApplyRequest\n")
	game := _game.(*Game)
	if a.Index < 0 || a.Index >= len(game.Players) {
		return nil
	}
	fmt.Printf("Applied: %v\n", game)
	return []core.Update{
		Advance{
			Index:  a.Index,
			NewPos: game.Players[a.Index].Pos + 1,
		},
	}
}
func (a Advance) ApplyUpdate(_game core.Game) {
	fmt.Printf("Advance: ApplyUpdate\n")
	game := _game.(*Game)
	if a.Index < 0 || a.Index >= len(game.Players) {
		return
	}
	game.Players[a.Index].Pos = a.NewPos
	fmt.Printf("Applied: %v\n", game)
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
		game := &Game{
			Players: []Player{
				Player{Pos: 1},
				Player{Pos: 2},
			},
		}
		host.Start(game)
		client0.Start()
		client1.Start()

		client1.GameMutex.RLock()
		fmt.Printf("game: %v\n", client1.Game)
		client1.GameMutex.RUnlock()

		adv := Advance{0, 2}
		client0.MakeRequest(adv)
		// client0.MakeMinorUpdate(adv)
		time.Sleep(time.Millisecond * 100)

		client1.GameMutex.RLock()
		fmt.Printf("game: %v\n", client1.Game)
		client1.GameMutex.RUnlock()

		c.Expect(true, gospec.Equals, false)
	})
}
