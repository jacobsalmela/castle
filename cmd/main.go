package main

import (
	"log"
	"runtime"

	"game/game"
	"game/pkg/config"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// ╔════════════════════════════════════════════════════════════════════════════════
	// ║                            GAME INITIALIZATION
	// ║
	// ║  This function runs ONCE at startup to set up the game.
	// ║  After initialization, Ebitengine takes over and calls these methods
	// ║  repeatedly every frame:
	// ║
	// ║  1. LayoutF() / Layout()  - Calculate screen dimensions
	// ║  2. Update()              - Game logic (60 TPS default)
	// ║  3. Draw(screen)          - Render frame to screen
	// ║
	// ║  See game/game.go for the Game interface implementation.
	// ╚═══════════════════════════════════════════════════════════════════════════════

	// === PHASE 1: CONFIGURATION ===
	cfgPath := "config.yml"
	var err error
	var mergeResult *config.MergeResult
	cfg, mergeResult, err := config.LoadConfigWithMerge(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	// Set global config for backwards compatibility with package-level code
	// that still uses config.Cfg (lighting system, transitions, physics fallbacks)
	// TODO: Remove this once all remaining references are migrated
	config.Cfg = cfg

	// Log initial config merge information
	if mergeResult != nil && len(mergeResult.NewFields) > 0 {
		log.Printf("Initialized config with %d new fields", len(mergeResult.NewFields))
	}

	go config.WatchConfig(cfgPath)

	// === PHASE 2: ASSET INITIALIZATION ===
	// Initialize assets with config (fonts, sprites, maps, procedural images)
	// This must be called before NewGame() which depends on loaded assets
	game.InitAssets(cfg)

	// === PHASE 3: GAME INSTANCE ===
	// Create game instance (loads map, spawns entities, applies save data)
	// See game.NewGame() for detailed initialization flow
	g := game.NewGame(cfg)

	// === PHASE 4: DEVELOPMENT TOOLS ===
	// Start save file watcher (hot-reload save.json for development)
	if cfg.DebugConsole {
		log.Println("Enabling save.json watcher...")
		go game.WatchSave("save.json", g.ReloadSave)
	}

	// === PHASE 5: EBITEN CONFIGURATION ===
	ebiten.SetWindowTitle("Chronovian Thanatome")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	// ebiten.SetVsyncEnabled(!cfg.Debug)
	ebiten.SetTPS(ebiten.SyncWithFPS)
	// Disable linear filtering for crisp pixel art on high DPI displays
	ebiten.SetScreenFilterEnabled(false)

	// Prevents Ebiten from using the Metal backend (avoids Metal-driver
	// warnings and Metal-related runtime panics on certain macOS/driver combos)
	op := &ebiten.RunGameOptions{}
	if runtime.GOOS == "darwin" {
		op.GraphicsLibrary = ebiten.GraphicsLibraryOpenGL
	}

	// === PHASE 6: START GAME LOOP ===
	// Hand control to Ebitengine - it will now call LayoutF(), Update(), Draw() repeatedly
	if err := ebiten.RunGameWithOptions(g, op); err != nil {
		log.Fatal(err)
	}
}
