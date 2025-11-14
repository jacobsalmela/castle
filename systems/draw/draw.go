package draw

import (
	"game/components"
	"game/ecs"
	"game/resources"
	"game/systems/draw/buffers"
	"game/systems/draw/debug"
	"game/systems/draw/lighting"
	"game/systems/draw/postprocess"
	"game/systems/draw/ui"
	"game/systems/draw/viewport"
	"game/systems/draw/world"

	"github.com/hajimehoshi/ebiten/v2"
)

// SpriteDrawState provides camera information for sprite and UI rendering.
// This interface is implemented by *ecs.World via its Camera field.
type SpriteDrawState interface {
	CameraInFrameRecter(resources.Recter, float64, float64) bool
	CameraPosition() (float64, float64)
}

// ╔════════════════════════════════════════════════════════════════════════════════╗
// ║                     SYSTEMS/DRAW: RENDERING ORCHESTRATION                      ║
// ║                                                                                ║
// ║  This is called from systems.Draw() every frame (60 FPS).                     ║
// ║                                                                                ║
// ║  Flow:                                                                         ║
// ║  DrawFrame(world, screen)                                                      ║
// ║    ├─ buffers.Update()           (create/clear logical screen, normal map)    ║
// ║    ├─ world.Update()             (render map, sprites, compose)               ║
// ║    ├─ lighting.Update()          (apply lighting shader)                      ║
// ║    ├─ ui.Update()                (render HUD, headbars, textboxes)            ║
// ║    ├─ postprocess.Update()       (transitions, effects)                       ║
// ║    ├─ viewport.Update()          (scale to device screen)                     ║
// ║    └─ debug.Update()             (debug overlays)                             ║
// ║                                                                                ║
// ║  All rendering logic lives in systems/draw/* subpackages.                     ║
// ║  This file is a pure orchestrator matching systems/update.go pattern.         ║
// ╚════════════════════════════════════════════════════════════════════════════════╝

// DrawFrame renders the complete game frame to the device screen.
// This is a pure orchestrator that delegates to draw subpackages.
//
// Parameters:
//   - w: ECS world instance
//   - screen: Device screen to render to
func DrawFrame(w *ecs.World, screen *ebiten.Image) {
	if w == nil || screen == nil {
		return
	}

	// Get viewport from world resources (needed by multiple phases)
	vp := ecs.Resource[components.ViewPort](w)
	if vp == nil {
		return
	}

	// PHASE 1: Create and clear rendering buffers
	logicalScreen, normalMap := buffers.Update(w, vp)
	if logicalScreen == nil {
		return
	}

	// PHASE 2: Render world (map, sprites, composition)
	world.Update(w, logicalScreen, normalMap, vp)

	// PHASE 3: Apply lighting shader
	lighting.Update(w, logicalScreen, normalMap)

	// PHASE 4: Render UI (HUD, headbars, textboxes)
	ui.Update(w, logicalScreen)

	// PHASE 5: Apply post-processing effects
	postprocess.Update(w, logicalScreen)

	// PHASE 6: Scale logical buffer to device screen
	viewport.Update(w, screen, logicalScreen)

	// PHASE 6.5: Render textbox text in device space for crisp fonts
	ui.RenderTextboxesDeviceSpace(w, screen, vp)
	ui.RenderDismissedIndicatorDeviceSpace(w, screen, vp)

	// PHASE 7: Render debug overlays in device space
	debug.Update(w, screen, vp)
}
