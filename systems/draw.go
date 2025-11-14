package systems

import (
	"game/ecs"
	"game/systems/draw"

	"github.com/hajimehoshi/ebiten/v2"
)

// ╔════════════════════════════════════════════════════════════════════════════════╗
// ║                     SYSTEMS PACKAGE: DRAW ORCHESTRATION                        ║
// ║                                                                                ║
// ║  This is called from game/game_draw.go every frame (60 FPS).                   ║
// ║                                                                                ║
// ║  Flow:                                                                         ║
// ║  systems.Draw()                                                                ║
// ║    └─ draw.DrawFrame()            (systems/draw/draw.go - 7 phases)            ║
// ║       ├─ buffers.Update()         (logical screen, normal map)                 ║
// ║       ├─ world.Update()           (map, sprites, composition)                  ║
// ║       ├─ lighting.Update()        (shader effects)                             ║
// ║       ├─ ui.Update()              (HUD, health bars)                           ║
// ║       ├─ postprocess.Update()     (transitions, effects)                       ║
// ║       ├─ viewport.Update()        (scale to device screen)                     ║
// ║       └─ debug.Update()           (device-space overlays)                      ║
// ║                                                                                ║
// ║  All rendering logic lives in systems/draw/* - this file just orchestrates.    ║
// ╚════════════════════════════════════════════════════════════════════════════════╝

// Draw renders the complete game frame to the screen.
// This is a pure orchestrator that delegates to draw subpackages.
//
// Parameters:
//   - world: ECS world instance
//   - screen: Ebiten screen to render to
func Draw(world *ecs.World, screen *ebiten.Image) {
	if world == nil {
		return
	}

	// Draw the complete frame using draw package
	draw.DrawFrame(world, screen)
}
