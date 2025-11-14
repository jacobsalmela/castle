package ui

import (
	"game/ecs"
)

// Update handles Phase 9: UI updates including textboxes and interactive UI elements.
func Update(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	// Update textbox proximity detection and interactions
	UpdateUITextbox(world, dt)
}
