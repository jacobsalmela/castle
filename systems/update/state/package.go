package state

import (
	"game/ecs"
)

// Update processes global state updates such as death timers and stamina regeneration.
func Update(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	// Death timer countdown and fade animation
	UpdateDeathState(world, dt)

	// Stamina regeneration for all entities
	UpdateStamina(world, dt)
}
