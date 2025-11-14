package state

import (
	"game/components"
	"game/ecs"
)

// UpdateStamina handles stamina regeneration for all entities with Stamina components.
// Stamina regenerates at RecoveryRate per second when not at maximum.
func UpdateStamina(world *ecs.World, dt float64) {
	if world == nil || dt <= 0 {
		return
	}

	// Update all Stamina components
	for _, eid := range world.EntitiesWith((*components.Stamina)(nil)) {
		stamina := ecs.GetComponent[components.Stamina](world, eid)
		if stamina == nil {
			continue
		}

		// Regenerate stamina if not at maximum
		if stamina.Current < stamina.Max && stamina.RecoveryRate > 0 {
			stamina.Current += stamina.RecoveryRate * dt

			// Clamp to maximum
			if stamina.Current > stamina.Max {
				stamina.Current = stamina.Max
			}
		}
	}
}
