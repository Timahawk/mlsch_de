package locator_io

import (
	"context"
	"encoding/json"
	"fmt"
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

	// distance away to lobby.CurrentLocation
	distance int
	// Points awarded for distance
	point int
	// totalPoints awarded during game
	points int
}

func (p *Player) String() string {
	return fmt.Sprintf("LobbyID: %s; UserID %s", p.lobby.LobbyID, p.User)
}

// SendMessage runs indefinitly and sends everything the hub assigns it.
func (p *Player) SendMessages() {

	defer func() {
		util.Sugar.Debugw("SendMessages stopped",
			"player", p.User,
			"lobby", p.lobby.LobbyID)
	}()

	util.Sugar.Debugw("SendMessages started",
		"player", p.User,
		"lobby", p.lobby.LobbyID)
	for {
		select {
		case toSend := <-p.toSend:
			// This is to stop multiple writes to Dead Connection.
			// Not really working.
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			//if p.lobby.state == 1 {
			//toSend = []byte(strings.Replace(string(toSend), "XXX", strconv.Itoa(p.distance), 1))
			//p.lobby.points[p.User] = p.lobby.points[p.User] + p.point
			// fmt.Println(p.lobby.points)
			// }
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

func (p *Player) WriteToConn(str string) {
	// This is stupid because it may be to short.
	err := p.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 50))
	if err != nil {
		util.Sugar.Debugw("WriteDeadline failed",
			"WaitRoom", p.lobby.LobbyID,
			"player", p.User,
			"error", err,
		)
		p.lobby.waitRoom.unregister <- p
	}
	err = p.conn.WriteMessage(websocket.TextMessage, []byte(str))
	if err != nil {
		util.Sugar.Debugw("WriteMessage failed",
			"WaitRoom", p.lobby.LobbyID,
			"player", p.User,
			"error", err,
		)
		p.lobby.waitRoom.unregister <- p
	}
}

// ReceiveMessages runs, parses and handles incoming messages.
func (p *Player) ReceiveMessages() {
	defer func() {
		util.Sugar.Debugw("ReceiveMessages stopped",
			"player", p.User,
			"lobby", p.lobby.LobbyID)
		p.lobby.unregister <- p
		p.conn.Close()
		p.cancel()
	}()

	util.Sugar.Debugw("ReceiveMessages started",
		"player", p.User,
		"lobby", p.lobby.LobbyID)
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

		util.Sugar.Debugw("Message received",
			"player", p.User,
			"lobby", p.lobby.LobbyID,
			"message", string(message))

		dist, err := p.processSubmit(message)
		if err != nil {
			util.Sugar.Warnw("Process failed",
				"player", p.User,
				"lobby", p.lobby.LobbyID,
				"message", string(message),
				"distance", dist,
				"error", err)
		}

		p.distance = int(dist)

		p.point = p.awardPoints(dist)

		util.Sugar.Debugw("Points awarded",
			"player", p.User,
			"lobby", p.lobby.LobbyID,
			"message", string(message),
			"distance", dist,
			"points", p.point,
		)

		//log.Println("Submit", p.lobby.CurrentLocation, p.User, string(message), "Dist", dist, p.lobby.game.Cities[p.lobby.CurrentLocation].Lat, p.lobby.game.Cities[p.lobby.CurrentLocation].Lng)
		// Handle Message and Calcualte Points:
		// p.lobby.points[p.User] = p.lobby.points[p.User] + int(dist)

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

	util.Sugar.Debugw("processing Message",
		"player", p.User,
		"lobby", p.lobby.LobbyID,
		"message", string(message))

	err := json.Unmarshal(message, &submit)
	if err != nil {
		util.Sugar.Warnw("Marshalling failed",
			"player", p.User,
			"lobby", p.lobby.LobbyID,
			"message", string(message),
			"error", err)
		return 0, err
	}

	city, StatusOK := p.lobby.game.Cities[p.lobby.CurrentLocation]
	if !StatusOK {
		util.Sugar.Warnw("City does not exist",
			"player", p.User,
			"lobby", p.lobby.LobbyID,
			"message", string(message),
			"error", err)
		return 0, fmt.Errorf("ḱey does not exists")
	}

	// log.Println("Submit:", submit)
	distance := math.Round(
		util.Distance(
			submit.Latitude,
			submit.Longitude,
			city.Lat,
			city.Lng) / 1000)
	return distance, nil
}

func (p *Player) awardPoints(dist float64) int {
	switch {
	case dist < 10.0:
		return 7
	case dist < 50.0:
		return 5
	case dist < 100.0:
		return 4
	case dist < 250.0:
		return 3
	case dist < 500.0:
		return 2
	case dist < 1000.0:
		return 1
	default:
		return 0
	}
}
