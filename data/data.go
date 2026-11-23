// Package data embeds the game's content files (P-8: tune without code changes).
package data

import _ "embed"

//go:embed map.json
var Map []byte

//go:embed text.json
var Text []byte

//go:embed avatar.json
var Avatar []byte

//go:embed names.json
var Names []byte

//go:embed stations.json
var Stations []byte

//go:embed items.json
var Items []byte

//go:embed spawns.json
var Spawns []byte

//go:embed jobs.json
var Jobs []byte

//go:embed animals.json
var Animals []byte

//go:embed discoveries.json
var Discoveries []byte
