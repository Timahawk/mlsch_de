package locator_v2

import (
	"errors"
	"fmt"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
)

// Using an interface now instead of Cities.
// This should allow me to easily introduce Other types of Geometrys
// like a Polygons by using Postgres.
type Locations interface {
	Distance(lat, lng float64) float64
	Geom() string
	Center() [2]float64
	// Current() *Locations
	// Next() *Locations
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
	Cities  map[string]Locations
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

func getGame(name string) (*Game, error) {
	if g, ok := LoadedGames[name]; ok {
		return g, nil
	}
	return nil, errors.New(fmt.Sprintln(name, "is not a available Game."))
}
