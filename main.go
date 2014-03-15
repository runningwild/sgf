package sgf

import (
	"github.com/runningwild/sgf/core"
)

// These are copied from core

type Game interface{}
type Update interface {
	// ApplyUpdate can modify the game state.  node will be -1 if the request
	// came from this client, 0 if it came from the host, and otherwise it will
	// be the sluice node of the client that it came from.
	ApplyUpdate(node int, game Game)
}
type Request interface {
	ApplyRequest(node int, game Game) []Update
}

type Engine struct{}

func (e *Engine) SendRequest(request core.Request) {}

// This is a Read-Lock, so multiple go-routines can all Pause() simultaneously.
func (e *Engine) Pause() {}

func (e *Engine) Unpause() {}

func (e *Engine) GetState() Game {}

// Returns the Id of this engine.  Every engine connected in a game has a unique
// id.
func (e *Engine) Id() int64 {}

// If this is the Host engine this function will return a list of the ids of all
// engines currently connected, including this engine.  If this is a client
// engine this function will return nil.
func (e *Engine) Ids() []int64 {}

func (e *Engine) IsHost() bool {}

func (e *Engine) Kill() {}
