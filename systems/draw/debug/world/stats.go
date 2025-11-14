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

// DrawStatsDebug is a placeholder for stats visualization.
// Stats text rendering is handled by DrawStatsDebugText() called from sprites.go.
func DrawStatsDebug(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryStats) {
		return
	}

	// Stats debug text is already rendered in sprites.go via DrawStatsDebugText()
	// This is a placeholder for any additional stats visualization
	_ = world
	_ = screen
	_ = cam
}

// DrawStatsDebugText renders health/stamina/poise above an entity.
// This is the Pure ECS version that uses separate Health, Stamina, Poise components.
// Uses NanoFont (6pt) for proper world-space scaling.
func DrawStatsDebugText(world *ecs.World, screen *ebiten.Image, eid entities.EntityId, entityPos ebiten.GeoM) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryStats) {
		return
	}

	health := ecs.GetComponent[components.Health](world, eid)
	stamina := ecs.GetComponent[components.Stamina](world, eid)
	poise := ecs.GetComponent[components.Poise](world, eid)

	// Only render if entity has at least one stat component
	if health == nil && stamina == nil && poise == nil {
		return
	}

	// Build text string with available stats
	text := ""
	if health != nil {
		text += fmt.Sprintf("H:%.1f/%.1f\n", health.Current, health.Max)
	}
	if stamina != nil {
		if text != "" {
			text += " "
		}
		text += fmt.Sprintf("S:%.1f/%.1f\n", stamina.Current, stamina.Max)
	}
	if poise != nil {
		if text != "" {
			text += " "
		}
		text += fmt.Sprintf("P:%.1f/%.1f\n", poise.Current, poise.Max)
	}

	// Get DPR from viewport resource so we position DPR-scaled fonts correctly
	dpr := 1.0
	if vp := ecs.Resource[components.ViewPort](world); vp != nil {
		dpr = vp.DPR
	}

	// Position text above entity (offsets scaled by DPR)
	op := &ebiten.DrawImageOptions{GeoM: entityPos}
	op.GeoM.Translate(-5*dpr, -16*dpr)

	// Use NanoFont for world-space rendering. NanoFont is created with DPR-scaled size,
	// so translate positions into device pixels by scaling offsets with DPR.
	utils.DrawText(screen, text, fonts.NanoFont, op)
}
