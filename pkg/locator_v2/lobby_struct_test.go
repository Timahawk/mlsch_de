package locator_v2

import (
	"fmt"
	"testing"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/stretchr/testify/assert"
)

func init() {
	util.InitLogger()
}

func Test_getNewLocation(t *testing.T) {

	l := NewLobby(3, 3, 3, &Game{
		Cities: make(map[string]Locations),
	})

	assert.Equal(t, "", l.getNewLocation(), "No entry available")
	l.game.Cities["test"] = &City{}
	assert.Equal(t, "test", l.getNewLocation(), "Only entry already played")
	l.game.Cities["nottest"] = &City{}
	assert.Equal(t, "nottest", l.getNewLocation(), "They should be equal")
}

func Test_getPlayer(t *testing.T) {
	l := NewLobby(3, 3, 3, &Game{
		Cities: make(map[string]Locations),
	})
	p, err := l.getPlayer("TEST")
	assert.Nil(t, p)
	assert.Equal(t, err, fmt.Errorf("%s not found for Lobby %s", "TEST", l.LobbyID))

	testplayer := &Player{}
	l.player["TEST"] = testplayer
	p, err = l.getPlayer("TEST")
	assert.Nil(t, err)
	assert.Equal(t, p, testplayer)
}

func Test_getLastLocations(t *testing.T) {
	l := &Lobby{player: make(map[string]*Player)}

	l.player["A"] = &Player{Name: "A", last_lat: 10, last_lng: 10}
	l.player["B"] = &Player{Name: "B", last_lat: 20, last_lng: 20}
	l.player["C"] = &Player{Name: "C", last_lat: 0, last_lng: 0}

	assert.Equal(t, `{"A":[10,10],"B":[20,20],"C":[0,0]}`, string(l.getLastLocations()))
}
