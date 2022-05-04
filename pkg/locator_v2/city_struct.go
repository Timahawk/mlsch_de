package locator_v2

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Timahawk/mlsch_de/assets"
	"github.com/Timahawk/mlsch_de/pkg/util"
)

// The City as per the file.
type City struct {
	Name       string  `json:"city"`
	Name_ascii string  `json:"city_ascii"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	Country    string  `json:"country"`
	Iso2       string  `json:"iso2"`
	Iso3       string  `json:"iso3"`
	AdminName  string  `json:"admin_name"`
	Capital    string  `json:"capital"`
	Population int     `json:"population"`
	Id         int     `json:"id"`
}

func LoadCities(file string) (map[string]Locations, error) {
	start := time.Now()

	defer func() {
		util.Sugar.Debugw("File/Cities loaded",
			"file", file,
			"duration", time.Since(start),
		)
	}()

	cities := make([]City, 0)

	file = strings.Replace(file, "assets/", "", 1)
	content, err := assets.Cities.ReadFile(file)
	// content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s, %v ", file, err)
	}
	err = json.Unmarshal(content, &cities)
	if err != nil {
		return nil, fmt.Errorf("%s, %v ", file, err)
	}

	citiesMap := make(map[string]Locations)

	//  what the fuck is the difference between those two?
	// TODO figure this out.
	for i := 0; i < len(cities); i++ {
		citiesMap[cities[i].Name] = &cities[i]
	}

	// This assigns each value the same pointer!
	//for _, city := range cities {
	//	cities_map[city.Name_ascii] = &city
	//}

	if len(citiesMap) == 0 {
		return nil, fmt.Errorf("%s Load Cites loaded 0 Locations", file)
	}

	return citiesMap, nil
}

func (c *City) Distance(lat, lng float64) float64 {
	return util.CalcDistance(c.Lat, c.Lng, lat, lng)
}

func (c *City) Geom() string {
	// https://jsonlint.com/
	return fmt.Sprintf(`
	{
	"type": "FeatureCollection",
	"crs": 
		{
		"type": "name",
		"properties": 
			{
			"name": "EPSG:4326"
			}
		},
	"features": 
		[
			{
			"type": "Feature",
			"geometry": 
				{
				"type": "Point",
				"coordinates": [%v, %v]
				}
			}
		]
	}`,
		c.Lng, c.Lat)
}

// Center returns the coords in array Lat, Lng
func (c *City) Center() [2]float64 {
	center := [2]float64{}
	center[0], center[1] = c.Lat, c.Lng
	return center
}

func (c *City) GetName() string {
	return c.Name
}
