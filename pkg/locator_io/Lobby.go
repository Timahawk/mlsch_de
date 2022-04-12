package locator_io

import (
	"fmt"
	"log"
	"time"
)

const ReviewTime = time.Second * 5

// Lobby maintains the set of active Player and broadcasts messages to the
// clients.
type Lobby struct {
	// Hub ID
	LobbyID string

	// this specifies when a new round comes
	RoundTime time.Duration

	// Registered clients.
	player map[string]*Player

	// Points for each Player
	points map[*Player]int

	// Submitted
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
}

func (l *Lobby) String() string {
	return fmt.Sprintf("LobbyID: %s, Game: %v", l.LobbyID, l.game)
}

type submit struct {
	playerID  string
	submitted bool
}

type Game struct {
	/*
		name    string
		center  []float64
		zoom    int
		maxZoom int
		minZoom int
		extent  []float64
		Cities  *[]City
	*/
}

type City struct {
	// json_featuretype string
	Name       string `json:"city"`
	Name_ascii string `json:"city_ascii"`
	Lat        float64
	Lng        float64
	Country    string
	Iso2       string
	Iso3       string
	// admin_name       string
	Capital    string
	Population int
	Id         int
}

func NewLobby(time time.Duration, game *Game) *Lobby {
	//id := util.RandString(8)
	id := "A"

	lobby := Lobby{
		id,
		time,
		make(map[string]*Player),
		make(map[*Player]int),
		make(map[string]bool),
		make(chan submit),
		make(chan *Player),
		make(chan *Player),
		5,
		game}
	go lobby.run()

	return &lobby
}

func (l *Lobby) run() {
	log.Println("Lobby ", l, "started")
	sendUpdate := time.NewTimer(time.Second * 10)

	for {
		select {

		// Spieler wird hinzugefügt.
		case newPlayer := <-l.register:

			l.player[newPlayer.User] = newPlayer
			log.Println("New Player added.", newPlayer)
			go newPlayer.SendMessages()
			go newPlayer.ReceiveMessages()

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

			// TODO: There should be a check if receveid in Game or in Review Mode!

			l.checkSubmits[newSubmit.playerID] = newSubmit.submitted
			if len(l.checkSubmits) == len(l.player) {
				log.Println("All Players submitted; Starting next round.")

				// Send them Round Review.
				sendUpdate = time.NewTimer(ReviewTime)

				// Reset the array.
				l.checkSubmits = make(map[string]bool)
			}

		// Something needs to be send!
		case <-sendUpdate.C:

			// Case 1 New Location
			if l.roundCounter%2 == 0 {
				// log.Println("Sending new Location")
				for _, player := range l.player {
					player.toSend <- []byte("New Location")
				}
				sendUpdate = time.NewTimer(l.RoundTime)
			}

			// Case 2 Review
			if l.roundCounter%2 != 0 {
				// log.Println("Sending Round review")
				for _, player := range l.player {
					player.toSend <- []byte("Round Review")
				}
				sendUpdate = time.NewTimer(ReviewTime)
			}
			l.roundCounter--
		}
	}
}
