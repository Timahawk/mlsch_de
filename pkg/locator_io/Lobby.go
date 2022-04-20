package locator_io

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
)

// ReviewTime determines the duration of the review Screen.
const ReviewTime = time.Second * 7

// Lobby maintains the set of active Player and broadcasts messages to the
// clients. It is the dreh & angelpunkt of the ganze Veranstaltung.
type Lobby struct {
	// Hub ID
	LobbyID string

	// Who creates the lobby. and starts the Lobby.
	// TODO
	//owner *Player

	//WaitingRoom
	waitRoom *Waitingroom

	// This determines the wether the lobby is started.
	started bool

	// Determines the duration of a guessing round.
	RoundTime int

	// Registered clients.
	player map[string]*Player

	// Points for each Player
	points map[string]int

	// Monitors if a player submitted a Guess
	checkSubmits map[string]bool

	// Inbound messages from the clients.
	submitReceived chan submit

	// Register requests from the clients.
	register chan *Player

	// Unregister requests from clients.
	unregister chan *Player

	// Number of Rounds to Play.
	roundCounter int

	// This is the actual game that is played.
	game *Game

	// This determines if review or Location
	// 0 = Currently Reviewing
	// 1 = Currently Guessing
	// 2 = finished
	// 3 = Startup Phase
	state     int
	nextState int

	CurrentLocation string
}

func (l *Lobby) String() string {
	return fmt.Sprintf("LobbyID: %s, Game: %v", l.LobbyID, l.game)
}

// Submit is the response struct send by the Client.
type submit struct {
	playerID  string
	submitted bool
}

// NewLobby creates a new Lobby.
func NewLobby(zeit int, game *Game) *Lobby {
	id := util.RandString(8)
	//id := "A"

	lobby := Lobby{
		id,
		&Waitingroom{
			register:     make(chan *Player),
			unregister:   make(chan *Player),
			players:      make(map[string]*Player),
			player_names: make([]string, 0),
			lobby:        &Lobby{}},
		false,
		zeit,
		make(map[string]*Player),
		make(map[string]int),
		make(map[string]bool),
		make(chan submit),
		make(chan *Player),
		make(chan *Player),
		0,
		game,
		3,
		0,
		"Wait for a sec."}

	lobby.waitRoom.lobby = &lobby

	// go lobby.run()
	Lobbies[id] = &lobby
	util.Sugar.Debugw("New Lobby created.",
		"id", id,
		"time", zeit,
		"state", 3,
		"nextState", 0,
		"roundCounter", 0,
		"game", game.Name,
	)
	return &lobby
}

// run organized the complete Game Magic.
// So far there are four major cases. Each notified via a chanel.
//
// 1) Add of a new Player
// 2) Remove of a Player. If no Players left, close Lobby.
// 3) All Player submitted a guess -> Start the review cycle
// 4) The timer is zero. Two possibilites:
// 	- Start a guess cycle.
// 	- Start a review cycle
// 	-> Reset counter.
func (l *Lobby) run() {
	util.Sugar.Debugw("Lobby running",
		"lobby", l.LobbyID,
	)
	// time.Sleep(time.Second * 5)
	// l.CurrentLocation = l.getNewLocation()
	sendUpdate := time.NewTimer(5 * time.Second)
	// l.state = 0

	for {
		select {

		// Spieler wird hinzugefügt.
		case newPlayer := <-l.register:

			str := l.addPlayer(newPlayer)
			newPlayer.toSend <- str

		// Spieler wurde entfernt.
		case removePlayer := <-l.unregister:

			l.removePlayer(removePlayer)

			// This checks if the Lobby is empty.
			if n := len(l.player); n == 0 {
				delete(Lobbies, l.LobbyID)
				util.Sugar.Debugw("Lobby closed.",
					"lobby", l.LobbyID,
				)
				return
			}

		// A Player submitted his guess.
		case newSubmit := <-l.submitReceived:

			util.Sugar.Debugw("Guess submitted",
				"lobby", l.LobbyID,
				"player", newSubmit.playerID,
				"submit", newSubmit.submitted,
			)

			// We are Currently in a review Round
			if l.nextState == 0 || l.nextState == 3 {
				// log.Println("Game State was ", l.state)
				break
			}

			l.checkSubmits[newSubmit.playerID] = newSubmit.submitted
			if len(l.checkSubmits) == len(l.player) {
				// log.Println("All Players submitted; Starting next round.")

				util.Sugar.Debugw("Sending new Round",
					"lobby", l.LobbyID,
				)

				// Send them Round Review.
				sendUpdate = time.NewTimer(ReviewTime)
				// Set to Location mode for next time the Timer ticks down.
				l.state = 1
				l.nextState = 0
				// Reset the array.
				l.checkSubmits = make(map[string]bool)

				util.Sugar.Debugw("Sending new Round",
					"lobby", l.LobbyID,
					"time", l.RoundTime,
					"state", l.state,
					"nextState", l.nextState,
				)

				// Sends an update
				l.sendPointsToClient()

				// Decrease Counter
				// l.roundCounter++
			}
		// Something needs to be send!
		case <-sendUpdate.C:

			// Case 1 New Location
			if l.nextState == 0 || l.nextState == 3 {
				// log.Println("Sending new Location")

				l.CurrentLocation = l.getNewLocation()
				str := fmt.Sprintf(`{"status":"location","Location":"%s", "state": "%v", "time":"%v"}`, l.CurrentLocation, l.state, l.RoundTime)

				for _, player := range l.player {
					player.toSend <- []byte(str)
					player.distance = 9999
					player.point = 0
				}

				sendUpdate = time.NewTimer(time.Duration(l.RoundTime) * time.Second)
				l.roundCounter++
				// Switch to review Mode.
				l.state = 0
				l.nextState = 1

				util.Sugar.Debugw("New Location",
					"lobby", l.LobbyID,
					"location", l.CurrentLocation,
					"time", l.RoundTime,
					"state", l.state,
					"nextState", l.nextState,
				)
				break
			}

			// Case 2 Review
			if l.nextState == 1 {
				// log.Println("Sending Round review")

				sendUpdate = time.NewTimer(ReviewTime)

				l.sendPointsToClient()
				l.state = 1
				l.nextState = 0

				util.Sugar.Debugw("Sending new Round",
					"lobby", l.LobbyID,
					"time", l.RoundTime,
					"state", l.state,
					"nextState", l.nextState,
				)

			}

		}
	}
}

// removePlayer, removes the Player from the game.
func (l *Lobby) removePlayer(p *Player) error {
	util.Sugar.Debugw("Player removed",
		"lobby", l.LobbyID,
		"player", p.User,
		"LobbysizeOld", len(l.player),
		"LobbysizeNew", len(l.player)-1,
	)

	delete(l.player, p.User)
	return nil
}

// addPlayer adds the Player to the Playerbase.
// Also starts the Send and Receive Goroutines for this Player.
// Returns a message to send to the new Player.
func (l *Lobby) addPlayer(newPlayer *Player) []byte {
	l.player[newPlayer.User] = newPlayer

	util.Sugar.Debugw("New Player added",
		"lobby", l.LobbyID,
		"player", newPlayer.User,
	)
	go newPlayer.SendMessages()
	go newPlayer.ReceiveMessages()
	str := fmt.Sprintf(`{"status":"Waiting", "Player":"newPlayer.User"}`)
	return []byte(str)
}

// sendPointsToClient is a little helper function to send Message to the Player struct.
func (l *Lobby) sendPointsToClient() {

	for _, p := range l.player {
		// toSend = []byte(strings.Replace(string(toSend), "XXX", strconv.Itoa(p.distance), 1))
		p.lobby.points[p.User] = p.lobby.points[p.User] + p.point
		points, _ := json.Marshal(l.points)

		str := fmt.Sprintf(
			`{"status":"review", "points":%s, "state": "%v", "Location":"%s", "Round":"%v", "time":"%v", "distance":"%v", "lat":"%v","lng":"%v", "awarded":"%v"}`,
			points,
			l.state,
			l.CurrentLocation,
			l.roundCounter,
			ReviewTime.Seconds(),
			p.distance,
			l.game.Cities[l.CurrentLocation].Lat,
			l.game.Cities[l.CurrentLocation].Lng,
			p.point)

		p.toSend <- []byte(str)
	}
	util.Sugar.Debugw("SendingPointsToClient",
		"lobby", l.LobbyID,
		"Location", l.CurrentLocation,
		"points", l.points,
	)
}

// getNewLocation helper function, gets a semi random new Location.
func (l *Lobby) getNewLocation() string {

	for key := range l.game.Cities {
		return key
	}
	return ""
}
