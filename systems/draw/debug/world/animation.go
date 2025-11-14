//go:build !release

package world

import (
	"fmt"
	"game/assets/fonts"
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/utils"
	"game/resources"
	"game/systems/draw/debug/primitives"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// DrawAnimDebug renders animation debug information.
// Enable with Cmd+6. Shows:
// - White text: Current animation state and frame number above each entity
// Uses NanoFont (6pt) for proper world-space scaling.
func DrawAnimDebug(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryAnim) {
		return
	}

	coords := primitives.NewWorldToScreen(cam)

	animEntities := world.EntitiesWith((*components.Animation)(nil), (*components.Transform)(nil))
	for _, eid := range animEntities {
		anim := ecs.GetComponent[components.Animation](world, eid)
		transform := ecs.GetComponent[components.Transform](world, eid)

		if anim == nil || transform == nil {
			continue
		}

		// Position text above entity (convert to device pixels using DPR)
		dpr := 1.0
		if vp := ecs.Resource[components.ViewPort](world); vp != nil {
			dpr = vp.DPR
		}

		textX, textY := coords.TransformF64(transform.X, transform.Y-20) // Above entity
		textX *= dpr
		textY *= dpr

		// Format: "StateName (Frame)"
		frameInfo := fmt.Sprintf("%s (%d)", anim.State, anim.Frame)

		// Create draw options
		opts := &text.DrawOptions{}
		opts.GeoM.Translate(textX, textY)
		opts.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})

		// Use NanoFont (created with DPR-scaled size). Positions above have been
		// converted to device pixels so the text follows the entity correctly.
		text.Draw(screen, frameInfo, fonts.NanoFont, opts)
	}
}

// DrawPlayerFrameDebug renders current frame info for the player sprite.
// This is ALWAYS enabled for debugging sprite rendering issues.
func DrawPlayerFrameDebug(world *ecs.World, screen *ebiten.Image, eid entities.EntityId, entityPos ebiten.GeoM) {
	anim := ecs.GetComponent[components.Animation](world, eid)
	render := ecs.GetComponent[components.Render](world, eid)

	if anim == nil || anim.Data == nil {
		return
	}

	// Get DPR so offsets line up with DPR-scaled font
	dpr := 1.0
	if vp := ecs.Resource[components.ViewPort](world); vp != nil {
		dpr = vp.DPR
	}

	// Position text above entity (scale offset by DPR)
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(0, -10*dpr)

	// Bright yellow color for high visibility
	op.ColorScale.ScaleWithColor(color.RGBA{R: 255, G: 255, B: 0, A: 255})

	// Format: "State Frame# [ImageOK]"
	imageStatus := "NIL"
	if render != nil && render.Image != nil {
		imageStatus = "OK"
	}

	text := fmt.Sprintf("%s F%d [%s]", anim.State, anim.Data.CurrentFrame, imageStatus)

	utils.DrawText(screen, text, fonts.NanoFont, op)
}
