package assets

import (
	embed "embed"
	"game/assets/fonts"
	"game/assets/maps"
	"game/pkg/config"
	"log"
	"sync"
)

var (
	// FS is the embedded file system for all assets.
	// Embeds all PNG and JSON files in assets/ root and images/ subdirectories
	//go:embed images/**/*.png images/**/*.json
	FS embed.FS

	// Note: Map assets are embedded separately in assets/maps/maps.go
	// to provide a dedicated namespace for map data
)

var once sync.Once
var wg sync.WaitGroup

// Init loads all assets synchronously.
// It is safe to call multiple times.
//
// Parameters:
//   - cfg: Game configuration. If nil, uses default config.
func Init(cfg *config.Config) {
	once.Do(func() {
		if cfg == nil {
			cfg = config.NewDefaultConfig()
		}

		log.Println("Loading font assets...")
		fonts.Init()

		// New unified loader (replaces InitImages + LoadAllSlices)
		if err := LoadAllAssets(FS, cfg.DebugConsole); err != nil {
			log.Printf("Error loading sprite assets: %v", err)
		}

		// Load procedural images (HUD bars, etc.)
		if err := initProceduralAssets(cfg); err != nil {
			log.Printf("Error loading procedural images: %v", err)
		}

		maps.Init(cfg.DebugConsole)

		// log.Println("Loading audio assets...")
		// audio.Init()
	})
}
