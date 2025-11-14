package physics

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/pkg/config"
)

// IntegrateMovement performs movement integration using the collision space.
// This system calls space.Move() to update entity positions while detecting collisions.
//
// Pure ECS Migration Complete: NOW queries Physics + Collider components exclusively.
// Logic moved from component to system, following Pure ECS pattern.
//
// Phase: PHYSICS (Phase 4) - Movement
// Order: Third in physics phase (after forces, before collision resolution)
//
// Responsibilities:
//   - Calculate goal position from current position + velocity * dt
//   - Create movement filter for collision detection using Collider
//   - Call space.Move() to integrate movement with collision detection
//   - Update Transform component with final position
//   - Collect collision results for resolution phase
//   - Record collisions for debug visualization
//
// Used by: systems/update/tick/tick.go (PHASE 4: PHYSICS)
func IntegrateMovement(world *ecs.World, space ecs.CollisionSpace, dt float64) {
	if world == nil || dt == 0 {
		return
	}

	// Query all entities with Physics component
	entityList := world.EntitiesWith((*components.Physics)(nil))

	for _, eid := range entityList {
		physics := ecs.GetComponent[components.Physics](world, eid)
		if physics == nil {
			continue
		}

		// Skip very fresh projectiles (SpawnGrace > 0) to allow Init to complete
		if IsProjectileSpawning(world, eid) {
			continue
		}

		transform := ecs.GetComponent[components.Transform](world, eid)
		if transform == nil {
			continue
		}

		// Skip truly static physics (no gravity/friction AND no velocity)
		// Entities with velocity but no gravity/friction (like flakes) still need movement
		if !physics.GravityEnabled && !physics.FrictionEnabled && physics.Velocity.X == 0 && physics.Velocity.Y == 0 {
			continue
		}

		// If no collision space, do simple position integration
		if space == nil {
			integrateWithoutCollisionPhysics(transform, physics, dt)
			continue
		}

		// Get Collider component for collision detection
		collider := ecs.GetComponent[components.Collider](world, eid)
		if collider == nil {
			// No collider = no collision detection, just move
			integrateWithoutCollisionPhysics(transform, physics, dt)
			continue
		}

		// Integrate with collision detection
		integrateWithCollisionPhysics(world, physics, collider, eid, transform, space, dt)
	}
}

// integrateWithoutCollisionPhysics performs simple position integration without collision detection.
// Phase 3: New function for Physics component.
func integrateWithoutCollisionPhysics(transform *components.Transform, physics *components.Physics, dt float64) {
	if transform == nil || physics == nil {
		return
	}

	// Simple Euler integration: position += velocity * dt
	transform.X += physics.Velocity.X * dt
	transform.Y += physics.Velocity.Y * dt

	// Not on ground when no collision detection
	physics.Grounded = false
	physics.OnPlatform = false
}

// integrateWithCollisionPhysics performs movement integration with collision detection.
// Phase 3: New function for Physics + Collider components.
func integrateWithCollisionPhysics(world *ecs.World, physics *components.Physics, collider *components.Collider, entityID entities.EntityId, transform *components.Transform, space ecs.CollisionSpace, dt float64) {
	if entityID == 0 {
		integrateWithoutCollisionPhysics(transform, physics, dt)
		return
	}

	// Get current position as bump.Rect (with collider offset)
	rect := collider.ToBumpRect(transform)

	// Calculate goal position: current + velocity * dt
	goal := bump.Vec2{
		X: rect.X + physics.Velocity.X*dt,
		Y: rect.Y + physics.Velocity.Y*dt,
	}

	// Prepare collision query tags from Collider (what to collide AGAINST)
	queryTags := collider.ToQueryTags()
	if len(queryTags) == 0 {
		queryTags = []bump.Tag{"body", "map", "solid"}
	}

	// Create movement filter (handles FilterOut, pass-through, etc.)
	filter := createMovementFilterForPhysics(physics, collider, space)

	// Perform collision-aware movement
	finalPos, collisions := space.Move(entityID, goal, filter, queryTags...)

	// Update transform with final position (accounting for collider offset)
	transform.X = finalPos.X - collider.OffsetX
	transform.Y = finalPos.Y - collider.OffsetY

	// Record collisions for debug visualization
	recordCollisionsForPhysics(world, transform, collisions, finalPos.X, finalPos.Y, collider)

	// Resolve collisions (velocity cancellation, grounding state)
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	resolveCollisionsForPhysics(world, entityID, physics, transform, space, collisions, dt, &cfg.Body)
}

// createMovementFilterForPhysics creates a collision filter for physics-based movement.
// Handles FilterOut entities, pass-through platforms, and other collision rules.
func createMovementFilterForPhysics(physics *components.Physics, collider *components.Collider, space ecs.CollisionSpace) bump.Filter {
	filterOut := collider.FilterOut
	dropping := physics.DroppingThrough

	return func(item, other bump.Item) (bump.ColType, bool) {
		if space == nil {
			return bump.Slide, false
		}

		// Handle entity-entity collisions
		if entityID, ok := other.(entities.EntityId); ok {
			if shouldIgnoreEntity(entityID, filterOut) {
				return 0, true // Ignore this collision
			}
			return getEntityCollisionType(entityID, space)
		}

		// Handle platform collisions
		return getPlatformCollisionType(item, other, space, dropping)
	}
}

// recordCollisionsForPhysics records collision events for debug visualization.
// Phase 3: New function for Physics component.
//
// TODO(debug): Re-enable collision recording when debug system supports it.
// Collision visualization will be added in Phase 2 of debug system refactor.
func recordCollisionsForPhysics(world *ecs.World, transform *components.Transform, cols []*bump.Collision, finalX, finalY float64, collider *components.Collider) {
	_ = world
	_ = transform
	_ = cols
	_ = finalX
	_ = finalY
	_ = collider
}

// resolveCollisionsForPhysics handles collision response for Physics component.
// Phase 3: New function replacing resolveCollisions for BodyKinematics.
func resolveCollisionsForPhysics(world *ecs.World, entityID entities.EntityId, physics *components.Physics, transform *components.Transform, space ecs.CollisionSpace, cols []*bump.Collision, dt float64, cfg *config.Body) {
	// Step 1: Cancel velocities from collisions
	cancelVelocityFromCollisions(physics, cols)

	// Step 2: Update grounding state
	grounded, onPlatform := updateGroundingFromCollisions(cols, space)
	physics.Grounded = grounded
	physics.OnPlatform = onPlatform

	// Step 3: Reset drop-through if not on platform
	if !onPlatform {
		physics.DroppingThrough = false
	}

	// Step 4: Update coyote time (jump grace period)
	updateCoyoteTime(physics, grounded, dt, cfg)
}
