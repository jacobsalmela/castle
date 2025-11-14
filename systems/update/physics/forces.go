package physics

import (
	"game/components"
	"game/ecs"
	"game/pkg/config"
	"math"
)

// ApplyForces applies physics forces (gravity, friction) to all entities with Physics component.
// This system runs after registration and before movement integration.
//
// Pure ECS Migration Phase 3: NOW queries Physics component instead of BodyKinematics.
// Logic moved from component to system, following Pure ECS pattern.
//
// Phase: PHYSICS (Phase 4) - Forces
// Order: Second in physics phase (after registration, before movement)
//
// Responsibilities:
//   - Apply gravity to vertical velocity based on weight (if GravityEnabled)
//   - Apply friction (ground/air) to horizontal velocity (if FrictionEnabled)
//   - Clamp velocities to max limits (MaxVelocity)
//   - Zero out velocities below friction epsilon
//
// Used by: systems/update/tick/tick.go (PHASE 4: PHYSICS)
func ApplyForces(world *ecs.World, dt float64) {
	if world == nil || dt == 0 {
		return
	}

	// Get config from world resource
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
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

		// Skip static physics (entities without gravity or friction enabled)
		if !physics.GravityEnabled && !physics.FrictionEnabled {
			continue
		}

		// Apply horizontal and vertical forces using Physics component
		applyHorizontalForcesToPhysics(physics, &cfg.Body, dt)
		applyVerticalForcesToPhysics(physics, &cfg.Body, dt)
	}
}

// applyHorizontalForcesToPhysics applies friction and velocity clamping to horizontal movement.
// Phase 3: New function for Physics component (replaces applyHorizontalForces for BodyKinematics).
func applyHorizontalForcesToPhysics(physics *components.Physics, cfg *config.Body, dt float64) {
	if physics == nil || cfg == nil || dt == 0 {
		return
	}

	maxX := physics.MaxVelocity.X
	if maxX == 0 {
		maxX = cfg.MaxX
	}

	// Check if friction should be applied
	noForce := !physics.Grounded || physics.PrevVelocity.X == physics.Velocity.X
	if (physics.FrictionEnabled && noForce) || math.Abs(physics.Velocity.X) > maxX {
		// Choose friction coefficient based on grounding state
		fric := cfg.GroundFriction
		if !physics.Grounded {
			fric = cfg.AirFriction
		}

		// Apply friction
		physics.Velocity.X -= physics.Velocity.X * fric * dt

		// Zero out very small velocities
		if math.Abs(physics.Velocity.X) < cfg.FrictionEpsilon {
			physics.Velocity.X = 0
		}
	}
}

// applyVerticalForcesToPhysics applies gravity and velocity clamping to vertical movement.
// Phase 3: New function for Physics component (replaces applyVerticalForces for BodyKinematics).
func applyVerticalForcesToPhysics(physics *components.Physics, cfg *config.Body, dt float64) {
	if physics == nil || cfg == nil || dt == 0 {
		return
	}

	// Apply gravity if enabled
	if physics.GravityEnabled {
		physics.Velocity.Y += cfg.Gravity * physics.Weight * dt
	}

	// Clamp vertical velocity
	maxY := physics.MaxVelocity.Y
	if maxY == 0 {
		maxY = cfg.MaxY
	}

	if physics.Velocity.Y > maxY {
		physics.Velocity.Y = maxY
		return
	}
	if physics.Velocity.Y < -maxY {
		physics.Velocity.Y = -maxY
	}
}
