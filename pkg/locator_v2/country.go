package locator_v2

import (
	"context"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/jackc/pgx/v4/pgxpool"
)

var conn *pgxpool.Pool

// func init() {
// 	// TODO fix that shit.
// 	url := "postgres://postgres:postgres@localhost:5432/mlsch_data"
// 	err := errors.New("")

// 	conn, err = pgxpool.Connect(context.Background(), url)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
// 		os.Exit(1)
// 	}

// }

type Country struct {
	// country name
	name0 string
	// name ob subdivision e.g Bundesland
	name1 string
	// English name
	name1Eng string
	// // Type of subdivision
	// typeC string
	// // Engliish type of subdivision
	// typeCEng  string
	sovereign string
	// Center of Polygon
	lngCenter float64
	latCenter float64
}

func NewWorldBorders() (map[string]Locations, error) {

	// fmt.Println("New World Border called!")

	locs := make(map[string]Locations)

	rows, err := conn.Query(context.Background(), `
		SELECT name_0, name_1, name_1_eng, sovereign, lng_center, lat_center 
		FROM world_borders `)

	if err != nil {
		return map[string]Locations{}, err
	}

	defer rows.Close()

	if err != nil {
		util.Sugar.Fatal("New World Border", err)
	}

	i := 0
	for rows.Next() {
		x := &Country{}
		rows.Scan(&x.name0, &x.name1, &x.name1Eng, &x.sovereign, &x.lngCenter, &x.latCenter)
		locs[x.name0] = x

		if i >= 262 {
			util.Sugar.Fatal("New World Border", err)
		}
		i++
	}
	return locs, nil
}

// Distance Gets the distance. if within its 0
func (c *Country) Distance(lat, lng float64) float64 {
	var distance float64

	err := conn.QueryRow(context.Background(), `
	SELECT ST_Distance(
		ST_Transform((select geom from world_borders WHERE name_0 = $1), 3857),
		ST_Transform(ST_SetSRID( ST_Point($2,$3), 4326), 3857))
	`, c.name0, lng, lat).Scan(&distance)

	if err != nil {
		util.Sugar.Errorw("Distance:", distance, "Error:", err, "Country:", c.name0, "lat, lng", lat, lng)
		return 9999
	}

	util.Sugar.Debugw("Distance",
		"p.name_0", c.GetName(),
		"lat", lat,
		"lng", lng,
		"distance", distance)

	return distance
}

func (c *Country) Geom() string {
	var geojson string
	err := conn.QueryRow(context.Background(),
		`SELECT json_build_object(
		'type', 'FeatureCollection',
		'features', json_agg(ST_AsGeoJSON(t.*)::json)
		)
	FROM (SELECT name_0, ST_Transform(geom,3857) FROM world_borders) t WHERE name_0=$1`,
		c.name0).Scan(&geojson)
	if err != nil {
		util.Sugar.Errorw("Distance",
			"p.name_0", c.GetName(),
			"error", err)
	}
	return geojson
}

// Center returns the coords in array Lat, Lng
func (c *Country) Center() [2]float64 {
	center := [2]float64{}
	center[0], center[1] = c.latCenter, c.lngCenter
	return center
}

func (c *Country) GetName() string {
	return c.name0
}
