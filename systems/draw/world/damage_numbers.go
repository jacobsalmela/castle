package world

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"game/assets/fonts"
	"game/components"
	"game/ecs"
	"game/resources"
)

// DrawDamageNumbers renders floating damage number VFX.
// Numbers fade out as they drift away from their spawn point.
func DrawDamageNumbers(world *ecs.World, screen *ebiten.Image, camera *resources.Camera) {
	if world == nil || screen == nil || camera == nil {
		return
	}

	entities := world.EntitiesWith((*components.DamageNumber)(nil))
	for _, eid := range entities {
		dmg := ecs.GetComponent[components.DamageNumber](world, eid)
		transform := ecs.GetComponent[components.Transform](world, eid)

		if dmg == nil || transform == nil {
			continue
		}

		// Calculate fade (1.0 at start, 0.0 at end)
		prog := dmg.Tween.Value()
		alpha := 1.0 - prog

		// Get damage text
		damageText := formatDamageNumber(dmg.Damage)

		// Choose font size and color based on critical status
		var fontSize float64
		var baseColor color.RGBA

		if dmg.Critical {
			// Critical hit: Larger, red-orange text
			fontSize = 12.0
			baseColor = color.RGBA{255, 100, 50, uint8(alpha * 255)} // Red-orange
		} else {
			// Normal hit: Standard white text
			fontSize = 10.0
			baseColor = color.RGBA{255, 255, 255, uint8(alpha * 255)} // White
		}

		// Convert world position to screen position
		camX, camY := camera.Position()
		screenX := transform.X - camX
		screenY := transform.Y - camY

		// Draw text centered on position
		drawDamageText(screen, damageText, screenX, screenY, fontSize, baseColor)
	}
}

// formatDamageNumber formats damage for display.
func formatDamageNumber(damage float64) string {
	return formatFloat(damage, 0)
}

// drawDamageText renders text centered at the given position.
func drawDamageText(screen *ebiten.Image, txt string, x, y, size float64, clr color.RGBA) {
	// Get font face based on size
	var face text.Face
	if size >= 20 {
		// Critical hit: Use larger font
		face = fonts.RobotoMediumFontFace
	} else {
		// Normal hit: Use standard font
		face = fonts.NanoFont
	}

	if face == nil {
		return
	}

	// Measure text bounds for centering
	textWidth, textHeight := text.Measure(txt, face, 0)

	// Draw centered text
	op := &text.DrawOptions{}
	op.GeoM.Translate(-textWidth/2, -textHeight/2) // Center on position
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)

	text.Draw(screen, txt, face, op)
}

// formatFloat formats a float with specified decimal places.
// Simplified to use standard library for better performance and maintainability.
func formatFloat(val float64, decimals int) string {
	return strconv.Itoa(int(val))
}
