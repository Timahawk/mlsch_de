package locator_v2

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Timahawk/mlsch_de/pkg/util"
)

// All currently loaded games
var LoadedGames = map[string]*Game{}

// Using an interface now instead of Cities.
// This should allow me to easily introduce Other types of Geometrys
// like a Polygons by using Postgres.
type Locations interface {
	Distance(lat, lng float64) float64
	Geom() string
	Center() [2]float64
	GetName() string
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
	Geom    string
	Cities  map[string]Locations
	// Scorevalue represents sth like the biggest difference possible between all Locations
	// Therefor as points are awarded they stay relativ to this max distance.
	Scorevalue float64
}

func LoadGames() error {
	// Bounding Boxes are created using https://boundingbox.klokantech.com/
	err := errors.New("")
	LoadedGames["world"], err = NewGame("world", "assets/cities/allcities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90}, "Point", 10000)
	if err != nil {
		return fmt.Errorf("Loading Games faile, %v ", err)
	}
	LoadedGames["cities_larger_250000"], err = NewGame("cities_larger_250000", "assets/cities/cities_larger_250000.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90}, "Point", 10000)
	if err != nil {
		return fmt.Errorf("Loading Games faile, %v ", err)
	}
	LoadedGames["capitals"], err = NewGame("capitals", "assets/cities/capital_cities.json", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90}, "Point", 10000)
	if err != nil {
		return fmt.Errorf("Loading Games faile, %v ", err)
	}
	// Germany
	LoadedGames["germany"], err = NewGame("germany", "assets/cities/german_cities.json", []float64{10.019531, 50.792047}, 1, 14, 1, []float64{-2.55, 42.18, 22.58, 58.86}, "Point", 750)
	if err != nil {
		return fmt.Errorf("Loading Games faile, %v ", err)
	}
	LoadedGames["germany_larger25000"], err = NewGame("germany_larger25000", "assets/cities/german_cities_larger25000.json", []float64{10.019531, 50.792047}, 1, 14, 1, []float64{-2.55, 42.18, 22.58, 58.86}, "Point", 750)
	if err != nil {
		return fmt.Errorf("Loading Games faile, %v ", err)
	}
	// Japan
	LoadedGames["japan_larger25000"], err = NewGame("japan_larger25000", "assets/cities/japan_cities_larger25000.json", []float64{138.3, 34.76}, 1, 14, 1, []float64{118.44, 20.8, 155.53, 52.0}, "Point", 1250)
	if err != nil {
		return fmt.Errorf("Loading Games faile, %v ", err)
	}
	// Region specific games
	LoadedGames["european_cities_larger_100000"], err = NewGame("european_cities_larger_100000", "assets/cities/european_cities_larger_100000.json", []float64{10.019531, 50.792047}, 1, 14, 1, []float64{-41.8, 27.0, 69.6, 73.7}, "Point", 2000)
	if err != nil {
		return fmt.Errorf("Loading Games faile, %v ", err)
	}
	LoadedGames["north_american_cities_larger_100000"], err = NewGame("north_american_cities_larger_100000", "assets/cities/north_american_cities_larger_100000.json", []float64{-100, 40}, 1, 14, 1, []float64{-180, -15, 40, 85}, "Point", 2000)
	if err != nil {
		return fmt.Errorf("Loading Games faile, %v ", err)
	}
	// ************************** Polygon Games ******************************* //

	LoadedGames["country"], err = NewGame("country", "pg/lvl_0/country", []float64{0, 0}, 1, 14, 1, []float64{180.0, -90, -180, 90}, "Polygon", 10000)
	if err != nil {
		return fmt.Errorf("Loading Games faile, %v ", err)
	}
	return nil
}

func NewGame(name, pfad string, center []float64, zoom, maxZoom, minZoom int, extent []float64, geom string, Scorevalue float64) (*Game, error) {
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
			"Geom", geom,
		)
	}()

	locs := make(map[string]Locations)
	err := errors.New("")

	switch {
	case strings.HasPrefix(pfad, "assets/cities"):
		locs, err = LoadCities(pfad)
		if err != nil {
			util.Sugar.Fatal(pfad, err)
		}
	case strings.HasPrefix(pfad, "pg/lvl_0/country"):
		locs, err = NewWorldBorders()
		if err != nil {
			util.Sugar.Fatal(pfad, err)
		}
	default:
		util.Sugar.Fatalf("%s, pfad could not be laoded", pfad)
	}

	newGame := Game{name, center, zoom, maxZoom, minZoom, extent, geom, locs, Scorevalue}

	return &newGame, nil
}

func getGame(name string) (*Game, error) {
	if g, ok := LoadedGames[name]; ok {
		return g, nil
	}
	return nil, errors.New(fmt.Sprintln(name, "is not a available Game."))
}
