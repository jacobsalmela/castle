package physics

import (
	"game/ecs"
)

// UpdateGroundingState is a placeholder for future grounding state decoupling.
//
// Currently, grounding state resolution is handled inline during collision resolution
// in collision_resolution.go (updateGroundingFromCollisions, updateCoyoteTime).
//
// This system exists for architectural completeness and future refactoring.
// It may be used in the future to decouple grounding state updates from collision resolution,
// enabling independent grounding queries for AI/decision systems.
//
// TODO(physics): Re-enable this system when grounding state is decoupled from collision resolution.
// This will be added in Phase 2 of physics system architecture improvements.
func UpdateGroundingState(world *ecs.World, space ecs.CollisionSpace, dt float64) {
	// Grounding logic currently handled in:
	// - collision_resolution.go: updateGroundingFromCollisions()
	// - collision_resolution.go: updateCoyoteTime()
	_ = world
	_ = space
	_ = dt
}
