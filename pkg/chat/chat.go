package chat

import (
	"errors"
	"fmt"

	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

var Hubs = map[string]*Hub{}

func init() {
	// Set a random Seed.
	rand.Seed(time.Now().UnixNano())
}

// r.POST("/")
func PostCreateNewHub(c *gin.Context) {
	hub := newHub()
	Hubs[hub.HubID] = hub
	c.Redirect(303, fmt.Sprintf("/%s/chat", hub.HubID))
}

// r.GET(":room")
func GetChatRoom(c *gin.Context) {
	room := c.Param("room")
	_, err := getHub(room)
	if err != nil {
		c.String(404, "Room not found")
		return
	}
	c.HTML(200, "chat.html", nil)
}

// r.GET(":room/ws"
func GetRoomWebsocket(c *gin.Context) {
	room := c.Param("room")
	user := c.Query("user")

	hub, err := getHub(room)
	if err != nil {
		// logger.Info("RoomID not available In Websocket", err)
		c.String(404, "Room not found")
		return // errors.New(fmt.Sprintln("Room not found", err))
	}
	// Handles the Websocket, for this particular requests.
	serveWs(hub, user, c.Writer, c.Request)

	// Goroutine that checks if OpenHubs are connected to,
	// if not Hub is deleted.
	// TODO check if all depending goroutines are stopped/closed
	go CloseClientlessHubs(closeTime)
	// return nil
}

// Simple Helper function to check if Hub exists.
func getHub(room string) (*Hub, error) {

	if hub, ok := Hubs[room]; ok {
		return hub, nil
	}
	return &Hub{}, errors.New(fmt.Sprintln("room not found for Room", room))
}
