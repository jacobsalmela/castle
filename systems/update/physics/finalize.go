package physics

import (
	"game/components"
	"game/ecs"
)

// FinalizeKinematics saves the current velocity to PrevVelocity for all physics entities.
// This system runs at the end of the physics phase after all movement and collision resolution.
//
// Pure ECS Migration Phase 3: Queries Physics component for velocity history.
// This ensures velocity history is preserved for systems that need to detect velocity changes.
//
// Phase: PHYSICS (Phase 4) - Finalization
// Order: Last in physics phase (after movement, collision resolution)
//
// Responsibilities:
//   - Save current velocity to PrevVelocity for next frame comparison
//   - Used by systems that need to detect velocity changes (impacts, damage, etc.)
//
// Used by: systems/update/tick/tick.go (PHASE 4: PHYSICS)
func FinalizeKinematics(world *ecs.World) {
	if world == nil {
		return
	}

	// Query all entities with Physics component
	entityList := world.EntitiesWith((*components.Physics)(nil))

	for _, eid := range entityList {
		physics := ecs.GetComponent[components.Physics](world, eid)
		if physics == nil {
			continue
		}

		// Save current velocity for next frame
		physics.PrevVelocity = physics.Velocity
	}
}
