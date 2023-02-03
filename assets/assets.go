package assets

import "embed"

//go:embed cities/*.json
var Cities embed.FS

//go:embed mvt/*.json
var Mvt embed.FS
