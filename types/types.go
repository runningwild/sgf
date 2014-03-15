package types

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
