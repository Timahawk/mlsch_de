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
	c.HTML(200, "locator_v2/CreateOrJoinLobby.html", gin.H{"Games": LoadedGames})
}

// CreateLobbyPOST checks form, creates User, creates Lobby, adds User as Owner,
// starts the Lobbies sendWaitRoom goroutine, sends user to add channel,
// Adds the Lobby to Global Lobbies.
func CreateLobbyPOST(c *gin.Context) {
	// fmt.Println("CreateLobby called")
	roundTime, err := strconv.Atoi(c.PostForm("roundTime"))
	if err != nil {
		roundTime = 0
	}
	reviewTime, err := strconv.Atoi(c.PostForm("reviewTime"))
	if err != nil {
		reviewTime = 0
	}
	rounds, err := strconv.Atoi(c.PostForm("rounds"))
	if err != nil {
		rounds = 0
	}
	gameset := c.PostForm("gameset")
	username := c.PostForm("username")

	if roundTime == 0 || reviewTime == 0 || rounds == 0 || gameset == "" || username == "" {
		// log.Println("Inputs:", roundTime, reviewTime, rounds, gameset, username)
		c.JSON(200, gin.H{"status": "CreateLobbyPost failed, due to faulty Form Input."})
		return
	}

	g, err := getGame(gameset)
	if err != nil {
		c.JSON(200, gin.H{"status": "CreateLobbyPost failed, due to Game not available."})
		return
	}

	ctx, cancelCtx := context.WithCancel(contextbg)
	p := NewPlayer(ctx, cancelCtx, &Lobby{}, username)

	l := NewLobby(roundTime, reviewTime, rounds, g)
	Lobbies[l.LobbyID] = l

	p.lobby = l

	go l.serveWaitRoom()

	l.player[p.Name] = p

	path := c.Request.URL.Path
	path = strings.Replace(path, "/create", "", 1)
	c.Redirect(303, fmt.Sprintf("%s/%s?user=%s", path, l.LobbyID, username))
}

// JoinLobbyPOST checks form, creates User, sends user to add channel.
func JoinLobbyPOST(c *gin.Context) {
	lobbyID := c.PostForm("LobbyID")
	username := c.PostForm("username")

	if lobbyID == "" || username == "" {
		c.JSON(213, gin.H{"status": "JoinLobbyPost failed, due to faulty Form Input.", "your input": lobbyID})
		return
	}

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(213, gin.H{"status": "JoinLobbyPost failed, due to Lobby not Exists.", "your input": lobbyID})
		return
	}

	if l.started == true {
		c.JSON(213, gin.H{"status": "You cannot join already started lobby."})
		return
	}

	if _, err := l.getPlayer(username); err == nil {
		c.JSON(213, gin.H{"status": "Your username was already taken."})
		return
	}

	ctx, cancelCtx := context.WithCancel(contextbg)
	p := NewPlayer(ctx, cancelCtx, l, username)

	l.player[p.Name] = p

	path := c.Request.URL.Path
	path = strings.Replace(path, "/join", "", 1)
	c.Redirect(303, fmt.Sprintf("%s/%s?user=%s", path, l.LobbyID, username))
}

func WaitingRoom(c *gin.Context) {

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

	if l.started {
		path := c.Request.URL.Path
		c.Redirect(303, fmt.Sprintf("%s/game?user=%s", path, username))
		return
	}

	ctx, cancelCtx := context.WithCancel(contextbg)
	p.ctx = ctx
	p.ctxcancel = cancelCtx
	l.add <- p

	c.HTML(200, "locator_v2/WaitingRoom.html", gin.H{
		"lobby": lobbyID,
		"title": lobbyID,
		"user":  p.Name,
		"game": fmt.Sprintln(
			"Game:", l.game.Name,
			"Rounds:", l.Rounds,
			"Time to Guess:", l.RoundTime,
			"Time to Review:", l.ReviewTime),
	})
}

func WaitingRoomWS(c *gin.Context) {

	lobbyID := c.Param("lobby")

	l, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(213, gin.H{"status": "WaitingRoomWS failed, due to Lobby not Exists."})
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
		util.Sugar.Warnw("Upgrade to Websocket failed",
			"Error", err,
			"Player", username,
			"Lobby", lobbyID)
		c.JSON(213, gin.H{"status": "WaitingRoomWS failed, due to Websocket Error"})
		return
	}

	p.conn = conn
	p.connected = true

	go p.WriteToConn()
	go p.ReceiveFromConn()

	util.Sugar.Debugw("WaitingRoomWs enabled",
		"Lobby", p.lobby.LobbyID,
		"player", p.Name)

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
		c.JSON(213, gin.H{"status": "PlayGame failed, due to Lobby not started or already finished"})
		return
	}

	username := c.Query("user")
	p, err := l.getPlayer(username)
	if err != nil {
		c.JSON(213, gin.H{"status": "PlayGame failed, due to faulty Parameter User"})
		return
	}

	util.Sugar.Debugw("Game Room send",
		"Lobby", p.lobby.LobbyID,
		"player", p.Name)

	c.HTML(200, "locator_v2/GameRoom.html", gin.H{
		"lobby":   lobbyID,
		"title":   lobbyID,
		"user":    p.Name,
		"center":  l.game.Center,
		"zoom":    l.game.Zoom,
		"maxZoom": l.game.MaxZoom,
		"minZoom": l.game.MinZoom,
		"extent":  l.game.Extent})
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
		util.Sugar.Warnw("Upgrade to Websocket failed",
			"Error", err,
			"Player", username,
			"Lobby", lobbyID)
		c.JSON(213, gin.H{"status": "GameRoomWS failed, due to Websocket Error"})
		return
	}

	p.conn = conn
	p.connected = true
	p.ctx, p.ctxcancel = context.WithCancel(contextbg)

	go p.WriteToConn()
	go p.ReceiveFromConn()

	util.Sugar.Debugw("GameRoomWs enabled",
		"Lobby", p.lobby.LobbyID,
		"player", p.Name)

	// This is for when reconnection during match.
	if l.state == "guessing" {
		str := fmt.Sprintf(`{"status":"location","Location":"%s", "state": "%v"}`, l.location, l.state)
		p.toConn <- str
	}
	if l.state == "reviewing" {
		str := fmt.Sprintf(`{"status":"location","Location":"%s", "state": "%v"}`, l.location, l.state)
		p.toConn <- str
	}
}
