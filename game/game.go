package game

import (
	"log"

	"game/components"
	"game/ecs"
	"game/pkg/config"

	_ "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// ╔════════════════════════════════════════════════════════════════════════════════╗
// ║                          EBITENGINE GAME LOOP FLOW                             ║
// ║                                                                                ║
// ║  Ebitengine calls these methods repeatedly every frame (in order):            ║
// ║                                                                                ║
// ║  1. LayoutF(outsideW, outsideH float64) (screenW, screenH float64)           ║
// ║     └─ Calculate screen dimensions and viewport scaling                       ║
// ║     └─ Called BEFORE Update/Draw when window resizes or DPI changes          ║
// ║     └─ See: game/game_layout.go                                              ║
// ║                                                                                ║
// ║  2. Update() error                                                            ║
// ║     └─ Runs game logic (input, physics, AI, collisions, etc.)                ║
// ║     └─ Called at TPS rate (60 times per second by default)                   ║
// ║     └─ See: game/game_update.go → systems/update.go                          ║
// ║                                                                                ║
// ║  3. Draw(screen *ebiten.Image)                                                ║
// ║     └─ Renders the game frame to the screen buffer                            ║
// ║     └─ Called at FPS rate (synced with monitor refresh)                      ║
// ║     └─ See: game/game_draw.go → systems/draw.go                              ║
// ║                                                                                ║
// ║  Key Points:                                                                   ║
// ║  • Layout may be called less frequently (only when window/DPI changes)        ║
// ║  • Update and Draw are called every frame                                     ║
// ║  • Update runs at fixed TPS (game logic), Draw at variable FPS (rendering)   ║
// ║  • All actual logic lives in systems/ package (game/ just orchestrates)      ║
// ╚════════════════════════════════════════════════════════════════════════════════╝

type Game struct {
	world    *ecs.World
	viewport *components.ViewPort

	// saveData holds the current save state (player position, opened items, etc.)
	saveData *SaveData

	// layoutDebug stores debug state for Layout function
	layoutDebug interface{}
}

// NewGame constructs and initializes a new game instance.
// This is a thin Ebitengine adapter that delegates all world initialization
// to the world.go file.
//
// Parameters:
//   - cfg: Game configuration. If nil, uses default config.
//
// All initialization logic lives in world.go - this just constructs the Game struct.
func NewGame(cfg *config.Config) *Game {
	result, err := InitializeWorld(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize game: %v", err)
	}

	return &Game{
		world:    result.World,
		viewport: result.Viewport,
		saveData: result.SaveData,
	}
}
