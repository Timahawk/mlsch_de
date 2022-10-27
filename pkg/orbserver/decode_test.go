package orbserver

import (
	"testing"

	"github.com/Timahawk/mlsch_de/assets"
)

func TestLoadEmbeddedFC(t *testing.T) {
	LoadEmbeddedFC(assets.Mvt)

	t.Fail()
}
