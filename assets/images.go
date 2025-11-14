package assets

import (
	"game/pkg/config"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// Procedurally-generated images (not file-based)
// These are created at runtime and don't go in AssetRegistry

// HUD bar images (reusable 1px images for scaling)
var (
	HudFullEmptyBarImage  *ebiten.Image
	HudFullBarImage       *ebiten.Image
	HudFullLagBarImage    *ebiten.Image
	HudFullAttackBarImage *ebiten.Image
	HudFullCountBar       *ebiten.Image
)

// Headbar images (enemy health bars)
var (
	HeadbarBarImage       *ebiten.Image
	HeadbarInnerBarImage  *ebiten.Image
	HeadbarLagBarImage    *ebiten.Image
	HeadbarFillerBarImage *ebiten.Image
)

// Transition images
var (
	TransitionFadeImage *ebiten.Image // Black screen-sized image for fade effects
)

// initProceduralAssets creates all procedurally-generated images (HUD bars, health bars, transitions).
// This is exported for use by the unified asset loader.
func initProceduralAssets(cfg *config.Config) error {
	initHudBarImages(cfg)
	if cfg.DebugConsole {
		log.Println("  ✓ HUD bar images initialized")
	}

	initHeadbarImages(cfg)
	if cfg.DebugConsole {
		log.Println("  ✓ Headbar images initialized")
	}

	initTransitionImages(cfg)
	if cfg.DebugConsole {
		log.Println("  ✓ Transition images initialized")
	}

	return nil
}

// initHudBarImages creates the reusable 1px bar images for HUD rendering.
// These images are scaled during rendering to create bars of various lengths.
func initHudBarImages(cfg *config.Config) {
	emptyColor := color.RGBA{89, 86, 82, 255}
	lagColor := color.RGBA{251, 242, 54, 255}
	borderColor := color.RGBA{34, 32, 52, 255}

	HudFullEmptyBarImage = ebiten.NewImage(1, cfg.Hud.InnerBarH)
	HudFullEmptyBarImage.Fill(emptyColor)

	HudFullBarImage = ebiten.NewImage(1, cfg.Hud.BarH)

	HudFullLagBarImage = ebiten.NewImage(1, cfg.Hud.InnerBarH)
	HudFullLagBarImage.Fill(lagColor)

	HudFullAttackBarImage = ebiten.NewImage(1, cfg.Hud.BarH)
	HudFullAttackBarImage.Fill(borderColor)

	HudFullCountBar = ebiten.NewImage(1, cfg.Hud.BarH+2)
	HudFullCountBar.Fill(borderColor)
}

// initHeadbarImages creates the reusable images for enemy health bars.
func initHeadbarImages(cfg *config.Config) {
	borderColor := color.RGBA{34, 32, 52, 255}
	emptyColor := color.RGBA{89, 86, 82, 255}
	lagColor := color.RGBA{251, 242, 54, 255}
	healthColor := color.RGBA{172, 50, 50, 255}

	HeadbarBarImage = ebiten.NewImage(int(cfg.Hud.EnemyBarW)+2, cfg.Hud.InnerBarH)
	HeadbarBarImage.Fill(borderColor)

	HeadbarInnerBarImage = ebiten.NewImage(int(cfg.Hud.EnemyBarW), 1)
	HeadbarInnerBarImage.Fill(emptyColor)

	HeadbarLagBarImage = ebiten.NewImage(1, 1)
	HeadbarLagBarImage.Fill(lagColor)

	HeadbarFillerBarImage = ebiten.NewImage(1, 1)
	HeadbarFillerBarImage.Fill(healthColor)
}

// initTransitionImages creates the reusable images for transitions.
func initTransitionImages(cfg *config.Config) {
	blackColor := color.RGBA{0, 0, 0, 255}

	TransitionFadeImage = ebiten.NewImage(int(cfg.Screen.Width), int(cfg.Screen.Height))
	TransitionFadeImage.Fill(blackColor)
}
