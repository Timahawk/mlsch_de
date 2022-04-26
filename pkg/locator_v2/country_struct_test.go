package locator_v2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Distance(t *testing.T) {
	germany := Country{
		name_0: "Germany",
	}

	// Point in Germany
	assert.Equal(t, 0.0, germany.Distance(50, 10))
	// Point not in Germany
	assert.Equal(t, 6089768.964387792, germany.Distance(0, 0))
}

func Test_Geom(t *testing.T) {
	marino := Country{
		name_0: "San Marino",
	}

	assert.Equal(t,
		`{"type" : "FeatureCollection", "features" : [{"type": "Feature", "geometry": {"type":"MultiPolygon","coordinates":[[[[12.510080338,43.956047058],[12.513968468,43.948554992],[12.513464927,43.943706513],[12.5062809,43.937854768],[12.501674652,43.928112029],[12.495362281,43.925945283],[12.491223335,43.921176912],[12.491756,43.914852],[12.493819,43.91267],[12.487819672,43.905078888],[12.48621273,43.899791717],[12.465455055,43.898387909],[12.461083,43.895160999],[12.452511813,43.894943239],[12.447931291,43.89679718],[12.446157455,43.899635315],[12.446612999,43.902004],[12.440995016,43.905402996],[12.438079999,43.904835],[12.435917001,43.902054],[12.429788999,43.900818],[12.425123026,43.902027002],[12.417428971,43.898269653],[12.412594795,43.897838593],[12.405896186,43.899745942],[12.41050911,43.906032562],[12.405804634,43.918910981],[12.411922456,43.925662995],[12.412653923,43.929325104],[12.409431457,43.934036255],[12.401365281,43.938602448],[12.400666236,43.948101044],[12.40213108,43.950473785],[12.41063,43.949337],[12.417675,43.955097],[12.433817864,43.955169677],[12.443712102,43.964752043],[12.46723938,43.978675843],[12.490000725,43.985607148],[12.493667602,43.988780975],[12.499191285,43.990112305],[12.501441956,43.993392946],[12.505843162,43.995422363],[12.510206222,43.994190216],[12.511385917,43.995693206],[12.5131464,43.988609314],[12.507678033,43.984661102],[12.509955407,43.976978302],[12.50414753,43.974708558],[12.507146836,43.966300964],[12.504815101,43.961811066],[12.510080338,43.956047058]]]]}, "properties": {"name_0": "San Marino", "uid": 287532, "gid_0": "SMR", "gid_1": "SMR.9_1", "name_1": "Serravalle", "varname_1": " ", "nl_name_1": " ", "iso_1": " ", "hasc_1": "SM.SE", "cc_1": " ", "type_1": "Castello", "name_1_eng": "Municipality", "validfr_1": "Unknown", "sovereign": "San Marino", "lng_center": 12.461037374373083, "lat_center": 43.939040683656}}]}`,
		marino.Geom())
}

func Test_NewWorldBorder(t *testing.T) {
	x, err := NewWorldBorders()
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println("X", x, len(x))
	for key, value := range x {
		fmt.Println(key, value.GetName())
	}
}
