package ui

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/systems/update/physics"
)

const flickerTime = 0.5

// UpdateUITextbox updates all TextboxData components for proximity detection,
// flicker animation, and input handling.
func UpdateUITextbox(world *ecs.World, dt float64) {
	if world == nil || dt <= 0 {
		return
	}

	// Get player input component
	var playerInput *components.Input
	for _, pid := range world.EntitiesWith((*components.Player)(nil), (*components.Input)(nil)) {
		playerInput = ecs.GetComponent[components.Input](world, pid)
		break // Only one player
	}
	if playerInput == nil {
		return // Cannot process textbox input without player input
	}

	// Update TextboxData components
	for _, eid := range world.EntitiesWith((*components.TextboxData)(nil)) {
		textboxData := ecs.GetComponent[components.TextboxData](world, eid)
		if textboxData == nil {
			continue
		}
		updateTextboxLogic(world, eid, textboxData, dt, playerInput)
	}
}

// updateTextboxLogic handles proximity detection, flicker timer, and input processing.
func updateTextboxLogic(world *ecs.World, eid entities.EntityId, data *components.TextboxData, dt float64, input *components.Input) {
	if data == nil {
		return
	}

	playerNearby := isPlayerNearTextbox(world, eid, data)
	wasActive := data.Active

	if !playerNearby {
		handlePlayerLeftArea(data, wasActive)
		return
	}

	updateFlickerTimer(data, dt)
	handleTextboxInput(data, wasActive, input)
}

// isPlayerNearTextbox checks if the player is within the textbox activation area.
func isPlayerNearTextbox(world *ecs.World, eid entities.EntityId, data *components.TextboxData) bool {
	if world == nil || data == nil {
		return false
	}

	areaFunc := data.Area
	if areaFunc == nil {
		return false
	}

	// Query for player bodies in the textbox area
	rect := areaFunc()
	space := physics.GetCollisionSpace(world)
	collisions := physics.QueryItems(space, eid, rect, "body")

	for _, collidedEID := range collisions {
		// Check if entity has Player component (is the player)
		if team := ecs.GetComponent[components.Team](world, collidedEID); team != nil && team.Type == components.TeamPlayer {
			return true
		}
	}
	return false
}

// handlePlayerLeftArea deactivates the textbox when player leaves.
// The Dismissed flag is preserved so the textbox won't auto-show on next approach.
func handlePlayerLeftArea(data *components.TextboxData, wasActive bool) {
	if wasActive {
		data.AdvanceState = 0
	}
	data.Active = false
	// Keep Dismissed flag - player must press Up to reopen
}

// updateFlickerTimer updates the advance indicator flicker animation.
func updateFlickerTimer(data *components.TextboxData, dt float64) {
	data.AdvanceFlickerTimer += dt
	if data.AdvanceFlickerTimer > flickerTime {
		data.AdvanceFlickerTimer = 0
		data.AdvanceFlicker = !data.AdvanceFlicker
	}
}

// handleTextboxInput processes player input for textbox navigation.
func handleTextboxInput(data *components.TextboxData, wasActive bool, input *components.Input) {
	wasDismissed := data.Dismissed

	// Handle different states
	if handleDismissedState(data, wasActive, wasDismissed, input) {
		return
	}
	if handleFirstApproach(data, wasActive, wasDismissed) {
		return
	}
	handleActiveNavigation(data, wasActive, input)
}

// handleDismissedState processes reopening a dismissed textbox.
func handleDismissedState(data *components.TextboxData, wasActive, wasDismissed bool, input *components.Input) bool {
	if !wasActive && wasDismissed {
		if input.KeyPressed[components.InputKeyUp] {
			data.AdvanceState = 0
			data.Active = true
			data.Dismissed = false
		}
		return true
	}
	return false
}

// handleFirstApproach auto-activates textbox on first player approach.
func handleFirstApproach(data *components.TextboxData, wasActive, wasDismissed bool) bool {
	if !wasActive && !wasDismissed {
		data.Active = true
		return true
	}
	return false
}

// handleActiveNavigation processes navigation while textbox is active.
func handleActiveNavigation(data *components.TextboxData, wasActive bool, input *components.Input) {
	// Down key: advance to next page or close
	if input.KeyPressed[components.InputKeyDown] {
		if data.AdvanceState < data.AdvanceMax {
			// Multi-page textbox: advance to next page
			data.AdvanceState++
		} else {
			// Already at last page (or single-page textbox with maxState=0): dismiss
			data.Active = false
			data.Dismissed = true
			return // Exit early to prevent reactivation below
		}
	}

	// Up key: go back to previous page (only if not on first page)
	if input.KeyPressed[components.InputKeyUp] && data.AdvanceState > 0 {
		data.AdvanceState--
	}

	// Keep active if it was active and not dismissed
	if wasActive {
		data.Active = true
	}
}

// UpdateTextboxText assigns dialogue text to an entity with TextboxData.
// This is the unified text setter function - use instead of prefab-specific setters.
// Resets pagination state to show the updated text from the beginning.
func UpdateTextboxText(world *ecs.World, eid entities.EntityId, text string) {
	if world == nil {
		return
	}

	textboxData := ecs.GetComponent[components.TextboxData](world, eid)
	if textboxData == nil {
		return
	}

	textboxData.Text = text
	textboxData.AdvanceState = 0
	textboxData.AdvanceFlickerTimer = 0
	textboxData.AdvanceFlicker = false
	// Clear processed text to trigger reprocessing on next draw
	textboxData.ProcessedText = ""
}
