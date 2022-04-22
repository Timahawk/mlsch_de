package locator_io

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Timahawk/mlsch_de/pkg/util"
)

type Waitingroom struct {
	// Registered clients.
	players map[string]*Player

	// Player names
	player_names []string

	// Register requests from the clients.
	register chan *Player

	// Unregister requests from clients.
	unregister chan *Player

	// Lobby
	lobby *Lobby
}

func (w *Waitingroom) run() {
	log.Println("Starting WaitingRoom")
	for {
		select {
		case newPlayer := <-w.register:

			// log.Println("Message Received.")
			w.players[newPlayer.User] = newPlayer
			w.player_names = append(w.player_names, newPlayer.User)

			util.Sugar.Debugw("New Player added",
				"WaitRoom", w.lobby.LobbyID,
				"player", newPlayer.User,
			)
			list, _ := json.Marshal(w.player_names)
			str := fmt.Sprintf(`{"status":"Waiting", "Player":%s}`, list)

			for _, p := range w.players {
				go p.WriteToConn(str)
			}

		case oldPlayer := <-w.unregister:
			log.Println("Player unregisterd", oldPlayer)
			delete(w.players, oldPlayer.User)
			for idx, name := range w.player_names {
				if name == oldPlayer.User {
					remove(w.player_names, idx)
				}
			}

			list, _ := json.Marshal(w.player_names)
			str := fmt.Sprintf(`{"status":"Waiting", "Player":%s}`, list)

			for _, p := range w.players {
				go p.WriteToConn(str)
			}
		}
	}
}

// remove deletes elemeent at i in array; Stolen from
// https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
// func remove[K any](s []K, i int) []K {
// 	s[i] = s[len(s)-1]
// 	return s[:len(s)-1]
// }
func remove[K any](s []K, i int) []K {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
