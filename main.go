package sgf

type Game interface {
	HostThink()
	ClientThink()
}

// MinorEvent
// MajorEvent
// Request
type MinorUpdate interface {
	Apply(game Game)
}
type MajorUpdate interface {
	Apply(game Game)
}
type MinorRequest interface {
	Apply(game Game)
}
type MajorRequest interface {
	Apply(game Game)
}

type Engine struct{}

func (e *Engine) SendRequest(request Request) {}

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
