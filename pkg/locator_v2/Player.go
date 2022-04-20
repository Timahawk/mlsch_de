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
}

func (p *Player) WriteToConn(str string) {
	// This is stupid because it may be to short.
	err := p.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 50))
	if err != nil {
		util.Sugar.Debugw("WriteDeadline failed",
			"WaitRoom", p.lobby.LobbyID,
			"player", p.Name,
			"error", err,
		)
		p.lobby.remove <- p
	}
	err = p.conn.WriteMessage(websocket.TextMessage, []byte(str))
	if err != nil {
		util.Sugar.Debugw("WriteMessage failed",
			"WaitRoom", p.lobby.LobbyID,
			"player", p.Name,
			"error", err,
		)
		p.lobby.remove <- p
	}
}
