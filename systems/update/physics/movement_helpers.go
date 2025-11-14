package physics

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
)

// calculateMovementGoal computes the target position for physics movement.
func calculateMovementGoal(physics *components.Physics, collider *components.Collider,
	transform *components.Transform, dt float64) (bump.Rect, bump.Vec2) {
	rect := collider.ToBumpRect(transform)
	goal := bump.Vec2{
		X: rect.X + physics.Velocity.X*dt,
		Y: rect.Y + physics.Velocity.Y*dt,
	}
	return rect, goal
}

// executeCollisionMove performs collision-aware movement in the spatial hash.
func executeCollisionMove(entityID entities.EntityId, goal bump.Vec2,
	collider *components.Collider, physics *components.Physics,
	space ecs.CollisionSpace) (bump.Vec2, []*bump.Collision) {
	queryTags := collider.ToQueryTags()
	if len(queryTags) == 0 {
		queryTags = []bump.Tag{"body", "map", "solid"}
	}

	filter := createMovementFilterForPhysics(physics, collider, space)
	return space.Move(entityID, goal, filter, queryTags...)
}

// applyMovementResults updates transform and resolves collisions.
func applyMovementResults(world *ecs.World, entityID entities.EntityId,
	transform *components.Transform, physics *components.Physics,
	collider *components.Collider, finalPos bump.Vec2,
	collisions []*bump.Collision, space ecs.CollisionSpace, dt float64) {
	// Update transform with collision offset
	transform.X = finalPos.X - collider.OffsetX
	transform.Y = finalPos.Y - collider.OffsetY

	// Record collisions for debug visualization
	recordCollisionsForPhysics(world, transform, collisions, finalPos.X, finalPos.Y, collider)

	// Resolve velocity and grounding
	cfg := getConfig(world)
	resolveCollisionsForPhysics(world, entityID, physics, transform, space, collisions, dt, &cfg.Body)
}
