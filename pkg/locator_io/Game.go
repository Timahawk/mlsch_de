package locator_io

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// The City as per the file.
type City struct {
	// json_featuretype string
	Name       string `json:"city"`
	Name_ascii string `json:"city_ascii"`
	Lat        float64
	Lng        float64
	Country    string
	Iso2       string
	Iso3       string
	// admin_name       string
	Capital    string
	Population int
	Id         int
}

// Game is the actual game that is played within the Lobby.
type Game struct {
	// CurrentLocation string
	Name    string
	Center  []float64
	Zoom    int
	MaxZoom int
	MinZoom int
	Extent  []float64
	Cities  map[string]*City
}

func (g *Game) String() string {
	return fmt.Sprintf(" with %v Locations \n", len(g.Cities))
}

func NewGame(name, pfad string, center []float64, zoom, maxZoom, minZoom int, extent []float64) (*Game, error) {
	cities, err := LoadCities(pfad)
	if err != nil {
		return &Game{}, err
	}
	newGame := Game{name, center, zoom, maxZoom, minZoom, extent, cities}
	// Games[name] = &newGame
	return &newGame, nil
}

func LoadCities(file string) (map[string]*City, error) {
	cities := make([]City, 0)
	cities_map := make(map[string]*City)
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s, %v ", file, err)
	}
	err = json.Unmarshal(content, &cities)
	if err != nil {
		return nil, fmt.Errorf("%s, %v ", file, err)
	}

	for _, city := range cities {
		cities_map[city.Name_ascii] = &city
	}

	return cities_map, nil
}
