package viewport

import (
	"game/components"
	"game/ecs"

	"github.com/hajimehoshi/ebiten/v2"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PHASE 6: VIEWPORT SCALING
// Scale logical buffer to device screen with DPI, scaling, and letterboxing.
// This belongs in draw phase because it's a rendering operation (compositing buffers).
// ═══════════════════════════════════════════════════════════════════════════════

// Update scales the logical screen buffer to the device screen.
// Handles DPI scaling, viewport scaling, and letterboxing offsets.
//
// Parameters:
//   - world: ECS world instance (for viewport)
//   - screen: Device screen (final output)
//   - logicalScreen: Logical buffer to scale
func Update(world *ecs.World, screen, logicalScreen *ebiten.Image) {
	if world == nil || screen == nil || logicalScreen == nil {
		return
	}

	// Get viewport from world resources
	vp := ecs.Resource[components.ViewPort](world)
	if vp == nil {
		return
	}

	// Scale logical buffer to device screen with DPR+Scale+Offset
	scale := vp.Scale * vp.DPR
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(vp.OffsetX*vp.DPR, vp.OffsetY*vp.DPR)
	screen.DrawImage(logicalScreen, op)
}
