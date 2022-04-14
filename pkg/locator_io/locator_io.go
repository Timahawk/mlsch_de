package locator_io

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

// All currently active Lobbies
var Lobbies = map[string]*Lobby{}
var LoadedGames = map[string]*Game{}

// CreateLobby sends the Template for the Creation of new Lobbies.
func CreateLobbyGET(c *gin.Context) {
	// go lobby.run()
	//c.JSON(200, gin.H{"status": "Lobby started!"})
	c.HTML(200, "locator_io/createLobby.html", gin.H{})
}

func CreateLobbyPOST(c *gin.Context) {

	ti := c.PostForm("roundTime")
	g := c.PostForm("gameset")

	if ti == "" || g == "" {
		c.JSON(400, gin.H{"status": "Form not good"})
		log.Println(ti, g)
		return
	}

	timeINT, _ := strconv.Atoi(ti)

	game, err := getGame(g)
	if err != nil {
		fmt.Println(LoadedGames)
		c.JSON(200, gin.H{"status": "CREATE GAME -> Game not found."})
		return
	}

	lobby := NewLobby(timeINT, game)
	log.Println("Created new Lobby", lobby.LobbyID, lobby.game.Name, lobby.RoundTime)

	Lobbies[lobby.LobbyID] = lobby

	c.Redirect(303, fmt.Sprintf("%s/%s", c.Request.URL.Path, lobby.LobbyID))
}

// Join Lobby is the function which sends the actual gamepage.
func JoinLobby(c *gin.Context) {
	lobbyID := c.Param("lobby")

	_, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(200, gin.H{"status": "Lobby not found!"})
		return
	}
	c.HTML(200, "locator_io/game.html", gin.H{})
}

// ServeLobby creates the Websocket connection.
// Also creates the new Player struct, and adds it to the game.
func ServeLobby(c *gin.Context) {
	lobbyID := c.Param("lobby")

	lobby, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(200, gin.H{"status": "Lobby not found!"})
		return
	}

	user := util.RandString(7)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatalln(err)
		return
	}

	// I dont know if this is the best way to handle the Context,
	// but so far it works.
	ctx, fn := context.WithCancel(context.Background())

	player := Player{ctx, lobby, user, conn, make(chan []byte), fn}

	// log.Println("New Player registered", player)
	player.lobby.register <- &player
}

// getLobby helper function to get the Lobby, if exists.
func getLobby(room string) (*Lobby, error) {

	if lobby, ok := Lobbies[room]; ok {
		return lobby, nil
	}
	return nil, errors.New(fmt.Sprintln("room not found for Room", room))
}

// getLobby helper function to get the Lobby, if exists.
func getGame(room string) (*Game, error) {

	if lobby, ok := LoadedGames[room]; ok {
		return lobby, nil
	}
	return nil, errors.New(fmt.Sprintln("room not found for Room", room))
}
