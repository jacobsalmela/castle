package physics

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
)

// RunPhysics advances basic physics for entities we know how to move.
// It runs after input/control systems and before combat/AI dependent on positions.
func RunPhysics(world *ecs.World, dt float64) {
	if world == nil || dt == 0 {
		return
	}

	// Get collision space from resources
	space := ecs.Resource[bump.Space](world)
	if space == nil {
		return
	}

	// For now, update a known set: Player and Rat components.
	// Their Body comps already contain the movement parameters/state.
	advanceBodies(world, space, dt)
}

// advanceBodies processes all entities with Physics and Body components,
func advanceBodies(world *ecs.World, space ecs.CollisionSpace, dt float64) {
	// STEP 1: Register all physics entities in collision space
	RegisterInSpace(world, space)

	// STEP 2: Apply forces (gravity, friction, velocity clamping)
	ApplyForces(world, dt)

	// STEP 3: Integrate movement with collision detection
	IntegrateMovement(world, space, dt)

	// STEP 4: Finalize kinematics (save PrevVelocity for next frame)
	FinalizeKinematics(world)

	// STEP 5: Collision resolution and grounding (currently inline in IntegrateMovement)
}

// shouldRemoveFromSpace checks if an entity should be removed from collision space (dead entities).
func shouldRemoveFromSpace(world *ecs.World, eid entities.EntityId) bool {
	// Check Health component
	if health := ecs.GetComponent[components.Health](world, eid); health != nil && health.Current <= 0 {
		return true
	}

	return false
}
