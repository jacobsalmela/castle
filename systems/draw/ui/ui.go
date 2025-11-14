package ui

import (
	"game/components"
	"game/ecs"
	"game/resources"
	worldPackage "game/systems/draw/world"

	"github.com/hajimehoshi/ebiten/v2"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PHASE 4: UI RENDERING
// Render HUD, health bars, textboxes, and UI elements.
// ═══════════════════════════════════════════════════════════════════════════════

// Update renders the player HUD, enemy headbars, and textboxes after lighting.
// HUD is rendered separately because it needs to be drawn AFTER lighting is applied.
//
// This function:
//  1. Renders HUD to queue
//  2. Renders world-space UI elements (headbars, textboxes) to queue
//  3. Composes queue to screen (skips normal map - HUD doesn't need lighting)
//
// Parameters:
//   - world: ECS world instance
//   - screen: Logical screen buffer
func Update(world *ecs.World, screen *ebiten.Image) {
	if world == nil || screen == nil {
		return
	}

	queue := ecs.Resource[resources.RenderQueue](world)
	if queue == nil {
		panic("RenderQueue not available - this should not happen in Phase 14.5+")
	}

	// Get camera position for world-space UI elements
	cx, cy := 0.0, 0.0
	camera := ecs.Resource[resources.Camera](world)
	if camera != nil {
		cx, cy = camera.Position()
	}

	// Render HUD for player entity
	for _, eid := range world.EntitiesWith((*components.Player)(nil)) {
		RenderPlayerStatsHUD(world, queue, eid)
		break // Only one player
	}

	// Render enemy headbars
	renderHeadbars(world, queue, cx, cy)

	// Render textboxes
	renderTextboxes(world, queue)

	// Compose all UI commands to screen (skip normal map - UI doesn't need lighting)
	worldPackage.ComposeRenderQueue(screen, nil, queue)
}

// renderHeadbars renders enemy headbars using ECS components and RenderQueue.
func renderHeadbars(world *ecs.World, queue *resources.RenderQueue, cx, cy float64) {
	if world == nil || queue == nil {
		return
	}

	// Render headbars for each enemy type using components.
	// Each enemy type has migrated to Health/Poise/HeadHealthTimer components.
	// RenderEntityHeadbar handles the actual rendering logic for all entity types.
	for _, eid := range world.EntitiesWith((*components.Rat)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
	for _, eid := range world.EntitiesWith((*components.Bat)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
	for _, eid := range world.EntitiesWith((*components.Ghoul)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
	for _, eid := range world.EntitiesWith((*components.Skeleman)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
	for _, eid := range world.EntitiesWith((*components.Crawler)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
	for _, eid := range world.EntitiesWith((*components.Ent)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
	for _, eid := range world.EntitiesWith((*components.Knight)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
	for _, eid := range world.EntitiesWith((*components.Oscar)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
	for _, eid := range world.EntitiesWith((*components.Ferragus)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
	for _, eid := range world.EntitiesWith((*components.Gram)(nil), (*components.Transform)(nil)) {
		RenderEntityHeadbar(world, queue, eid, cx, cy)
	}
}

// renderTextboxes renders textboxes using ECS components and RenderQueue.
func renderTextboxes(world *ecs.World, queue *resources.RenderQueue) {
	if world == nil || queue == nil {
		return
	}

	camera := ecs.Resource[resources.Camera](world)
	if camera == nil {
		return
	}

	// Render Pure ECS TextboxData components
	for _, eid := range world.EntitiesWith((*components.TextboxData)(nil)) {
		textboxData := ecs.GetComponent[components.TextboxData](world, eid)
		if textboxData == nil {
			continue
		}

		// For entities (like graves), populate entity bounds from Transform if not set
		if textboxData.EntityX == 0 && textboxData.EntityY == 0 {
			if transform := ecs.GetComponent[components.Transform](world, eid); transform != nil {
				textboxData.EntityX = transform.X
				textboxData.EntityY = transform.Y
				textboxData.EntityW = transform.W
				textboxData.EntityH = transform.H
			}
		}

		// Render dismissed indicator ("i" above entity) if textbox is dismissed and not active
		if textboxData.Dismissed && !textboxData.Active {
			RenderDismissedIndicator(world, queue, textboxData, camera)
			continue // Skip rendering the textbox itself
		}

		// Skip inactive textboxes
		if !textboxData.Active {
			continue
		}

		// Prepare textbox data (compute lines, pagination) if not already prepared
		if textboxData.ProcessedText == "" {
			PrepareTextboxData(textboxData)
		}

		// Render via queue
		RenderTextbox(world, queue, textboxData, camera)
	}
}
