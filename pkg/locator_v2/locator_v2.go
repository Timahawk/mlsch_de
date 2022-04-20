package locator_v2

import (
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

// All currently active Lobbies
var Lobbies = map[string]*Lobby{}

var upgrader = websocket.Upgrader{}

// getLobby helper function to get the Lobby, if exists.
func getLobby(room string) (*Lobby, error) {

	if lobby, ok := Lobbies[room]; ok {
		return lobby, nil
	}
	return nil, errors.New(fmt.Sprintln("room not found for Room", room))
}
