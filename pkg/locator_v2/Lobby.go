package locator_v2

import (
	"errors"
	"fmt"
	"time"

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
	drop chan *Player

	// Receives calls that players are ready
	ready chan *Player
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
		add:       make(chan *Player, 10),
		drop:      make(chan *Player, 10),
		ready:     make(chan *Player, 10),
	}

	Lobbies[id] = &lobby

	util.Sugar.Infow("New Lobby created.",
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
	defer func() {
		util.Sugar.Infow("serveLobby stopped",
			"Lobby", l.LobbyID)
	}()

	timer := new(time.Timer)

	// Das macht einen nilPointer error??
	// util.Sugar.Infow("serveLobby started",
	// 	"Lobby", l.LobbyID)
	for {
		select {
		case p := <-l.add:
			// if _, err := l.getPlayer(p.Name); err == nil {
			// 	util.Sugar.Infow("Add but Player already existed.",
			// 		"Lobby", l.LobbyID,
			// 		"Player", p.Name)
			// 	l.player[p.Name] = p
			// 	continue
			// }
			util.Sugar.Infow("Add",
				"Lobby", l.LobbyID,
				"Player", p.Name)
			l.player[p.Name] = p
			p.connected = true

			for _, pl := range l.player {
				if pl.conn != nil && pl.connected == true {
					pl.toConn <- fmt.Sprintf("%s joined the lobby.", p.Name)
				}
			}
		case p := <-l.drop:
			util.Sugar.Infow("Remove",
				"Lobby", l.LobbyID,
				"Player", p.Name)
			// delete(l.player, p.Name)
			p.connected = false
			p.ready = false
			p.ctxcancel()
			for _, pl := range l.player {
				if pl.conn != nil && pl.connected == true {
					pl.toConn <- fmt.Sprintf("%s left the lobby.", p.Name)
				}
			}
		case p := <-l.ready:
			util.Sugar.Infow("Ready",
				"Lobby", l.LobbyID,
				"Player", p.Name)
			// delete(l.player, p.Name)
			for _, pl := range l.player {
				if pl.conn != nil && pl.connected == true {
					pl.toConn <- fmt.Sprintf("%s is ready to Play.", p.Name)
				}
			}
			if l.areAllActivePlayersReady() {
				for _, pl := range l.player {
					if pl.conn != nil && pl.connected == true {
						pl.toConn <- fmt.Sprintf("Lobby will start in 5 Seconds!")
					}
				}
				timer = time.NewTimer(time.Second * 5)
			}
		case <-timer.C:
			// Start the Gameplay management goroutine.
			go l.serveGame()
			// Start the Lobby
			l.started = true
			// Send message to all connected clients
			for _, p := range l.player {
				if p.conn != nil && p.connected == true {
					p.toConn <- fmt.Sprintf("Consider yourself redirected.")
					// p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				}
			}
			// Stop all still running Goroutines.
			for _, p := range l.player {
				if p.conn != nil && p.connected == true {
					p.ctxcancel()
				}
			}
			// Reset Connected.
			for _, p := range l.player {
				if p.conn != nil && p.connected == true {
					p.connected = false
					// p.conn = nil
				}
			}
			// Stop this function.
			// return

			// Start the Lobby
			l.started = true
		}
	}
}

func (l *Lobby) serveGame() {}

func (l *Lobby) getPlayer(name string) (*Player, error) {
	if p, ok := l.player[name]; ok {
		return p, nil
	}
	return nil, errors.New(fmt.Sprintln(name, "not found for Lobby", l.LobbyID))
}

func (l *Lobby) getActivePlayers() string {
	liste := ""
	for _, p := range l.player {
		if p.connected != false {
			liste = liste + " " + p.Name
		}
	}
	return liste
}

func (l *Lobby) areAllActivePlayersReady() bool {
	for _, p := range l.player {
		if p.connected == false {
			continue
		}
		if p.ready == false {
			return false
		}
	}
	return true
}
