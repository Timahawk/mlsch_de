package assets

import "embed"

//go:embed cities/*.json
var Cities embed.FS
