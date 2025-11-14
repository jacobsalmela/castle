package posttick

import (
	"game/ecs"
	"game/resources"
	"game/systems/draw/lighting"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Update handles post-tick operations after the main ECS update loop.
// This includes:
//   - Frame cleanup (reset frame-specific component state)
//   - Shader updates (lighting animation)
//   - Game-level input handling (ESC to quit)
//   - Deferred reset execution
//
// Returns ebiten.Termination if user presses ESC to quit.
func Update(world *ecs.World, dt float64) error {
	// === FRAME CLEANUP ===
	// Reset frame-specific component state
	ResetActionIntents(world)

	// === GAME-SPECIFIC EFFECTS ===
	// Shader updates (lighting, torch effects, etc.)
	lighting.UpdateTime(dt)

	// Update camera (smooth following, room transitions, shake)
	if camera := ecs.Resource[resources.Camera](world); camera != nil {
		camera.Update(dt)
	}

	// === INPUT ===
	// Handle game-level input (ESC to quit)
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	// === DEFERRED RESET ===
	// Execute any pending reset after all systems have finished updating
	// This prevents race conditions between Update and Draw
	if signals := ecs.Resource[resources.GameSignals](world); signals != nil {
		signals.ConsumePendingReset()
	}

	return nil
}
