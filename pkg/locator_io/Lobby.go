package locator_io

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
)

// ReviewTime determines the duration of the review Screen.
const ReviewTime = time.Second * 10

// Lobby maintains the set of active Player and broadcasts messages to the
// clients. It is the dreh & angelpunkt of the ganze Veranstaltung.
type Lobby struct {
	// Hub ID
	LobbyID string

	// Determines the duration of a guessing round.
	RoundTime time.Duration

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
	state int

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
// Todo: der Lobbyname muss noch angepasst werden.
func NewLobby(time time.Duration, game *Game) *Lobby {
	id := util.RandString(8)
	//id := "A"

	lobby := Lobby{
		id,
		time,
		make(map[string]*Player),
		make(map[string]int),
		make(map[string]bool),
		make(chan submit),
		make(chan *Player),
		make(chan *Player),
		0,
		game,
		0,
		""}

	go lobby.run()
	Lobbies[id] = &lobby
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
	log.Println("Lobby ", l, "started")
	time.Sleep(time.Second * 5)
	l.CurrentLocation = l.getNewLocation()
	sendUpdate := time.NewTimer(time.Second * l.RoundTime)
	l.state = 0

	for {
		select {

		// Spieler wird hinzugefügt.
		case newPlayer := <-l.register:

			l.player[newPlayer.User] = newPlayer
			log.Println("New Player added.", newPlayer)
			go newPlayer.SendMessages()
			go newPlayer.ReceiveMessages()

			x := fmt.Sprintf(`{"status":"location","Location":"%s", "state": "%v"}`, l.CurrentLocation, l.state)
			newPlayer.toSend <- []byte(x)

		// Spieler wurde entfernt.
		case removePlayer := <-l.unregister:
			log.Println("Player to be removed.", removePlayer)

			delete(l.player, removePlayer.User)
			fmt.Println("Still ", len(l.player), "in the Lobby")

			// This checks if the Lobby is empty.
			if n := len(l.player); n == 0 {
				log.Println("Lobby is empty and will be closed.")
				delete(Lobbies, l.LobbyID)
				return
			}

		// A Player submitted his guess.
		case newSubmit := <-l.submitReceived:

			log.Println("Submit Received", newSubmit)

			// We are Currently in a review Round
			if l.state == 0 {
				log.Println("Game State was 1")
				break
			}

			l.checkSubmits[newSubmit.playerID] = newSubmit.submitted
			if len(l.checkSubmits) == len(l.player) {
				log.Println("All Players submitted; Starting next round.")

				// Send them Round Review.
				sendUpdate = time.NewTimer(ReviewTime)
				// Set to Location mode for next time the Timer ticks down.
				l.state = 0
				// Reset the array.
				l.checkSubmits = make(map[string]bool)

				// Sends an update
				l.sendPointsToClient()

				// Decrease Counter
				l.roundCounter--
			} else {
				log.Println("Not enough Player submitted!", len(l.checkSubmits), "of", len(l.player))
			}

		// Something needs to be send!
		case <-sendUpdate.C:

			// Case 1 New Location
			if l.state == 0 {
				// log.Println("Sending new Location")

				l.CurrentLocation = l.getNewLocation()
				str := fmt.Sprintf(`{"status":"location","Location":"%s", "state": "%v" }`, l.CurrentLocation, l.state)

				for _, player := range l.player {
					player.toSend <- []byte(str)
				}

				sendUpdate = time.NewTimer(l.RoundTime)
				l.roundCounter--
				// Switch to review Mode.
				l.state = 1

				break
			}

			// Case 2 Review
			if l.state == 1 {
				// log.Println("Sending Round review")

				sendUpdate = time.NewTimer(ReviewTime)

				l.sendPointsToClient()
				l.state = 0
			}

		}
	}
}

// sendPointsToClient is a little helper function to send Message to the Player struct.
func (l *Lobby) sendPointsToClient() {
	points, _ := json.Marshal(l.points)

	str := fmt.Sprintf(`{"status":"review", "points":%s, "state": "%v"}`, points, l.state)

	for _, player := range l.player {
		player.toSend <- []byte(str)
	}
}

// getNewLocation helper function, gets a semi random new Location.
func (l *Lobby) getNewLocation() string {

	for key := range l.game.Cities {
		return key
	}
	return ""
}
