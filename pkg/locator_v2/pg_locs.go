package locator_v2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Timahawk/mlsch_de/pkg/util"
	"github.com/jackc/pgx/v4/pgxpool"
)

var conn *pgxpool.Pool

func init() {

	url := "postgres://postgres:postgres@localhost:5432/mlsch_data"
	err := errors.New("")

	conn, err = pgxpool.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	// defer conn.Close(context.Background())

	// var x pg_locs
	// err = conn.QueryRow(context.Background(), "select name_0, name_1, name_1_eng, sovereign, lng_center, lat_center from  world_borders WHERE name_0 = $1", "Germany").Scan(&x.name_0, &x.name_1, &x.name_1_eng, &x.sovereign, &x.lng_center, &x.lat_center)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Println("X = ", x)
}

type pg_locs struct {
	// country name
	name_0 string
	// name ob subdivision e.g Bundesland
	name_1 string
	// English name
	name_1_eng string
	// Type of subdivision
	type_c string
	// Engliish type of subdivision
	type_c_eng string
	sovereign  string
	// Center of Polygon
	lng_center float64
	lat_center float64
}

func NewWorldBorders() (map[string]Locations, error) {

	fmt.Println("New World Border called!")

	locs := make(map[string]Locations)

	rows, err := conn.Query(context.Background(), `
		SELECT name_0, name_1, name_1_eng, sovereign, lng_center, lat_center 
		FROM world_borders `)

	defer rows.Close()

	if err != nil {
		log.Fatalln("New World Border", err)
	}
	for rows.Next() {
		x := &pg_locs{}
		rows.Scan(&x.name_0, &x.name_1, &x.name_1_eng, &x.sovereign, &x.lng_center, &x.lat_center)
		locs[x.name_0] = x
	}
	return locs, nil
}

// Gets the distance. if within its 0
func (p *pg_locs) Distance(lat, lng float64) float64 {
	var distance float64

	err := conn.QueryRow(context.Background(), `
	SELECT ST_Distance(
		ST_Transform((select geom from world_borders WHERE name_0 = $1), 3857),
		ST_Transform(ST_SetSRID( ST_Point($2,$3), 4326), 3857))
	`, p.name_0, lng, lat).Scan(&distance)

	if err != nil {
		log.Fatalln("Distance", err)
	}

	util.Sugar.Warnw("Distance",
		"p.name_0", p.GetName(),
		"lat", lat,
		"lng", lng,
		"distance", distance)

	return distance
}

func (p *pg_locs) Geom() string {
	var geojson string
	err := conn.QueryRow(context.Background(),
		`SELECT json_build_object(
		'type', 'FeatureCollection',
		'features', json_agg(ST_AsGeoJSON(t.*)::json)
		)
	FROM (SELECT name_0, ST_Transform(geom,3857) FROM world_borders) t WHERE name_0=$1`,
		p.name_0).Scan(&geojson)
	if err != nil {
		log.Fatalln("Geom", err)
	}
	return geojson
}

// Center returns the coords in array Lat, Lng
func (p *pg_locs) Center() [2]float64 {
	center := [2]float64{}
	center[0], center[1] = p.lat_center, p.lng_center
	return center
}

func (p *pg_locs) GetName() string {
	return p.name_0
}
