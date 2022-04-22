package locator_io

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
)

// The City as per the file.
type City struct {
	// json_featuretype string
	Name       string  `json:"city"`
	Name_ascii string  `json:"city_ascii"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
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

// func (g *Game) String() string {
// 	return fmt.Sprintf(" with %v Locations", len(g.Cities))
// }

func NewGame(name, pfad string, center []float64, zoom, maxZoom, minZoom int, extent []float64) (*Game, error) {
	start := time.Now()
	defer func() {
		util.Sugar.Debugw("New Game created",
			"duration", time.Since(start),
			"name", name,
			"pfad", pfad,
			"center", center,
			"zoom", zoom,
			"maxZoom", maxZoom,
			"minZoom", minZoom,
			"extent", extent,
		)
	}()

	cities, err := LoadCities(pfad)
	// log.Println(cities)
	if err != nil {
		return &Game{}, err
	}
	newGame := Game{name, center, zoom, maxZoom, minZoom, extent, cities}
	// Games[name] = &newGame

	return &newGame, nil
}

func LoadCities(file string) (map[string]*City, error) {
	start := time.Now()

	defer func() {
		util.Sugar.Debugw("File/Cities loaded",
			"file", file,
			"duration", time.Since(start),
		)
	}()

	cities := make([]City, 0)

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s, %v ", file, err)
	}
	err = json.Unmarshal(content, &cities)
	if err != nil {
		return nil, fmt.Errorf("%s, %v ", file, err)
	}

	// log.Println(cities)

	cities_map := make(map[string]*City)

	//  what the fuck is the difference between those two?
	// TODO figure this out.
	for i := 0; i < len(cities); i++ {
		cities_map[cities[i].Name_ascii] = &cities[i]
	}

	// This assigns each value the same pointer!
	//for _, city := range cities {
	//	cities_map[city.Name_ascii] = &city
	//}

	return cities_map, nil
}
