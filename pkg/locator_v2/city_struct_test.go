package locator_v2

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// //go:embed cities/*.json
// var CitiesFS embed.FS

func Test_LoadCities(t *testing.T) {
	// files, _ := CitiesFS.ReadDir("cities")
	// for _, file := range files {
	// 	fmt.Println("File name: ", file.Name(), file.Type())
	// }
	//CitiesFS.ReadFile("cities/allci")

	_, err := LoadCities("cities/allcities.json")

	assert.Nil(t, err)
	//a, b := data.CitiesFS.ReadDir("cities/allcities.json")
	// fmt.Println(a, b)
	// assert.Equal(t, 1, 2)
}

func Test_komischeUmlaute(t *testing.T) {
	content, err := CitiesFS.ReadFile("cities/testfile.json")
	assert.Nil(t, err)

	cities := make([]City, 0)
	err = json.Unmarshal(content, &cities)
	if err != nil {
		assert.Nil(t, err)
	}

	cities_map := make(map[string]Locations)

	//  what the fuck is the difference between those two?
	// TODO figure this out.
	for i := 0; i < len(cities); i++ {
		cities_map[cities[i].Name] = &cities[i]
	}

	fmt.Println(cities[0].GetName())

	//assert.Equal(t, 1, 2)
}
