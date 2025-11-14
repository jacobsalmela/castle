package maps

import (
	"embed"
	"log"
)

var (
	// FS is the embedded file system for all map assets.
	// Embeds all map files (.tmx, .tsx, .png) from the maps/ subdirectory
	//

	//go:embed **/*.png **/*.tmx **/*.tsx
	FS embed.FS
)

// Init initializes map assets.
// Maps are loaded on-demand by the tilemap package using the FS embed.
//
// Parameters:
//   - debugConsole: Whether to log verbose debug messages
func Init(debugConsole bool) {
	if debugConsole {
		log.Println("  ✓ Map assets ready (embed.FS)")
	}
	// Future: Could add map validation, preprocessing, or caching here
}
