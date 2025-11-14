package physics

import (
	"game/components"
	"game/ecs"
	"game/pkg/bump"
	"game/pkg/config"
)

// updateGroundingFromCollisions determines grounding state from collision results.
func updateGroundingFromCollisions(collisions []*bump.Collision,
	space ecs.CollisionSpace) (grounded, onPlatform bool) {
	for _, col := range collisions {
		if col == nil {
			continue
		}

		// Check for ground collision (normal points up)
		if col.Type == bump.Slide && col.Normal.Y < 0 {
			grounded = true
		}

		// Check for platform
		if space.Has(col.Other, bump.Tag("passthrough")) && col.Overlaps {
			onPlatform = true
		}
	}
	return grounded, onPlatform
}

// cancelVelocityFromCollisions stops velocity in collision directions.
func cancelVelocityFromCollisions(physics *components.Physics, collisions []*bump.Collision) {
	for _, col := range collisions {
		if col == nil || col.Type != bump.Slide {
			continue
		}

		// Cancel horizontal velocity on wall collision
		if col.Normal.X != 0 {
			physics.Velocity.X = 0
		}

		// Cancel vertical velocity on floor/ceiling collision
		if col.Normal.Y < 0 || (col.Normal.Y > 0 && physics.Velocity.Y < 0) {
			physics.Velocity.Y = 0
		}
	}
}

// updateCoyoteTime manages jump grace period after leaving ground.
func updateCoyoteTime(physics *components.Physics, grounded bool, dt float64, cfg *config.Body) {
	if grounded {
		physics.AirTime = 0
		physics.CanJump = true
	} else {
		physics.AirTime += dt
		if physics.AirTime > cfg.CoyoteTimeSeconds {
			physics.CanJump = false
		}
	}
}
