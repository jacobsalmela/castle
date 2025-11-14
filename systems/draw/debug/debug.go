package debug

import (
	"game/components"
	"game/ecs"
	"game/pkg/config"

	"github.com/hajimehoshi/ebiten/v2"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PHASE 7: DEBUG OVERLAYS
// Render debug visualizations in device-space for crisp pixel fonts.
// ═══════════════════════════════════════════════════════════════════════════════

// Update renders debug overlays in device-space for crisp pixel fonts.
// This includes:
//   - ECS status overlay (sprite count, debug flags) - Cmd+Shift+1
//   - Performance profiler - Cmd+Shift+0
//   - Debug legend (explains visualization colors) - Auto-shows with active overlays
//   - Debug keyboard shortcuts help - Bottom-left corner
//   - Debug text labels
//
// Call this AFTER scaling the logical buffer to device screen.
//
// Parameters:
//   - world: ECS world instance
//   - screen: Device screen (after viewport scaling)
//   - vp: Viewport for coordinate conversion
func Update(world *ecs.World, screen *ebiten.Image, vp *components.ViewPort) {
	if world == nil || screen == nil || vp == nil {
		return
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	if !cfg.Debug {
		return
	}

	// Draw ECS status overlay with unified debug legends (top-right) - auto-shows when debug overlays are active
	// This includes the collision/physics/hitbox/etc legends in one unified panel
	drawECSOverlayIfEnabled(world, screen, vp)

	// Draw keyboard shortcuts help (bottom-left) - always visible when debug enabled
	// drawDebugKeyboardShortcuts(screen, vp) // TODO: toggle with cmd+shift+h

	// Draw performance profiler (if enabled with Cmd+Shift+0)
	drawProfilerIfEnabled(world, screen)

	// Draw debug text labels in device-space
	drawTextLabels(world, screen, vp)
}
