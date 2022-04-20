package locator_v2

import (
	"errors"
	"fmt"

	"github.com/Timahawk/mlsch_de/pkg/util"
)

// Lobby maintains the set of active Player and broadcasts messages to the
// clients. It is the dreh & angelpunkt of the ganze Veranstaltung.
type Lobby struct {
	// Hub ID
	LobbyID string

	// Who creates the lobby. and starts the Lobby.
	owner *Player

	// This determines the weather the lobby is started.
	started bool

	// Determines the duration of a guessing round.
	RoundTime int

	// Registered clients.
	player map[string]*Player

	// Register requests from the clients.
	add chan *Player

	// Unregister requests from clients.
	remove chan *Player
}

// NewLobby creates a new Lobby.
func NewLobby(zeit int, game *Game, owner *Player) *Lobby {
	id := util.RandString(8)

	lobby := Lobby{
		LobbyID:   id,
		owner:     owner,
		started:   false,
		RoundTime: zeit,
		player:    make(map[string]*Player),
		add:       make(chan *Player, 5),
		remove:    make(chan *Player, 5),
	}

	Lobbies[id] = &lobby

	util.Sugar.Debugw("New Lobby created.",
		"id", id,
		"time", zeit,
		"state", 3,
		"nextState", 0,
		"roundCounter", 0,
		// "game", game.Name,
	)
	return &lobby
}

func (l *Lobby) serveWaitRoom() {
	for {
		select {
		case p := <-l.add:
			l.player[p.Name] = p

		case p := <-l.remove:
			delete(l.player, p.Name)
		}
	}
}

func (l *Lobby) getUser(name string) (*Player, error) {
	if p, ok := l.player[name]; ok {
		return p, nil
	}
	return nil, errors.New(fmt.Sprintln(name, "not found for Lobby", l.LobbyID))
}
