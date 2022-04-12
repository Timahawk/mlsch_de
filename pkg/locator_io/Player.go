package locator_io

import (
	"context"
	"fmt"
	"log"
	"time"

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
}

func (p *Player) String() string {
	return fmt.Sprintf("LobbyID: %s; UserID %s", p.lobby.LobbyID, p.User)
}

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

func (p *Player) ReceiveMessages() {
	defer func() {
		log.Println("Receive Messages for", p, "stopped.")
		p.lobby.unregister <- p
		p.conn.Close()
		p.ctx.Done()
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

		log.Printf("%s %s", p.User, message)

		// Handle Message and Calcualte Points:
		p.lobby.points[p] = func() int {
			return 17
		}()

		p.lobby.submitReceived <- submit{p.User, true}
	}
}
