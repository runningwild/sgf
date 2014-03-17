package sgf

import (
	"github.com/runningwild/sgf/core"
	"github.com/runningwild/sgf/types"
)

type Engine interface {
	RegisterGame(v interface{})
	RegisterUpdate(v interface{})
	RegisterRequest(v interface{})
	RLock()
	RUnlock()
	Game() types.Game
}

type HostEngine interface {
	Engine
	Start(initialState types.Game)
}

type host struct {
	*core.Host
}

func MakeHost(addr string, port int) (HostEngine, error) {
	coreHost, err := core.MakeHost(addr, port)
	if err != nil {
		return nil, err
	}
	var h host
	h.Host = coreHost
	return &h, nil
}
func (h *host) RLock() {
	h.Host.GameMutex.RLock()
}
func (h *host) RUnlock() {
	h.Host.GameMutex.RUnlock()
}
func (h *host) Game() types.Game {
	return h.Host.Game
}

type ClientEngine interface {
	Engine
	Start()
}

type Client struct {
	*core.Client
}

func MakeClient(addr string, port int) (ClientEngine, error) {
	coreClient, err := core.MakeClient(addr, port)
	if err != nil {
		return nil, err
	}
	var c Client
	c.Client = coreClient
	return &c, nil
}
func (c *Client) RLock() {
	c.Client.GameMutex.RLock()
}
func (c *Client) RUnlock() {
	c.Client.GameMutex.RUnlock()
}
func (c *Client) Game() types.Game {
	return c.Client.Game
}
