package locator_v2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/gin-gonic/gin"
)

func CreateOrJoinLobby(c *gin.Context) {
	c.HTML(200, "locator_v2/CreateOrJoinLobby.html", gin.H{})
}

// CreateLobbyPost checks form, creates User, creates Lobby, adds User as Owner,
// starts the Lobbies sendWaitRoom goroutine, sends user to add channel,
// Adds the Lobby to Global Lobbies.
func CreateLobbyPOST(c *gin.Context) {
	// fmt.Println("CreateLobby called")
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
	ctx, cancelCtx := context.WithCancel(contextbg)
	p := NewPlayer(ctx, cancelCtx, &Lobby{}, username)

	l := NewLobby(roundTime, nil, p)
	p.lobby = l

	go l.serveWaitRoom()

	l.add <- p

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
		c.JSON(213, gin.H{"status": "JoinLobbyPost failed, due to faulty Form Input."})
		return
	}

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(213, gin.H{"status": "JoinLobbyPost failed, due to Lobby not Exists."})
		return
	}
	ctx, cancelCtx := context.WithCancel(contextbg)
	p := NewPlayer(ctx, cancelCtx, l, username)
	l.add <- p
	// c.JSON(200, gin.H{"status": "JoinLobbyPost", "Lobby": l})
	path := c.Request.URL.Path
	path = strings.Replace(path, "/join", "", 1)
	c.Redirect(303, fmt.Sprintf("%s/%s?user=%s", path, l.LobbyID, username))
}

func WaitingRoom(c *gin.Context) {
	// fmt.Println("WaitingRoom called")
	lobbyID := c.Param("lobby")

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(213, gin.H{"status": "WaitingRoom failed, due to Lobby not Exists."})
		return
	}

	username := c.Query("user")
	p, err := l.getPlayer(username)
	if err != nil {
		c.JSON(213, gin.H{"status": "WaitingRoom failed, due to faulty Parameter User", "user": username, "error": err, "player": p})
		return
	}
	ctx, cancelCtx := context.WithCancel(contextbg)
	p.ctx = ctx
	p.ctxcancel = cancelCtx
	l.add <- p

	c.HTML(200, "locator_v2/WaitingRoom.html", gin.H{"title": lobbyID, "user": p.Name})
}

func WaitingRoomWS(c *gin.Context) {
	// fmt.Println("WaitingRoomWS called")
	lobbyID := c.Param("lobby")

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(213, gin.H{"status": "WaitingRoomWS failed, due to Lobby not Exists."})
		return
	}

	username := c.Query("user")
	// if username == "" {
	// 	c.JSON(200, gin.H{"status": "WaitingRoomWS failed, due to faulty Parameter User"})
	// 	return
	// }
	p, err := l.getPlayer(username)
	if err != nil {
		c.JSON(213, gin.H{"status": "WaitingRoomWS failed, due to faulty Parameter User"})
		return
	}

	// fmt.Println("This is before")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	// fmt.Println("This is after")
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	p.conn = conn

	go p.WriteToConn()
	go p.ReceiveFromConn()
	// util.Sugar.Infow("WaitingRoomWs enabled",
	// 	"Lobby", p.lobby.LobbyID,
	// 	"player", p.Name)

	p.toConn <- fmt.Sprintf("Already joined Players: %s", l.getActivePlayers())
}

func GameRoom(c *gin.Context) {
	lobbyID := c.Param("lobby")

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(213, gin.H{"status": "PlayGame failed, due to Lobby not Exists."})
		return
	}
	if l.started == false {
		c.JSON(213, gin.H{"status": "PlayGame failed, due to Lobby Ready"})
		return
	}

	username := c.Query("user")
	p, err := l.getPlayer(username)
	if err != nil {
		c.JSON(213, gin.H{"status": "PlayGame failed, due to faulty Parameter User"})
		return
	}
	c.HTML(200, "locator_v2/GameRoom.html", gin.H{"title": lobbyID, "user": p.Name})
}

func GameRoomWS(c *gin.Context) {
	lobbyID := c.Param("lobby")

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(213, gin.H{"status": "WaitingRoomWS failed, due to Lobby not Exists."})
		return
	}
	if l.started == false {
		c.JSON(213, gin.H{"status": "PlayGameWS failed, due to Lobby Ready"})
		return
	}

	username := c.Query("user")
	p, err := l.getPlayer(username)
	if err != nil {
		c.JSON(213, gin.H{"status": "WaitingRoomWS failed, due to faulty Parameter User"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	p.conn = conn

	go p.WriteToConn()
	go p.ReceiveFromConn()
	util.Sugar.Infow("GameRoomWs enabled",
		"Lobby", p.lobby.LobbyID,
		"player", p.Name)

	p.toConn <- fmt.Sprintf("Active Players: %s", l.getActivePlayers())
}
