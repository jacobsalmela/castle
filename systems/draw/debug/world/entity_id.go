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

	"github.com/hajimehoshi/ebiten/v2"
)

// DrawEntityIDDebug is a placeholder for entity ID visualization.
// Entity ID text rendering is handled by DrawEntityIDText() called from sprites.go.
func DrawEntityIDDebug(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryEntityID) {
		return
	}

	// Entity ID text is already rendered in sprites.go via DrawEntityIDText()
	// This is a placeholder for any additional entity ID visualization
	_ = world
	_ = screen
	_ = cam
}

// DrawEntityIDText renders the entity ID centered on the entity.
// Uses NanoFont (6pt) for proper world-space scaling.
func DrawEntityIDText(screen *ebiten.Image, eid entities.EntityId, transform *components.Transform, entityPos ebiten.GeoM) {
	if transform == nil {
		return
	}

	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	// Center the text on the entity
	op.GeoM.Translate(transform.W/2-8, transform.H/2-3)

	// Use NanoFont for world-space rendering (entityPos is supplied by caller).
	// Keep existing behavior: caller is responsible for constructing entityPos
	utils.DrawText(screen, fmt.Sprintf("%d", eid), fonts.NanoFont, op)
}
