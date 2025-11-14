package game

import (
	"game/systems"

	"github.com/hajimehoshi/ebiten/v2"
)

// ╔════════════════════════════════════════════════════════════════════════════════╗
// ║                     EBITENGINE FRAME STEP 3: DRAW                              ║
// ║                                                                                ║
// ║  Draw() is called by Ebitengine to render the game frame.                      ║
// ║  This is called at FPS rate (synced with monitor refresh, typically 60fps).    ║
// ║                                                                                ║
// ║  Flow:                                                                         ║
// ║  Game.Draw(screen)                                                             ║
// ║    └─ systems.Draw()              (systems/draw.go)                            ║
// ║                                                                                ║
// ║  Game.Draw() is a thin orchestrator - all logic lives in systems/draw/*        ║
// ║                                                                                ║
// ║  End of frame - Ebitengine presents screen to display, then loops back to      ║
// ║  step 1 (LayoutF if needed, otherwise Update).                                 ║
// ╚════════════════════════════════════════════════════════════════════════════════╝

func (g *Game) Draw(screen *ebiten.Image) {
	// Guard against nil world during reset transitions
	if g.world == nil {
		return
	}

	// Delegate all rendering to systems package (passing world explicitly)
	systems.Draw(g.world, screen)
}
