package locator_io

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second * 60

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Player is a middleman between the websocket connection and the hub.
type Player struct {
	ctx context.Context

	lobby *Lobby

	// The Username as provided by the User
	User string

	// The websocket connection.
	conn *websocket.Conn

	toSend chan []byte

	cancel context.CancelFunc
}

func (p *Player) String() string {
	return fmt.Sprintf("LobbyID: %s; UserID %s", p.lobby.LobbyID, p.User)
}

// SendMessage runs indefinitly and sends everything the hub assigns it.
func (p *Player) SendMessages() {

	defer func() {
		log.Println("Send Message stopped", p)
	}()

	for {
		select {
		case toSend := <-p.toSend:
			// This is to stop multiple writes to Dead Connection.
			// Not really working.
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := p.conn.WriteMessage(websocket.TextMessage, toSend)
			if err != nil {
				p.lobby.unregister <- p
				p.conn.Close()
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// ReceiveMessages runs, parses and handles incoming messages.
func (p *Player) ReceiveMessages() {
	defer func() {
		log.Println("Receive Messages for", p, "stopped.")
		p.lobby.unregister <- p
		p.conn.Close()
		p.cancel()
	}()

	log.Println("ReceiveMessage started!")
	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(
		func(string) error {
			p.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
	for {
		_, message, err := p.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// logger.Error("error:", err)
				return
			}
		}

		// This is my way to stop this goroutine, if the Connection is dead.
		if string(message) == "" {
			return
		}

		log.Printf("%s %v", p.User, string(message))

		dist, err := p.processSubmit(message)
		if err != nil {
			log.Println("Wrong message.")

		}
		// Handle Message and Calcualte Points:
		p.lobby.points[p.User] = p.lobby.points[p.User] + func(dist float64) int {
			return int(dist)
		}(dist)

		p.lobby.submitReceived <- submit{p.User, true}
	}
}

// Struct which is the incoming new Guess.
type Submit_guess struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

// processSubmit calculates the distance and stuff.
func (p *Player) processSubmit(message []byte) (float64, error) {
	var submit Submit_guess
	err := json.Unmarshal(message, &submit)
	if err != nil {
		return 0, err
	}

	city, StatusOK := p.lobby.game.Cities[p.lobby.CurrentLocation]
	if !StatusOK {
		log.Println("processFailed", StatusOK)
		return 0, fmt.Errorf("ḱey does not exists")
	}

	// log.Println("Submit:", submit)
	distance := math.Round(
		util.Distance(
			submit.Latitude,
			submit.Longitude,
			city.Lat,
			city.Lat) / 1000)
	return distance, nil
}
