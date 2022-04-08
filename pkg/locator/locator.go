package locator

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Capital struct {
	CountryName   string
	Name          string  `json:"CapitalName"`
	Lat           float64 `json:"CapitalLatitude,string"`
	Lng           float64 `json:"CapitalLongitude,string"`
	CountryCode   string
	ContinentName string
}

type Worldcity struct {
	// json_featuretype string
	Name_w  string `json:"city"`
	Name    string `json:"city_ascii"`
	Lat     float64
	Lng     float64
	Country string
	Iso2    string
	Iso3    string
	// admin_name       string
	Capital    string
	Population int
	Id         int
}

var Capitals []Capital
var Worldcities []Worldcity

// Hier wird der Datensatz ausgewählt.
//  Muss dann auch in "main" geändert werden.
var Gameset []Worldcity

// func main() {
// 	r := setupRouter()
// 	loadCapitals("./data/country-capitals.json")
// 	loadCities("./data/worldcities.json")

// 	Gameset = Worldcities
// 	rand.Seed(time.Now().UnixNano())
// 	r.Run("localhost:8080")
// }

// func setupRouter() *gin.Engine {
// 	router := gin.Default()

// 	router.LoadHTMLGlob("public/*")

// 	router.GET("/", HandleGame)
// 	router.POST("/submit", HandleGameSubmit)
// 	router.POST("/newGuess", HandleNewGuess)

// 	return router
// }

func LoadCapitals(file string) {
	// read file
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	err = json.Unmarshal(content, &Capitals)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
}
func LoadCities(file string) {
	// read file
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	err = json.Unmarshal(content, &Worldcities)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
}

func HandleGame(c *gin.Context) {
	length := len(Gameset)
	capital := Gameset[rand.Intn(length)]

	c.HTML(200,
		"locators/game.html",
		gin.H{
			"city": capital.Name})
}

type Submit_guess struct {
	City      string
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"long"`
}

func HandleGameSubmit(c *gin.Context) {
	var submit Submit_guess
	if err := c.ShouldBindJSON(&submit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// var capital_name string
	var capital_lat float64
	var capital_lng float64
	for _, capi := range Gameset {
		if capi.Name == submit.City {
			capital_lat = capi.Lat
			capital_lng = capi.Lng
			// capital_name = capi.Name
			break
		}
	}

	distance := Distance(
		capital_lat,
		capital_lng,
		submit.Latitude,
		submit.Longitude)

	c.JSON(200, gin.H{
		"city":     submit.City,
		"lat":      capital_lat,
		"long":     capital_lng,
		"distance": math.Round(distance / 1000)})
}

func HandleNewGuess(c *gin.Context) {
	length := len(Gameset)
	capital := Gameset[rand.Intn(length)]

	c.JSON(200, gin.H{"city": capital.Name})
}
