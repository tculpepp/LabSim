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
