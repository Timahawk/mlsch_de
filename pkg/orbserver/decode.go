package orbserver

import (
	"embed"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/mvt"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/maptile"
	"github.com/paulmach/orb/simplify"
	"github.com/rs/zerolog/log"
)

// deepCopy returns a DeepCopy only of the included orb.Geometries.
// Properties are only shallow copies. Still VERY Memory intensive.
// Needed because [mvt.Layers] projects all underlying geoms.
func deepCopy(fc *geojson.FeatureCollection) *geojson.FeatureCollection {
	fcCp := new(geojson.FeatureCollection)

	fcCp.Type = fc.Type
	fcCp.BBox = fc.BBox
	fcCp.ExtraMembers = fc.ExtraMembers.Clone()

	for _, feature := range fc.Features {

		newFc := &geojson.Feature{}

		xtype := reflect.TypeOf(feature.Geometry)
		la := fmt.Sprintf("%v", xtype)
		switch la {
		case "orb.MultiPolygon":
			xvalue := reflect.ValueOf(feature.Geometry).Interface().(orb.MultiPolygon)
			newFc = geojson.NewFeature(xvalue.Clone())
		case "orb.Polygon":
			xvalue := reflect.ValueOf(feature.Geometry).Interface().(orb.Polygon)
			newFc = geojson.NewFeature(xvalue.Clone())
		case "orb.Point":
			xvalue := reflect.ValueOf(feature.Geometry).Interface().(orb.Point)
			newFc.Geometry = xvalue
		case "orb.LineString":
			xvalue := reflect.ValueOf(feature.Geometry).Interface().(orb.LineString)
			newFc = geojson.NewFeature(xvalue.Clone())
		case "orb.MultiLineString":
			xvalue := reflect.ValueOf(feature.Geometry).Interface().(orb.MultiLineString)
			newFc = geojson.NewFeature(xvalue.Clone())
		default:
			log.Warn().Str("la", la).Msg("This should not have been called!")
		}

		newFc.BBox = feature.BBox
		newFc.Type = feature.Type
		newFc.Properties = feature.Properties.Clone()
		newFc.ID = feature.ID
		fcCp.Features = append(fcCp.Features, newFc)

	}

	return fcCp
}

func LoadEmbeddedFC(f embed.FS) map[string]*geojson.FeatureCollection {
	entries, err := f.ReadDir("mvt")
	if err != nil {
		log.Panic()
	}

	collections := map[string]*geojson.FeatureCollection{}

	for _, entry := range entries {
		start := time.Now()
		// log.Info().Str("Layer", entry.Name()).Msg("")
		str, err := f.ReadFile("mvt/" + entry.Name())
		if err != nil {
			log.Panic()
		}
		fc, err := geojson.UnmarshalFeatureCollection(str)
		if err != nil {
			log.Panic().Stack().Err(err).Msg("")
		}

		// Needed because I filter on the Bounding Box to include the layer.
		collection := orb.Collection{}
		for _, feature := range fc.Features {
			collection = append(collection, feature.Geometry)
		}
		bound := collection.Bound()
		fc.BBox = geojson.NewBBox(bound)
		log.Info().Str("geojson", entry.Name()).Dur("Duration(ms)", time.Since(start)).Msg("Loading successful")
		collections[entry.Name()] = fc

	}

	return collections
}

// LoadFeatureClasses takes a pattern and tries to unmarshall all files that match.
// The name of the layer will be the name of the file without extensions.
//
//	LoadFeatureClasses("./data/*.json") // -> Load all .json files in subdirectory "data"
//	LoadFeatureClasses("./data/countries.json") -> filename will be "countries"
func LoadFeatureClasses(pattern string) map[string]*geojson.FeatureCollection {

	collections := make(map[string]*geojson.FeatureCollection)

	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("")
	}
	log.Info().Int("Number", len(files)).Msg("Searching Files")

	for _, filePath := range files {

		start := time.Now()
		// Get only the name of the file without extension
		file := filepath.Base(filePath)
		file = file[:len(file)-len(filepath.Ext(file))]

		str, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal().Stack().Err(err).Msg("")
		}
		fc, err := geojson.UnmarshalFeatureCollection(str)
		if err != nil {
			log.Fatal().Stack().Err(err).Msg("")
		}

		// Needed because I filter on the Bounding Box to include the layer.
		collection := orb.Collection{}
		for _, feature := range fc.Features {
			collection = append(collection, feature.Geometry)
		}
		bound := collection.Bound()
		fc.BBox = geojson.NewBBox(bound)
		log.Info().Str("Layer", file).Str("Bound", fmt.Sprintf("%v", bound)).Msg("")

		collections[file] = fc

		log.Info().Str("geojson", file).Dur("Duration(ms)", time.Since(start)).Msg("Loading successfull")
	}
	return collections
}

// num2deg calculates the North-West Lat Long Point for tile x,y at zoom z.
func num2deg(x, y, z int) (lat float64, long float64) {
	n := math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long = float64(x)/math.Exp2(float64(z))*360.0 - 180.0
	return lat, long
}

// MvtGin takes a map of geojson collections and returns a gin.HandlerFunc.
// The func takes the input parameters for the tile x,y,z and
// returns the fitting pbf tile. Only FeatureCollections where the bounds
// overlap with the bounds of the tile are included in the response.
// Geometries are not simplified. Data is returned unzipped.
// For each requests a [DeepCopy] of all FeatureCollections is createad, which makes this very memory intensive.
//
//	r := gin.Default()
//	r.GET("/mvt/:z/:x/:y/pbf", jsonorb.Handler(collections))
func MvtGin(collections map[string]*geojson.FeatureCollection) gin.HandlerFunc {

	fn := func(c *gin.Context) {

		z, err := strconv.Atoi(c.Param("z"))
		if err != nil {
			log.Fatal().Stack().Err(err).Str("Param", c.Param("z")).Msg("Z")
		}
		x, err := strconv.Atoi(c.Param("x"))
		if err != nil {
			log.Fatal().Stack().Err(err).Str("Param", c.Param("z")).Msg("X")
		}
		y, err := strconv.Atoi(c.Param("y"))
		if err != nil {
			log.Fatal().Stack().Err(err).Str("Param", c.Param("z")).Msg("Y")
		}

		layers := mvt.Layers{}

		lat, lng := num2deg(x, y, z)
		northWest := orb.Point{lng, lat} // NorthWest due to 0,0 is -180;90 in WGS84

		lat, lng = num2deg(x+1, y+1, z) // +1 for south_east.
		southEast := orb.Point{lng, lat}

		bound := orb.Collection{northWest, southEast}.Bound()

		for key, value := range collections {

			if bound.Intersects(value.BBox.Bound()) {
				layers = append(layers, mvt.NewLayer(key, deepCopy(value)))
			}
		}

		layers.ProjectToTile(
			maptile.New(
				uint32(x),
				uint32(y),
				maptile.Zoom(z)))

		// to correct extent
		layers.Clip(orb.Bound{
			Min: orb.Point{0, 0},
			Max: orb.Point{4096, 4096},
		})

		// Simplify the geometry now that it's in the tile coordinate space.
		layers.Simplify(simplify.DouglasPeucker(1.0))

		// Depending on use-case remove empty geometry, those two small to be
		// represented in this tile space.
		// In this case lines shorter than 1, and areas smaller than 1.
		layers.RemoveEmpty(1.0, 1.0)

		// encoding using the Mapbox Vector Tile protobuf encoding.
		data, err := mvt.Marshal(layers) // this data is NOT gzipped.
		if err != nil {
			log.Fatal().Stack().Err(err).Msg("")
		}
		c.Data(http.StatusOK, "application/x-protobuf", data)
	}

	return fn
}

// type MVT struct {
// 	FC geojson.FeatureCollection
// }

// func (mvt *MVT) ServeHTTP(w http.ResponseWriter, r *http.Request) {

// }

// func MVTHandler(h http.HandlerFunc, fc *geojson.FeatureCollection) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		h(w, r)
// 	}

// }
