package vfx

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

// UpdateDebris advances debris VFX particles (destruction particles).
// Handles rotation, grounding friction, and lifetime cleanup.
func UpdateDebris(world *ecs.World, _ interface{}, dt float64) {
	if world == nil || dt == 0 {
		return
	}

	for _, eid := range world.EntitiesWith((*components.Debris)(nil)) {
		debris := ecs.GetComponent[components.Debris](world, eid)
		if debris == nil {
			continue
		}

		updateDebris(world, eid, debris, dt)
	}
}

func updateDebris(world *ecs.World, eid entities.EntityId, debris *components.Debris, dt float64) {
	// Get components
	render := ecs.GetComponent[components.Render](world, eid)
	physics := ecs.GetComponent[components.Physics](world, eid)
	if render == nil {
		return
	}

	// Apply rotation decay when grounded
	if physics != nil && physics.Grounded {
		debris.RotationSpeed *= 0.98 * dt
	}

	// Update rotation
	render.R += debris.RotationSpeed * dt

	// Decrement lifetime timer
	debris.Timer -= dt
	if debris.Timer <= 0 {
		// Cleanup: destroy the debris entity
		world.DestroyEntity(eid)
	}
}
