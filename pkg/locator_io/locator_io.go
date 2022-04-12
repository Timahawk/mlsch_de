package locator_io

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

var Lobbies = map[string]*Lobby{}

func CreateLobby(c *gin.Context) {
	lobby := NewLobby(10*time.Second, &Game{})
	go lobby.run()
	c.JSON(200, gin.H{"status": "Lobby started!"})
}

func JoinLobby(c *gin.Context) {
	lobbyID := c.Param("lobby")

	_, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(200, gin.H{"status": "Lobby not found!"})
		return
	}
	c.HTML(200, "locator_io/game.html", gin.H{})
}

func ServeLobby(c *gin.Context) {
	lobbyID := c.Param("lobby")

	lobby, err := getLobby(lobbyID)
	if err != nil {
		c.JSON(200, gin.H{"status": "Lobby not found!"})
		return
	}
	user := util.RandString(7)
	createWS(lobby, user, c.Writer, c.Request)
}

func createWS(lobby *Lobby, user string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalln(err)
		return
	}

	ctx, _ := context.WithCancel(context.Background())

	player := Player{ctx, lobby, user, conn, make(chan []byte)}
	// log.Println("New Player registered", player)
	player.lobby.register <- &player
}

func getLobby(room string) (*Lobby, error) {

	if lobby, ok := Lobbies[room]; ok {
		return lobby, nil
	}
	return nil, errors.New(fmt.Sprintln("room not found for Room", room))
}
