//go:build !(js && wasm)

package save

import (
	"log"
	"os"
	"path/filepath"
)

// DefaultStore saves under the user's config dir, e.g.
// ~/Library/Application Support/LabSim/save.json on macOS.
func DefaultStore() Store {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("no config dir (%v); progress will not be saved", err)
		return &MemStore{}
	}
	return FileStore{Path: filepath.Join(dir, "LabSim", "save.json")}
}
