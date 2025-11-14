package postprocess

import (
	"game/ecs"
	"game/pkg/config"
	"game/resources"

	"github.com/hajimehoshi/ebiten/v2"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PHASE 5: POST PROCESSING
// Apply transition effects (fades, wipes, etc).
// ═══════════════════════════════════════════════════════════════════════════════

// Update renders transition effects to the logical screen.
// Examples: fade in/out, wipe transitions, screen shake effects.
//
// Parameters:
//   - world: ECS world instance
//   - screen: Logical screen buffer
func Update(world *ecs.World, screen *ebiten.Image) {
	if world == nil || screen == nil {
		return
	}

	// Draw transitions via ECS resource (fades, wipes, etc.)
	tm := ecs.Resource[resources.TransitionManager](world)
	if tm != nil && tm.IsActive() {
		cfg := ecs.Resource[config.Config](world)
		tm.Draw(screen, cfg)
	}
}
