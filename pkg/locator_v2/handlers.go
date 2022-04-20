package locator_v2

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateOrJoinLobby(c *gin.Context) {
	c.HTML(200, "locator_v2/CreateOrJoinLobby.html", gin.H{})
}

// CreateLobbyPost checks form, creates User, creates Lobby, adds User as Owner,
// starts the Lobbies sendWaitRoom goroutine, sends user to add channel,
// Adds the Lobby to Global Lobbies.
func CreateLobbyPOST(c *gin.Context) {
	roundTime, err := strconv.Atoi(c.PostForm("roundTime"))
	if err != nil {
		roundTime = 0
	}
	gameset := c.PostForm("gameset")
	username := c.PostForm("username")

	if roundTime == 0 || gameset == "" || username == "" {
		c.JSON(200, gin.H{"status": "CreateLobbyPost failed, due to faulty Form Input."})
		return
	}

	p := Player{&Lobby{}, username, nil}
	l := NewLobby(roundTime, nil, &p)

	go l.serveWaitRoom()

	l.add <- &p

	Lobbies[l.LobbyID] = l

	// c.JSON(200, gin.H{"status": "CreateLobbyPost", "Lobby": l})
	path := c.Request.URL.Path
	path = strings.Replace(path, "/create", "", 1)
	c.Redirect(303, fmt.Sprintf("%s/%s?user=%s", path, l.LobbyID, username))
}

// JoinLobbyPost checks form, creates User, sends user to add channel.
func JoinLobbyPOST(c *gin.Context) {
	lobbyID := c.PostForm("LobbyID")
	username := c.PostForm("username")

	if lobbyID == "" || username == "" {
		c.JSON(200, gin.H{"status": "JoinLobbyPost failed, due to faulty Form Input."})
		return
	}

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(200, gin.H{"status": "JoinLobbyPost failed, due to Lobby not Exists."})
		return
	}

	p := Player{&Lobby{}, username, nil}

	l.add <- &p

	// c.JSON(200, gin.H{"status": "JoinLobbyPost", "Lobby": l})
	path := c.Request.URL.Path
	path = strings.Replace(path, "/join", "", 1)
	c.Redirect(303, fmt.Sprintf("%s/%s?user=%s", path, l.LobbyID, username))
}

func WaitingRoom(c *gin.Context) {
	lobbyID := c.Param("lobby")

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(200, gin.H{"status": "WaitingRoom failed, due to Lobby not Exists."})
		return
	}

	username := c.Query("user")
	if username == "" {
		c.JSON(200, gin.H{"status": "WaitingRoom failed, due to faulty Parameter User"})
		return
	}

	user, err := l.getUser(username)
	if err != nil {
		c.JSON(200, gin.H{"status": "WaitingRoomWS failed, due to faulty Parameter User"})
		return
	}

	c.HTML(200, "locator_v2/Waitingroom.html", gin.H{"title": lobbyID, "user": user.Name})
}

func WaitingRoomWS(c *gin.Context) {
	lobbyID := c.Param("lobby")

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(200, gin.H{"status": "WaitingRoomWS failed, due to Lobby not Exists."})
		return
	}

	username := c.Query("user")
	if username == "" {
		c.JSON(200, gin.H{"status": "WaitingRoomWS failed, due to faulty Parameter User"})
		return
	}

	user, err := l.getUser(username)
	if err != nil {
		c.JSON(200, gin.H{"status": "WaitingRoomWS failed, due to faulty Parameter User"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error", err)
		return
	}

	user.conn = conn
}

func PlayGame(c *gin.Context) {

}
