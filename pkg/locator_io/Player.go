package locator_io

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// Player is a middleman between the websocket connection and the hub.
type Player struct {
	lobby *Lobby

	// The Username as provided by the User
	User string

	// The websocket connection.
	conn *websocket.Conn

	toSend chan []byte
}

func (p *Player) sendNewLocation() error {
	fmt.Println("sendNewLocation")

	err := p.conn.WriteMessage(websocket.TextMessage, []byte("New Location"))
	if err != nil {
		p.cleanBrokenConn()
	}
	return err
}

func (p *Player) sendRoundReview() error {
	fmt.Println("sendRoundReview")
	w, err := p.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	w.Write([]byte("New RoundReview"))

	if err := w.Close(); err != nil {
		return err
	}
	return nil
}
func (p *Player) sendFinalReview() error {
	fmt.Println("sendFinalReview")
	w, err := p.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	w.Write([]byte("New Location"))

	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

func (p *Player) cleanBrokenConn() {
	fmt.Println(len(p.lobby.player))
	p.lobby.unregister <- p
	p.conn.Close()
	fmt.Println(len(p.lobby.player))
}

func (p *Player) checkOnConnection() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("checkFailed", err)
				p.cleanBrokenConn()
				return
			}
		}
	}
}
