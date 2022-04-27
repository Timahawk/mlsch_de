package locator_v2

import (
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
