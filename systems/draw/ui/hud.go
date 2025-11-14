package ui

import (
	"fmt"
	"game/assets"
	"game/assets/fonts"
	"game/components"
	"game/ecs"
	"game/pkg/config"
	"game/resources"
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	text2 "github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	minAttackMultToShow = 0.1
	staminaVisualScale  = 0.8
)

var (
	// HUD images - loaded lazily on first use
	hudImage       *ebiten.Image
	barEndImage    *ebiten.Image
	middleBarImage *ebiten.Image
	iconsImage     *ebiten.Image

	// Colors
	healthColor  = color.RGBA{172, 50, 50, 255}
	staminaColor = color.RGBA{55, 148, 110, 255}
)

// ensureHUDImagesLoaded lazily loads HUD images from the unified asset system.
// This is called once on first RenderHUD() call to avoid init-time loading issues.
func ensureHUDImagesLoaded(cfg *config.Config) {
	if hudImage != nil {
		return // Already loaded
	}

	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Try to get from unified asset system first
	hudImage = assets.GetSpriteImage("hud")
	if hudImage == nil {
		// Fallback: try to load directly
		var err error
		hudImage, _, err = ebitenutil.NewImageFromFileSystem(assets.FS, "images/ui/hud.png")
		if err != nil {
			// Create fallback images if loading fails
			hudImage = ebiten.NewImage(64, cfg.Hud.BarH)
			barEndImage = ebiten.NewImage(cfg.Hud.BarEndX2-cfg.Hud.BarEndX1, cfg.Hud.BarH)
			middleBarImage = ebiten.NewImage(cfg.Hud.MiddleBarX2-cfg.Hud.MiddleBarX1, cfg.Hud.BarH)
			iconsImage = ebiten.NewImage(int(cfg.Hud.HudIconsX), cfg.Hud.BarH)
			return
		}
	}

	// Slice the HUD image into components
	barEndImage = hudImage.SubImage(image.Rect(cfg.Hud.BarEndX1, 0, cfg.Hud.BarEndX2, cfg.Hud.BarH)).(*ebiten.Image)
	middleBarImage = hudImage.SubImage(image.Rect(cfg.Hud.MiddleBarX1, 0, cfg.Hud.MiddleBarX2, cfg.Hud.BarH)).(*ebiten.Image)
	iconsImage = hudImage.SubImage(image.Rect(0, 0, int(cfg.Hud.HudIconsX), hudImage.Bounds().Dy())).(*ebiten.Image)
}

// RenderHUD draws the player HUD using the render queue.
func RenderHUD(world *ecs.World, queue *resources.RenderQueue, hud *components.HUDData) {
	if world == nil || queue == nil || hud == nil {
		return
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Ensure HUD images are loaded (lazy initialization)
	ensureHUDImagesLoaded(cfg)

	baseGeoM := ebiten.GeoM{}
	baseGeoM.Translate(1, 1)

	// Draw HUD icons background
	queue.Push(resources.RenderCommand{
		Image:      iconsImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       baseGeoM,
	})

	// Draw icons to normal map (punch out for lighting)
	iconNormalGeoM := baseGeoM
	queue.Push(resources.RenderCommand{
		Image:      iconsImage,
		TargetType: resources.TargetNormal,
		Layer:      config.PipelineUILayer,
		GeoM:       iconNormalGeoM,
		ColorScale: components.NormalMaskColor, // Used for blend mode punch-out
	})

	// Draw health bar
	renderSegment(queue, baseGeoM, 0, hud.Health, hud.MaxHealth, hud.HealthLag, healthColor, cfg)

	// Draw stamina bar (with visual scaling)
	renderSegment(
		queue,
		baseGeoM,
		1,
		hud.Stamina*staminaVisualScale,
		hud.MaxStamina*staminaVisualScale,
		hud.StaminaLag*staminaVisualScale,
		staminaColor,
		cfg,
	)

	// Draw attack multiplier (if above threshold)
	if hud.ShowAttackMult {
		renderAttackMult(world, queue, baseGeoM, hud.AttackMult, hud.MaxHealth)
	}

	// Draw heal count
	renderCount(world, queue, baseGeoM, 2, hud.Heal, 0)

	// Draw experience
	renderCount(world, queue, baseGeoM, 3, hud.Exp, 2)
}

// renderSegment draws a single HUD bar segment (health or stamina).
func renderSegment(queue *resources.RenderQueue, baseGeoM ebiten.GeoM, y, current, max, lag float64, barColor color.Color, cfg *config.Config) {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Normal map punch-out for bar background
	normalGeoM := ebiten.GeoM{}
	normalGeoM.Scale(max+2, 1)
	normalGeoM.Concat(baseGeoM)
	normalGeoM.Translate(float64(cfg.Hud.HudIconsX), float64(cfg.Hud.BarMiddleH)*y)

	queue.Push(resources.RenderCommand{
		Image:      assets.HudFullAttackBarImage,
		TargetType: resources.TargetNormal,
		Layer:      config.PipelineUILayer,
		GeoM:       normalGeoM,
		ColorScale: components.NormalMaskColor,
	})

	// Bar background border
	barGeoM := ebiten.GeoM{}
	barGeoM.Scale(max, 1)
	barGeoM.Concat(baseGeoM)
	barGeoM.Translate(float64(cfg.Hud.HudIconsX), float64(cfg.Hud.BarMiddleH)*y)

	queue.Push(resources.RenderCommand{
		Image:      middleBarImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       barGeoM,
	})

	// Empty bar filler
	fillerGeoM := ebiten.GeoM{}
	fillerGeoM.Concat(baseGeoM)
	fillerGeoM.Translate(float64(cfg.Hud.HudIconsX), float64(cfg.Hud.BarMiddleH)*y+2)

	emptyGeoM := ebiten.GeoM{}
	emptyGeoM.Scale(max, 1)
	emptyGeoM.Concat(fillerGeoM)

	queue.Push(resources.RenderCommand{
		Image:      assets.HudFullEmptyBarImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       emptyGeoM,
	})

	// Lag indicator
	if lag > 0 {
		lagGeoM := ebiten.GeoM{}
		lagGeoM.Scale(lag+1, 1)
		lagGeoM.Concat(fillerGeoM)

		queue.Push(resources.RenderCommand{
			Image:      assets.HudFullLagBarImage,
			TargetType: resources.TargetScreen,
			Layer:      config.PipelineUILayer,
			GeoM:       lagGeoM,
		})
	}

	// Current value bar
	if current > 0 {
		// Create a fresh bar image with the appropriate color for this segment
		// Must create new image because queue executes later (can't reuse shared image)
		coloredBarImage := ebiten.NewImage(1, cfg.Hud.InnerBarH)
		coloredBarImage.Fill(barColor)

		currentGeoM := ebiten.GeoM{}
		currentGeoM.Scale(current+1, 1)
		currentGeoM.Concat(fillerGeoM)

		queue.Push(resources.RenderCommand{
			Image:      coloredBarImage,
			TargetType: resources.TargetScreen,
			Layer:      config.PipelineUILayer,
			GeoM:       currentGeoM,
		})
	}

	// Bar end cap
	endCapGeoM := ebiten.GeoM{}
	endCapGeoM.Concat(baseGeoM)
	endCapGeoM.Translate(float64(cfg.Hud.BarMiddleH)+max, float64(cfg.Hud.BarMiddleH)*y)

	queue.Push(resources.RenderCommand{
		Image:      barEndImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       endCapGeoM,
	})
}

// renderCount draws a numeric counter (heal or exp).
func renderCount(world *ecs.World, queue *resources.RenderQueue, baseGeoM ebiten.GeoM, y float64, count int, offset float64) {
	if world == nil || queue == nil {
		return
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	text := strconv.Itoa(count)

	// Measure text width using NanoFont (matches world-space debug text)
	w, _ := text2.Measure(text, fonts.NanoFont, 0)

	// Position for text
	textGeoM := ebiten.GeoM{}
	textGeoM.Concat(baseGeoM)
	textGeoM.Translate(float64(cfg.Hud.HudIconsX), float64(cfg.Hud.BarMiddleH)*y+offset)

	// Background bar
	bgGeoM := ebiten.GeoM{}
	bgGeoM.Scale(w+2, 1)
	bgGeoM.Concat(textGeoM)

	queue.Push(resources.RenderCommand{
		Image:      assets.HudFullCountBar,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       bgGeoM,
	})

	// Normal map punch-out
	queue.Push(resources.RenderCommand{
		Image:      assets.HudFullCountBar,
		TargetType: resources.TargetNormal,
		Layer:      config.PipelineUILayer,
		GeoM:       bgGeoM,
		ColorScale: components.NormalMaskColor,
	})

	// Create text image at logical size (not device pixels)
	textWidth := int(w) + 2
	textHeight := cfg.Hud.BarH + 2
	textImage := ebiten.NewImage(textWidth, textHeight)

	// Draw text at logical size using NanoFont
	textOpts := &text2.DrawOptions{}
	textOpts.GeoM.Translate(0, 2)
	textOpts.PrimaryAlign = text2.AlignStart
	text2.Draw(textImage, text, fonts.NanoFont, textOpts)

	// Position text (no scaling needed - already at logical size)
	textDrawGeoM := textGeoM
	textDrawGeoM.Translate(0, 2)

	queue.Push(resources.RenderCommand{
		Image:      textImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       textDrawGeoM,
		ColorScale: nil, // Use source color directly
	})
}

// renderAttackMult draws the attack multiplier indicator.
func renderAttackMult(world *ecs.World, queue *resources.RenderQueue, baseGeoM ebiten.GeoM, attackMult, maxHealth float64) {
	if world == nil || queue == nil || attackMult < minAttackMultToShow {
		return
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	endImgW := float64(barEndImage.Bounds().Dx())
	posGeoM := ebiten.GeoM{}
	posGeoM.Concat(baseGeoM)
	posGeoM.Translate(maxHealth+float64(cfg.Hud.BarMiddleH)+endImgW, 0)

	text := fmt.Sprintf("x%.1fATK", 1+attackMult)

	// Measure text width using NanoFont (matches world-space debug text)
	w, _ := text2.Measure(text, fonts.NanoFont, 0)

	// Background
	bgGeoM := ebiten.GeoM{}
	bgGeoM.Scale(w+2, 1)
	bgGeoM.Concat(posGeoM)

	queue.Push(resources.RenderCommand{
		Image:      assets.HudFullAttackBarImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       bgGeoM,
	})

	// Normal map punch-out
	queue.Push(resources.RenderCommand{
		Image:      assets.HudFullAttackBarImage,
		TargetType: resources.TargetNormal,
		Layer:      config.PipelineUILayer,
		GeoM:       bgGeoM,
		ColorScale: components.NormalMaskColor,
	})

	// Create text image at logical size (not device pixels)
	textWidth := int(w) + 2
	textHeight := cfg.Hud.BarH
	textImage := ebiten.NewImage(textWidth, textHeight)

	// Draw text at logical size using NanoFont
	textOpts := &text2.DrawOptions{}
	textOpts.GeoM.Translate(0, 1)
	textOpts.PrimaryAlign = text2.AlignStart
	text2.Draw(textImage, text, fonts.NanoFont, textOpts)

	// Position text (no scaling needed - already at logical size)
	textGeoM := posGeoM

	queue.Push(resources.RenderCommand{
		Image:      textImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       textGeoM,
	})
}
