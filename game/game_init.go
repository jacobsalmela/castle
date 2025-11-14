package game

import (
	"game/assets"
	"game/pkg/config"
	"log"
)

// InitAssets initializes game assets and validates entity bindings.
// This should be called from main() after config is loaded but before NewGame().
//
// Parameters:
//   - cfg: Game configuration. If nil, uses default config.
func InitAssets(cfg *config.Config) {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	log.Println("Initializing game assets...")
	assets.Init(cfg)

	// Validate entity bindings
	if err := ValidateEntityBindings(cfg); err != nil {
		log.Fatalf("Entity binding validation failed: %v", err)
	}
}
