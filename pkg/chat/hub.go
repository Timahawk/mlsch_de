package chat

import (
	"log"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
)

const (
	// Period to loop through all Hubs and Close those without Clients.
	closeTime = 5 * time.Minute
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Hub ID
	HubID string

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// newHub creates and runs a new Hub.
func newHub() *Hub {
	hub := &Hub{
		HubID:      util.RandString(6), // TODO muss was bessers her.
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
	go hub.run()
	return hub
}

// run manages incoming messages and and distributes them to the client.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {

				// I dont get why this is nessesarry
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// CloseClientless checks at closeTime each hub, if it has any clients.
// If not the hub is closed.
func CloseClientlessHubs(closeTime time.Duration) {
	ticker := time.NewTicker(closeTime)
	defer func() {
		ticker.Stop()
	}()
	fail := make(chan bool)
	for {
		select {
		case <-ticker.C:
			for _, hub := range Hubs {
				if x := len(hub.clients); x == 0 {

					delete(Hubs, hub.HubID)
				}
			}
		// This is utterly stupid
		// But don´t know how to fix.
		case <-fail:
			log.Panic("This should NEVER be run")
		}
	}
}
