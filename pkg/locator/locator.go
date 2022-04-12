package locator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/gin-gonic/gin"
)

var Games []Game

type Game struct {
	name    string
	center  []float64
	zoom    int
	maxZoom int
	minZoom int
	extent  []float64
	Cities  []City
}

func NewGame(name, pfad string, center []float64, zoom, maxZoom, minZoom int, extent []float64) error {
	cities, err := LoadCities(pfad)
	if err != nil {
		return err
	}
	Games = append(Games, Game{name, center, zoom, maxZoom, minZoom, extent, cities})
	return nil
}

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

func LoadCities(file string) ([]City, error) {
	cities := make([]City, 0)

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s, %v ", file, err)
	}
	err = json.Unmarshal(content, &cities)
	if err != nil {
		return nil, fmt.Errorf("%s, %v ", file, err)
	}
	return cities, nil
}

func getGame(name string) (Game, error) {
	for _, game := range Games {
		if game.name == name {
			return game, nil
		}
	}
	return Game{}, fmt.Errorf("Game not found")
}

func HandleGame(c *gin.Context) {
	name := c.Param("country")

	game, err := getGame(name)
	if err != nil {
		c.String(404, "Game not found!")
		return
	}

	length := len(game.Cities)
	city := game.Cities[rand.Intn(length)]

	c.HTML(200,
		"locators/game.html",
		gin.H{
			"title":   fmt.Sprintf("Locator " + game.name),
			"city":    city.Name,
			"center":  game.center,
			"zoom":    game.zoom,
			"maxZoom": game.maxZoom,
			"minZoom": game.minZoom,
			"extent":  game.extent,
		})
}

type Submit_guess struct {
	City      string
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

func HandleGameSubmit(c *gin.Context) {
	name := c.Param("country")
	game, err := getGame(name)
	if err != nil {
		c.String(404, "Game not found!")
		return
	}

	var submit Submit_guess
	if err := c.ShouldBindJSON(&submit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var city_lat float64
	var city_lng float64
	for _, city := range game.Cities {
		if city.Name == submit.City {
			city_lat = city.Lat
			city_lng = city.Lng
			// capital_name = capi.Name
			break
		}
	}

	distance := util.Distance(
		city_lat,
		city_lng,
		submit.Latitude,
		submit.Longitude)

	c.JSON(200, gin.H{
		"city":     submit.City,
		"lat":      city_lat,
		"long":     city_lng,
		"distance": math.Round(distance / 1000)})
}

func HandleNewGuess(c *gin.Context) {
	name := c.Param("country")

	game, err := getGame(name)
	if err != nil {
		c.String(404, "Game not found!")
		return
	}

	length := len(game.Cities)
	capital := game.Cities[rand.Intn(length)]

	c.JSON(200, gin.H{"city": capital.Name})
}
