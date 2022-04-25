package locator_v2

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"golang.org/x/exp/slices"
)

// Lobby maintains the set of active Player and broadcasts messages to the
// clients. It is the dreh & angelpunkt of the ganze Veranstaltung.
type Lobby struct {
	// Hub ID
	LobbyID string

	// Who creates the lobby. and starts the Lobby.
	// owner *Player

	// This determines the weather the lobby is started.
	started bool

	// Determines the duration of a guessing round.
	RoundTime int
	// Review Time determines the time spent reviewing
	ReviewTime int
	// The number of rounds played
	Rounds int

	// Registered clients.
	player map[string]*Player

	// Register requests from the clients.
	add chan *Player

	// Unregister requests from clients.
	drop chan *Player

	// Receives calls that players are ready
	ready chan *Player

	// Receives calls that players submitted.
	submitted chan *Player

	// The game which is played.
	game *Game

	// This determines if review or Location
	// reviewing
	// guessing
	// finished
	// startup
	state     string
	nextState string

	// The location used by the game
	location string
	// All played locations
	locations []string
}

// NewLobby creates a new Lobby.
func NewLobby(roundt, reviewt, rounds int, game *Game) *Lobby {
	id := util.RandString(8)

	lobby := Lobby{
		LobbyID: id,
		// owner:      owner,
		started:    false,
		RoundTime:  roundt,
		ReviewTime: reviewt,
		Rounds:     rounds,
		player:     make(map[string]*Player),
		add:        make(chan *Player, 10),
		drop:       make(chan *Player, 100),
		ready:      make(chan *Player, 10),
		submitted:  make(chan *Player, 10),
		game:       game,
		state:      "startup",
		nextState:  "guessing",
		location:   "",
		locations:  []string{},
	}

	// Lobbies[id] = &lobby

	util.Sugar.Infow("New Lobby created.",
		"id", id,
		"roundtime", roundt,
		"reviewtime", reviewt,
		"rounds", rounds,
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
			// l.player[p.Name] = p
			// p.connected = true

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
			// This should have always been handled on the Player side.
			// p.connected = false
			// p.ready = false
			// p.ctxcancel()
			// p.conn.Close()

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
						pl.toConn <- fmt.Sprintf("Lobby will start in 2 Seconds!")
					}
				}
				timer = time.NewTimer(time.Second * 2)
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
					// Should stop WriteToConn
					// p.ctxcancel()
					// Should stop ReadfromConn
					// p.conn.Close()

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
			return
		}
	}
}

func (l *Lobby) serveGame() {
	defer func() {
		l.started = false
		delete(Lobbies, l.LobbyID)
		util.Sugar.Infow("serveGame stopped and Lobby deleteted",
			"Lobby", l.LobbyID)
	}()
	util.Sugar.Infow("serveGame started",
		"Lobby", l.LobbyID)

	sendUpdate := time.NewTimer(2 * time.Second)

	for {
		select {

		// A Player submitted his guess.
		case p := <-l.submitted:
			util.Sugar.Debugw("A Player submitted sth.",
				"lobby", l.LobbyID,
				"player", p.Name,
				"state", l.state,
				"nextState", l.nextState,
			)

			// We are Currently in a review Round -> so do nothing.
			if l.state == "reviewing" || l.state == "startup" || l.state == "finished" {
				break
			}

			if l.allActivePlayersSubmitted() {

				// Set to Location mode for next time the Timer ticks down.
				l.state = "guessing"
				l.nextState = "reviewing"

				util.Sugar.Infow("All Players submitted.",
					"lobby", l.LobbyID,
					"time", l.RoundTime,
					"state", l.state,
					"nextState", l.nextState,
				)

				sendUpdate = time.NewTimer(0)
			}

		// Something needs to be send!
		case <-sendUpdate.C:

			// Case 1 New Location
			if l.nextState == "guessing" {

				l.state = "guessing"
				l.nextState = "reviewing"

				util.Sugar.Infow("guessing",
					"lobby", l.LobbyID,
					"location", l.location,
					"time", l.RoundTime,
					"state", l.state,
					"nextState", l.nextState,
				)

				// log.Println("Sending new Location")

				l.location = l.getNewLocation()

				str := fmt.Sprintf(`{"status":"location","Location":"%s", "state": "%v", "time":"%v", "rounds":"%v"}`, l.location, l.state, l.RoundTime, l.Rounds)

				for _, p := range l.player {
					if p.conn != nil && p.connected == true {
						p.toConn <- str
						p.submitted = false
						p.distance = 9999

					}
				}

				sendUpdate = time.NewTimer(time.Duration(l.RoundTime) * time.Second)

				// Case 2 Review
			} else if l.nextState == "reviewing" {

				l.state = "reviewing"
				l.nextState = "guessing"
				l.Rounds -= 1
				if l.Rounds == 0 {
					l.nextState = "finished"
				}

				util.Sugar.Infow("reviewing",
					"lobby", l.LobbyID,
					"location", l.location,
					"time", l.ReviewTime,
					"state", l.state,
					"nextState", l.nextState,
				)

				for _, p := range l.player {
					if p.conn != nil && p.connected == true {
						p.score = append(p.score, p.points)
						p.points = 0
					}
				}

				str := fmt.Sprintf(`{"status":"reviewing","Location":"%s", "state": "%v", "time":"%v", "lat":"%v", "lng":"%v", "points":%s`,
					l.location, l.state, l.ReviewTime, l.game.Cities[l.location].Lat, l.game.Cities[l.location].Lng, string(l.getScore()))
				for _, p := range l.player {
					if p.conn != nil && p.connected == true {
						p.toConn <- str
					}
				}

				sendUpdate = time.NewTimer(time.Duration(l.ReviewTime) * time.Second)
			} else if l.nextState == "finished" {
				util.Sugar.Infow("finished",
					"lobby", l.LobbyID,
					"location", l.location,
					"time", l.ReviewTime,
					"state", l.state,
					"nextState", l.nextState,
					"rounds", l.Rounds)

				str := fmt.Sprintf(`{"status":"finished","points":%s`, string(l.getScore()))
				for _, p := range l.player {
					if p.conn != nil && p.connected == true {
						p.toConn <- str
					}
				}
				time.Sleep(time.Second)
				return
			} else {
				util.Sugar.Warnw("Timer run down. Nothing happend...",
					"lobby", l.LobbyID,
					// "time", l.RoundTime,
					"state", l.state,
					"nextState", l.nextState,
				)
				return
			}

			// This simply checks if the serveLobby goroutine should be exited.
			i := 0
			for _, p := range l.player {
				if p.conn == nil {
					i += 1
				}
				if i == len(l.player) {
					util.Sugar.Infow("serveLobby will be stopped because all players are disconnected.",
						"Lobby", l.LobbyID)
					return
				}
			}
		}
	}
}

func (l *Lobby) getPlayer(name string) (*Player, error) {
	if p, ok := l.player[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("%s not found for Lobby %s", name, l.LobbyID)
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

// getNewLocation helper function, gets a random new Locations
// and appends it to l.locations
func (l *Lobby) getNewLocation() string {
	for key := range l.game.Cities {
		if slices.Contains(l.locations, key) {
			fmt.Println(l.locations, key)
			continue
		}
		l.locations = append(l.locations, key)
		return key
	}
	return ""
}

func (l *Lobby) allActivePlayersSubmitted() bool {
	for _, p := range l.player {
		if p.connected == false {
			continue
		}
		if p.submitted == false {
			return false
		}
	}
	return true
}

func (l *Lobby) getScore() []byte {
	liste := make(map[string]int)
	for _, p := range l.player {
		liste[p.Name] = p.calcScore()

	}
	res, err := json.Marshal(liste)
	if err != nil {
		log.Println(liste, err)
	}
	return res
}
