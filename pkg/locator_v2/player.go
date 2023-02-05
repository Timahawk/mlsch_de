package locator_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/gorilla/websocket"
)

// Player is a middleman between the websocket connection and the hub.
type Player struct {
	lobby *Lobby

	// The Username as provided by the User
	Name string

	// The websocket connection.
	conn *websocket.Conn

	// connected
	connected bool

	// Channel which receives messages to be sent to the client.
	toConn chan string
	// Channel which receives messages from the client.
	fromConn chan string

	//If the Player is ready to start the game
	ready bool

	// Context to properly stop WriteToConn & ReceiveFromConn
	ctx context.Context
	// Cancelfunc
	ctxcancel context.CancelFunc

	//If the Player submitted his guess
	submitted bool

	// The distance of the last guess.
	distance float64
	// Points awarded for the last round.
	points int
	// Points awareded each round -> sum is total points.
	score []int
	// The location of your last guess.
	lastLat float64
	lastLng float64
}

func NewPlayer(ctx context.Context, ctxcancel context.CancelFunc, lobby *Lobby, name string) *Player {
	return &Player{
		lobby:     lobby,
		Name:      name,
		conn:      nil,
		connected: false,
		toConn:    make(chan string),
		fromConn:  make(chan string),
		ready:     false,
		ctx:       ctx,
		ctxcancel: ctxcancel,
		submitted: false}
}

func (p *Player) WriteToConn() {
	util.Sugar.Debugw("WriteToConn started",
		"Lobby", p.lobby.LobbyID,
		"player", p.Name,
	)
	defer func() {
		util.Sugar.Debugw("WriteToConn stopped",
			"Lobby", p.lobby.LobbyID,
			"player", p.Name,
		)
	}()
	for {
		select {
		case str := <-p.toConn:
			// This is stupid because it may be to short.
			err := p.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
			if err != nil {
				util.Sugar.Warnw("WriteDeadline failed",
					"Lobby", p.lobby.LobbyID,
					"player", p.Name,
					"error", err,
				)
				p.lobby.drop <- p
			}
			// TODO this is stupid i think.
			if strings.Contains(str, "points") {
				str = str + fmt.Sprintf(`,"distance":"%v", "awarded":"%v"}`, p.distance, p.score[len(p.score)-1])
			}

			err = p.conn.WriteMessage(websocket.TextMessage, []byte(str))
			if err != nil {
				util.Sugar.Debugw("WriteMessage failed",
					"WaitRoom", p.lobby.LobbyID,
					"player", p.Name,
					"error", err,
				)
				p.lobby.drop <- p
			}
		case <-p.ctx.Done():
			// p.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 50))
			// p.conn.WriteMessage(websocket.CloseMessage, []byte("Consider yourself Closed!"))
			return
		}
	}
}

// ReceiveFromConn does not terminate proberly when the connection is closed.
func (p *Player) ReceiveFromConn() {
	util.Sugar.Debugw("ReceiveFromConn started",
		"Lobby", p.lobby.LobbyID,
		"player", p.Name,
	)
	defer func() {
		p.lobby.drop <- p
		p.conn.Close()

		// Dont know but this is needed for some reaseon.
		if p.conn != nil {
			p.ctxcancel()
		}
		p.conn = nil
		p.connected = false
		p.ready = false

		util.Sugar.Debugw("ReceiveFromConn stopped",
			"Lobby", p.lobby.LobbyID,
			"player", p.Name,
		)
	}()
	// Maximum message size allowed from peer.
	p.conn.SetReadLimit(512)
	// Time allowed to read the next pong message from the peer.
	p.conn.SetReadDeadline(time.Now().Add(60 * time.Second * 60))
	p.conn.SetPongHandler(
		func(string) error {
			// Time allowed to read the next pong message from the peer.
			p.conn.SetReadDeadline(time.Now().Add(60 * time.Second * 60))
			return nil
		})

	for {
		// this should prevent all the invalid memory address or nil pointer dereference errors
		if p.conn == nil {
			util.Sugar.Debugw("Conn == nil in ReceiveFromClientLoop",
				"Lobby", p.lobby.LobbyID,
				"Player", p.Name)
			return
		}

		_, message, err := p.conn.ReadMessage()
		if err != nil {

			util.Sugar.Debugw("Read Message failed",
				"Lobby", p.lobby.LobbyID,
				"Player", p.Name,
				"error", err,
				"conn", p.conn)
			return
		}

		// log.Println(string(message), err)
		if string(message) == "ready" {
			p.ready = true
			p.lobby.ready <- p
			continue
		}

		var submit SubmitGuess
		err = json.Unmarshal(message, &submit)
		if err == nil {

			p.processSubmit(submit)
			p.points = p.getPoints()

			util.Sugar.Debugw("Points calculated and awarded",
				"Lobby", p.lobby.LobbyID,
				"Player", p.Name,
				"conn", p.conn,
				"location", p.lobby.location,
				"locations", p.lobby.locations,
				"distance", p.distance,
				// this breaks at the first entry.
				// "awarded", p.score[len(p.score)-1],
				"score", p.score)

			// this is doubled in p.process_submit.
			p.lobby.submitted <- p
			p.submitted = true
		} else {
			// Dont know what those are, but yeah fuck them.
			util.Sugar.Debugw("Marshalling failed",
				"error", err,
				"Message", message,
				"Player", p.Name,
				"Lobby", p.lobby.LobbyID)
		}
	}
}

type SubmitGuess struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

func (p *Player) processSubmit(submit SubmitGuess) error {

	loc, StatusOK := p.lobby.game.Cities[p.lobby.location]
	if !StatusOK {
		return fmt.Errorf("failed to get City")
	}
	p.lastLat = submit.Latitude
	p.lastLng = submit.Longitude
	// log.Println("Submit:", submit)
	p.distance = loc.Distance(submit.Latitude, submit.Longitude)
	return nil
}

func (p *Player) getPoints() int {
	dist := p.distance / 1000
	scorevalue := p.lobby.game.Scorevalue

	switch {
	case dist < 5:
		return 10
	case scorevalue*0.025 > dist:
		return 7
	case scorevalue*0.05 > dist:
		return 6
	case scorevalue*0.1 > dist:
		return 5
	case scorevalue*0.2 > dist:
		return 4
	case scorevalue*0.3 > dist:
		return 3
	case scorevalue*0.4 > dist:
		return 2
	case scorevalue*0.5 > dist:
		return 1
	default:
		return 0
	}
}

func (p *Player) calcScore() int {
	score := 0
	for _, point := range p.score {
		score += point
	}
	return score
}