package entities

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/resources"
)

// UpdateGrave processes grave entities for save/rest functionality.
//
// Graves are interactive save points. When the player approaches a grave,
// the textbox system displays an interaction prompt. If the player presses
// Up while the textbox is active (first page), it triggers a save and reset.
//
// This system coordinates between the Grave, Textbox, and Transform components
// to handle proximity-based interactions and global save/reset flags.
func UpdateGrave(world *ecs.World, _ float64) {
	if world == nil {
		return
	}

	// Get player input component
	var playerInput *components.Input
	for _, pid := range world.EntitiesWith((*components.Player)(nil), (*components.Input)(nil)) {
		playerInput = ecs.GetComponent[components.Input](world, pid)
		break // Only one player
	}
	if playerInput == nil {
		return // Cannot process grave interactions without player input
	}

	graves := world.EntitiesWith((*components.Grave)(nil), (*components.TextboxData)(nil))

	for _, eid := range graves {
		processGrave(world, eid, playerInput)
	}
}

// processGrave handles a single grave entity's save/rest logic.
func processGrave(world *ecs.World, eid entities.EntityId, input *components.Input) {
	// Get required components
	textbox := ecs.GetComponent[components.TextboxData](world, eid)
	if textbox == nil {
		// debug.Log("Grave", "Grave %d missing valid textbox", eid)
		return
	}

	// Debug visualization (optional)
	debugGraveVisuals(world, eid)

	// Check for rest activation
	if isRestActivated(textbox, input) {
		activateGraveRest(world, eid, textbox)
	}
}

// debugGraveVisuals draws debug info for grave entities.
func debugGraveVisuals(world *ecs.World, eid entities.EntityId) {
	// Debug visualization (now handled by Pure ECS debug_grave.go system)
	// Logging functionality removed - use Pure ECS debug categories
	_ = world
	_ = eid
}

// isRestActivated checks if the player activated the grave's rest function.
// Returns true when:
//   - Textbox is active (player is near the grave)
//   - Textbox is on first page (AdvanceState == 0)
//   - Player pressed Up key
func isRestActivated(textbox *components.TextboxData, input *components.Input) bool {
	if textbox == nil {
		return false
	}
	return textbox.Active &&
		textbox.AdvanceState == 0 &&
		input.KeyPressed[components.InputKeyUp]
}

// activateGraveRest triggers the save and reset sequence.
// Uses GameSignals resource to request save and reset.
func activateGraveRest(world *ecs.World, eid entities.EntityId, textbox *components.TextboxData) {
	// debug.Log("Grave", "Grave %d: Activating save and reset", eid)
	_ = eid

	// Request save and reset via GameSignals resource
	signals := ecs.Resource[resources.GameSignals](world)
	if signals != nil {
		signals.RequestSave()
		signals.RequestReset()
	}

	// Dismiss and deactivate textbox
	if textbox != nil {
		textbox.Dismissed = true
		textbox.Active = false
	}
}
