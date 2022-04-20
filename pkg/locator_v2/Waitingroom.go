package locator_v2

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
