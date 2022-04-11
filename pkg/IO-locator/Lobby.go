package games

import (
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/gin-gonic/gin"
)

const reviewTime = time.Second * 5

// Lobby maintains the set of active Player and broadcasts messages to the
// clients.
type Lobby struct {
	// Hub ID
	LobbyID string

	// this specifies when a new round comes
	time time.Duration

	// Registered clients.
	player map[string]*Player

	// Inbound messages from the clients.
	// This is called, when the jsclient sends a message
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Player

	// Unregister requests from clients.
	unregister chan *Player

	// Number of Rounds to Play.
	roundCounter int

	// This is the actual game that is played.
	game *Game
}

type Game struct {
	name    string
	center  []float64
	zoom    int
	maxZoom int
	minZoom int
	extent  []float64
	Cities  *[]City
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
	id := util.RandString(8)

	lobby := Lobby{id, time, nil, nil, nil, nil, 5, game}

	return &lobby
}

func (l *Lobby) run() {

	// TODO maybe switch to timers...
	// This determines the time a new Location is send.
	newLocation := time.NewTicker(l.time)

	// This determines the roundReview Time
	roundReview := time.NewTicker(reviewTime)
	roundReview.Stop()

	for {
		select {

		// Neue Location Zeit:
		case <-newLocation.C:
			// loop through all players and send them the new Location.
			for _, player := range l.player {
				err := player.sendNewLocation()
				if err != nil {
					// DoSomething
					break
				}
			}
			// finally start the roundReview ticker:
			roundReview = time.NewTicker(reviewTime)
			// and stop the newLocation Ticker:
			newLocation.Stop()

		// Runden Review Zeit.
		case <-roundReview.C:
			// loop through all players and send them the RoundReview.
			for _, player := range l.player {
				err := player.sendRoundReview()
				if err != nil {
					// DoSomething
					break
				}
			}
			// Sends different stuff when last round.
			if l.roundCounter == 0 {
				// loop through all players and send them the FinalReview.
				for _, player := range l.player {
					err := player.sendFinalReview()
					if err != nil {
						// DoSomething
						break
					}
				}
			}
			// loop through all players and send them Round Review.

			// Decrease roundCounter.
			l.roundCounter--

			// finally start the newLocation ticker:
			newLocation = time.NewTicker(l.time)
			// and stop the newLocation Ticker:
			roundReview.Stop()

		// Spieler wird hinzugefügt.
		case newPlayer := <-l.register:
			l.player[newPlayer.User] = newPlayer

		// Spieler wurde entfernt.
		case removePlayer := <-l.unregister:
			delete(l.player, removePlayer.User)

			// This checks if the Lobby is empty.
			if n := len(l.player); n == 0 {
				return
			}
		}
	}
}

func CreateLobby(c *gin.Context) {
	lobby := NewLobby(10*time.Second, &Game{})
	lobby.run()
	c.JSON(200, gin.H{"status": "Lobby started!"})
}
