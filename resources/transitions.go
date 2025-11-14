package resources

import (
	"game/assets"
	"game/assets/fonts"
	"game/pkg/config"
	"game/pkg/tween"
	"game/pkg/utils"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// TransitionType indicates which type of transition is active.
type TransitionType int

const (
	TransitionNone TransitionType = iota
	TransitionRestart
	TransitionDeath
)

var (
	// Shared resources for transitions
	transitionTextImg *ebiten.Image
	grayscaleShader   *ebiten.Shader
)

func init() {
	// Initialize grayscale shader
	grayscaleShaderSrc := []byte(`
	package main
	var Force float
	func Fragment(position vec4, texCoord vec2, colorScale vec4) vec4 {
		color := imageSrc0At(texCoord)
		gray := 0.299 * color.r + 0.587 * color.g + 0.114 * color.b
		return vec4(mix(color.r, gray, Force), mix(color.g, gray, Force), mix(color.b, gray, Force), 1)
	}`)
	var err error
	if grayscaleShader, err = ebiten.NewShader(grayscaleShaderSrc); err != nil {
		log.Panic(err)
	}
}

// TransitionManager manages game transitions (restart, death) as an ECS resource.
type TransitionManager struct {
	activeType TransitionType

	// Restart transition state
	restartFadeTween *tween.Tween

	// Death transition state
	deathFreezeTime   float64
	deathFadeTween    *tween.Tween
	deathOverlayTween *tween.Tween
	deathOverlayImg   *ebiten.Image
	deathActionKeys   []ebiten.Key

	// Callbacks for transition completion
	onRestartComplete func()
	onDeathComplete   func()
}

// NewTransitionManager creates a new transition manager resource.
func NewTransitionManager() *TransitionManager {
	return &TransitionManager{
		activeType: TransitionNone,
	}
}

// IsActive returns true if any transition is currently active.
func (tm *TransitionManager) IsActive() bool {
	return tm.activeType != TransitionNone
}

// GetActiveType returns the currently active transition type.
func (tm *TransitionManager) GetActiveType() TransitionType {
	return tm.activeType
}

// StartRestart begins the restart transition with a fade-to-black effect.
func (tm *TransitionManager) StartRestart(onComplete func()) {
	if tm.activeType != TransitionNone {
		return // Don't interrupt existing transition
	}

	tm.activeType = TransitionRestart
	tm.restartFadeTween = tween.New(0, 1, 3, tween.EaseOutQuad)
	tm.onRestartComplete = onComplete
}

// StartDeath begins the death transition with grayscale effect and Game Over screen.
func (tm *TransitionManager) StartDeath(actionKeys []ebiten.Key, onComplete func(), cfg *config.Config) {
	if tm.activeType != TransitionNone {
		return // Don't interrupt existing transition
	}

	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	tm.activeType = TransitionDeath
	tm.deathFreezeTime = 0.5
	tm.deathFadeTween = tween.New(0, 1, 5, tween.EaseInQuad)
	tm.deathOverlayTween = tween.New(0, 1, 3, tween.EaseOutQuad)
	tm.deathActionKeys = actionKeys
	tm.onDeathComplete = onComplete

	// Prepare death overlay image with "Game Over" text
	if transitionTextImg == nil {
		transitionTextImg = ebiten.NewImage(int(cfg.Screen.Width), int(cfg.Screen.Height))
		textColor := color.RGBA{203, 219, 252, 255}

		// Draw "Game Over" title
		w, _ := utils.TextSize("Game Over", fonts.M6x11Font)
		titleOp := &ebiten.DrawImageOptions{}
		titleOp.GeoM.Translate(float64(cfg.Screen.Width-w)/2, 20)
		titleOp.ColorScale.ScaleWithColor(textColor)
		utils.DrawText(transitionTextImg, "Game Over", fonts.M6x11Font, titleOp)
	}

	// Create overlay with respawn instruction
	tm.deathOverlayImg, _ = transitionTextImg.SubImage(
		image.Rect(0, 0, int(cfg.Screen.Width), int(cfg.Screen.Height)),
	).(*ebiten.Image)

	text := "Press Attack Key to respawn"
	op := &ebiten.DrawImageOptions{}
	w, h := utils.TextSize(text, fonts.M5x7Font)
	op.GeoM.Translate(float64(cfg.Screen.Width-w)/2, cfg.Screen.Height-float64(h)-20)
	textColor := color.RGBA{203, 219, 252, 255}
	op.ColorScale.ScaleWithColor(textColor)
	utils.DrawText(tm.deathOverlayImg, text, fonts.M5x7Font, op)
}

// UpdateRestart updates the restart transition. Returns true when complete.
func (tm *TransitionManager) UpdateRestart(dt float64) bool {
	if tm.restartFadeTween == nil {
		return true
	}

	tm.restartFadeTween.Update(dt)
	if tm.restartFadeTween.IsDone() {
		if tm.onRestartComplete != nil {
			tm.onRestartComplete()
		}
		tm.activeType = TransitionNone
		tm.restartFadeTween = nil
		return true
	}

	return false
}

// UpdateDeath updates the death transition. Returns true when complete.
func (tm *TransitionManager) UpdateDeath(dt float64) bool {
	// Handle freeze time
	if tm.deathFreezeTime > 0 {
		tm.deathFreezeTime -= dt
		return false
	}

	// Update tweens
	if tm.deathFadeTween != nil {
		tm.deathFadeTween.Update(dt)
	}
	if tm.deathOverlayTween != nil {
		tm.deathOverlayTween.Update(dt)
	}

	// Check for respawn key press
	for _, key := range tm.deathActionKeys {
		if ebiten.IsKeyPressed(key) {
			if tm.onDeathComplete != nil {
				tm.onDeathComplete()
			}
			tm.activeType = TransitionNone
			tm.deathFadeTween = nil
			tm.deathOverlayTween = nil
			tm.deathOverlayImg = nil
			return true
		}
	}

	return false
}

// DrawRestart renders the restart transition fade effect.
func (tm *TransitionManager) DrawRestart(screen *ebiten.Image) {
	if tm.restartFadeTween == nil {
		return
	}

	alpha := float32(tm.restartFadeTween.Value())
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(assets.TransitionFadeImage, op)
}

// DrawDeath renders the death transition with grayscale shader and overlay.
func (tm *TransitionManager) DrawDeath(screen *ebiten.Image, cfg *config.Config) {
	if tm.deathOverlayTween == nil {
		return
	}

	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Apply grayscale shader
	overlayAlpha := float32(tm.deathOverlayTween.Value())
	newScreen := ebiten.NewImage(int(cfg.Screen.Width), int(cfg.Screen.Height))
	ops := &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{"Force": overlayAlpha},
		Images:   [4]*ebiten.Image{screen},
	}
	newScreen.DrawRectShader(int(cfg.Screen.Width), int(cfg.Screen.Height), grayscaleShader, ops)
	screen.DrawImage(newScreen, nil)

	// Draw fade overlay
	if tm.deathFadeTween != nil {
		alpha := float32(tm.deathFadeTween.Value())
		op := &ebiten.DrawImageOptions{}
		op.ColorScale.ScaleAlpha(alpha)
		screen.DrawImage(assets.TransitionFadeImage, op)
	}

	// Draw Game Over text overlay
	if tm.deathOverlayImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.ColorScale.ScaleAlpha(overlayAlpha)
		screen.DrawImage(tm.deathOverlayImg, op)
	}
}

// Draw renders the active transition if any.
func (tm *TransitionManager) Draw(screen *ebiten.Image, cfg *config.Config) {
	switch tm.activeType {
	case TransitionRestart:
		tm.DrawRestart(screen)
	case TransitionDeath:
		tm.DrawDeath(screen, cfg)
	}
}
