package assets

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestSameNumberOffFiles(t *testing.T) {
	files, err := filepath.Glob("cities/*.json")
	assert.Nil(t, err)
	inFs, err := Cities.ReadDir("cities")
	assert.Nil(t, err)

	assert.Equal(t, len(files), len(inFs))
}

func TestSameFileContent(t *testing.T) {
	german, err := os.ReadFile("cities/german_cities.json")
	assert.Nil(t, err)

	germanFS, err := Cities.ReadFile("cities/german_cities.json")
	assert.Nil(t, err)

	assert.Equal(t, german, germanFS)
}
