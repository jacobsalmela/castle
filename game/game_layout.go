package game

import (
	"game/assets/fonts"
	"game/ecs"
	"game/pkg/config"
	"math"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
)

// ╔════════════════════════════════════════════════════════════════════════════════╗
// ║                     EBITENGINE FRAME STEP 1: LAYOUT                            ║
// ║                                                                                ║
// ║  Layout() is called by Ebitengine to determine screen dimensions.             ║
// ║  This is called:                                                               ║
// ║  • On startup                                                                  ║
// ║  • When window is resized                                                      ║
// ║  • When DPI/scaling changes (e.g., moving to different monitor)               ║
// ║                                                                                ║
// ║  Layout() → LayoutF() (the real implementation with float64 precision)        ║
// ║                                                                                ║
// ║  Next step: Update() in game_update.go                                        ║
// ╚════════════════════════════════════════════════════════════════════════════════╝

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := g.LayoutF(float64(outsideWidth), float64(outsideHeight))
	return int(w), int(h)
}

func (g *Game) LayoutF(outsideWidth, outsideHeight float64) (screenWidth, screenHeight float64) {
	vp := g.viewport

	vp.OW, vp.OH = outsideWidth, outsideHeight

	cfg := config.Cfg // Fallback to global during early initialization
	if g.world != nil {
		if worldCfg := ecs.Resource[config.Config](g.world); worldCfg != nil {
			cfg = worldCfg
		}
	}

	// logical back‑buffer size (design units) (hard-coded or from config)
	logicalW, logicalH := cfg.Screen.Width, cfg.Screen.Height

	if runtime.GOOS == "ios" || runtime.GOOS == "android" {
		if outsideWidth < outsideHeight { // portrait
			logicalW, logicalH = logicalH, logicalW
		}
	}

	// this is the logical/design space and is fixed
	vp.LW, vp.LH = logicalW, logicalH

	// Scale logical to outside dimensions
	vp.Scale = math.Min(outsideWidth/logicalW, outsideHeight/logicalH)

	// Device pixel ratio - respect HighDpi config setting
	// Use manual DpiScale if configured (non-zero), otherwise auto-detect
	deviceScale := ebiten.Monitor().DeviceScaleFactor()

	if !cfg.Screen.HighDpi {
		vp.DPR = 1.0 // HighDPI disabled
	} else if cfg.Screen.DpiScale > 0 {
		vp.DPR = cfg.Screen.DpiScale // Manual override
	} else {
		vp.DPR = deviceScale // Auto-detect
	}

	// Build fonts for logical pixels (1.0) so bitmap/pixel fonts remain sharp
	// when rendered into the logical backbuffer and then scaled to device pixels.
	fonts.Ensure(vp.DPR)

	// fit as large as possible while preserving the aspect ratio (in points)
	// this is the size the game will occupy on screen (in points)
	vp.PW = logicalW * vp.Scale
	vp.PH = logicalH * vp.Scale
	// letterbox offset in points
	// vp.OffsetX = (outsideWidth - vp.PW) * 0.5
	// vp.OffsetY = (outsideHeight - vp.PH) * 0.5

	vp.PX = vp.PW * vp.DPR
	vp.PY = vp.PH * vp.DPR

	// Logical screen buffer now managed by systems/draw package as ECS resource

	return vp.PX, vp.PY
}
