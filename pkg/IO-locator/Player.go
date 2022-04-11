package games

import (
	"fmt"

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

func (p Player) sendNewLocation() error {
	fmt.Println("sendNewLocation")
	return nil
}
func (p Player) sendRoundReview() error {
	fmt.Println("sendRoundReview")
	return nil
}
func (p Player) sendFinalReview() error {
	fmt.Println("sendFinalReview")
	return nil
}
