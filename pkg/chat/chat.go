package chat

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

var Hubs = map[string]*Hub{}

// This is used for the template Names
var names = []string{}
var names_len int

func init() {
	// Set a random Seed.
	rand.Seed(time.Now().UnixNano())
	content, err := os.ReadFile("./pkg/chat/names.txt")
	if err != nil {
		log.Fatal(err)
	}
	names = strings.Split((string(content)), "\n")
	names_len = len(names)
}

// PostCreateNewHub is used to create and start a new Hub.
// The user will be redirected into this room.
// r.POST("/")
func PostCreateNewHub(c *gin.Context) {
	hub := newHub()
	Hubs[hub.HubID] = hub
	c.Redirect(303, fmt.Sprintf("%s/%s/chat", c.Request.URL.Path, hub.HubID))
}

// GetChatRoom is used to get into a chat room.
// It also establishes the Websocket connection.
// r.GET(":room")
func GetChatRoom(c *gin.Context) {
	room := c.Param("room")
	_, err := getHub(room)
	if err != nil {
		c.String(404, "Room not found")
		return
	}
	name := rand.Intn(names_len)
	c.HTML(200, "chats/chat.html", gin.H{"name": names[name]})
}

// GetRoomWebsocket handles the Websocket connections.
// It creates 2 Goroutines that handle writes and receives for the connection,
// and communicate with the Hub.
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

// getHub gets the Hub.
func getHub(room string) (*Hub, error) {

	if hub, ok := Hubs[room]; ok {
		return hub, nil
	}
	return &Hub{}, errors.New(fmt.Sprintln("room not found for Room", room))
}
