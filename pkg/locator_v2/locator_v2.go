package locator_v2

import (
	"context"
	"errors"
	"fmt"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/gorilla/websocket"
)

// All currently active Lobbies
var Lobbies = map[string]*Lobby{}

var upgrader = websocket.Upgrader{}
var contextbg = context.Background()

func SetupTest() {

	// log.Println("SetupTEST enabled.")

	ctx, cancelCtx := context.WithCancel(contextbg)
	testplayer1 := NewPlayer(ctx, cancelCtx, &Lobby{}, "test1")
	ctx, cancelCtx = context.WithCancel(contextbg)
	testplayer2 := NewPlayer(ctx, cancelCtx, &Lobby{}, "test2")
	ctx, cancelCtx = context.WithCancel(contextbg)
	testplayer3 := NewPlayer(ctx, cancelCtx, &Lobby{}, "test3")

	g, err := getGame("capitals")
	if err != nil {
		util.Sugar.Fatal(err)
	}
	// b := NewLobby(10, 10, 10, g)
	// Lobbies[b.LobbyID] = b
	Lobbies["AAAAAAAA"] = &Lobby{
		LobbyID: "AAAAAAAA",
		// owner:      testplayer1,
		started:    false,
		RoundTime:  60,
		ReviewTime: 15,
		Rounds:     10,
		player:     make(map[string]*Player),
		add:        make(chan *Player, 10),
		drop:       make(chan *Player, 10),
		ready:      make(chan *Player, 10),
		submitted:  make(chan *Player, 10),
		game:       g,
		state:      "startup",
		nextState:  "guessing",
		location:   "",
		locations:  []string{},
	}

	l := Lobbies["AAAAAAAA"]

	testplayer1.lobby = l
	testplayer2.lobby = l
	testplayer3.lobby = l
	l.player["test1"] = testplayer1
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
