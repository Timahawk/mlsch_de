package locator_v2

import (
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

	//If the Player is ready to start the game
	ready bool
}

func (p *Player) WriteToConn() {
	util.Sugar.Infow("WriteToConn started",
		"Lobby", p.lobby.LobbyID,
		"player", p.Name,
	)
	defer func() {
		util.Sugar.Infow("WriteToConn stopped",
			"Lobby", p.lobby.LobbyID,
			"player", p.Name,
		)
	}()
	for {
		select {
		case str := <-p.toConn:
			// This is stupid because it may be to short.
			err := p.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 50))
			if err != nil {
				util.Sugar.Debugw("WriteDeadline failed",
					"Lobby", p.lobby.LobbyID,
					"player", p.Name,
					"error", err,
				)
				p.lobby.drop <- p
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
		}

	}
}

func (p *Player) ReceiveFromConn() {
	util.Sugar.Infow("ReceiveFromConn started",
		"WaitRoom", p.lobby.LobbyID,
		"player", p.Name,
	)
	defer func() {
		util.Sugar.Infow("ReceiveFromConn stopped",
			"WaitRoom", p.lobby.LobbyID,
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
	// p.conn.SetCloseHandler(
	// 	func(code int, text string) error {
	// 		log.Println("SetCloseHandler Called.")

	// 		message := websocket.FormatCloseMessage(code, "")
	// 		// CloseMessage denotes a close control message. The optional message
	// 		// payload contains a numeric code and text. Use the FormatCloseMessage
	// 		// function to format a close message payload.
	// 		CloseMessage := 8
	// 		p.conn.WriteControl(CloseMessage, message, time.Now().Add(10*time.Second))
	// 		return fmt.Errorf("SetCloseHandler Called.")
	// 	})
	for {
		_, _, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				p.lobby.drop <- p
				util.Sugar.Infow("p.ReceiveFromConn",
					"Lobby", p.lobby.LobbyID,
					"Player", p.Name,
					"error", "IsUnexpectedCloseError")
				return
			}
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				p.lobby.drop <- p
				p.conn.Close()
				p.conn = nil
				util.Sugar.Infow("p.ReceiveFromConn",
					"Lobby", p.lobby.LobbyID,
					"Player", p.Name,
					"error", "IsCloseError")
				return
			}
		}
		//log.Println(string(message))
		p.ready = true
		p.lobby.ready <- p
	}
}
