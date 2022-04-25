package locator_v2

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

func LoadCities(file string) (map[string]Locations, error) {
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

	cities_map := make(map[string]Locations)

	//  what the fuck is the difference between those two?
	// TODO figure this out.
	for i := 0; i < len(cities); i++ {
		cities_map[cities[i].Name] = &cities[i]
	}

	// This assigns each value the same pointer!
	//for _, city := range cities {
	//	cities_map[city.Name_ascii] = &city
	//}

	return cities_map, nil
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
