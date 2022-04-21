package locator_v2

import (
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

// All currently active Lobbies
var Lobbies = map[string]*Lobby{}
var upgrader = websocket.Upgrader{}

func init() {
	// TODO if production dont do that.
	testplayer := &Player{&Lobby{}, "test", nil, false, make(chan string), false}
	testplayer2 := &Player{&Lobby{}, "test2", nil, false, make(chan string), false}
	testplayer3 := &Player{&Lobby{}, "test3", nil, false, make(chan string), false}

	Lobbies["AAAAAAAA"] = &Lobby{
		LobbyID:   "AAAAAAAA",
		owner:     testplayer,
		started:   false,
		RoundTime: 30,
		player:    make(map[string]*Player),
		add:       make(chan *Player, 10),
		drop:      make(chan *Player, 10),
		ready:     make(chan *Player, 10),
	}

	l := Lobbies["AAAAAAAA"]

	testplayer.lobby = l
	testplayer2.lobby = l
	testplayer3.lobby = l
	l.player["test"] = testplayer
	l.player["test2"] = testplayer2
	l.player["test3"] = testplayer3

	go l.serveWaitRoom()

}

// getLobby helper function to get the Lobby, if exists.
func getLobby(room string) (*Lobby, error) {

	if lobby, ok := Lobbies[room]; ok {
		return lobby, nil
	}
	return nil, errors.New(fmt.Sprintln("room not found for Room", room))
}
