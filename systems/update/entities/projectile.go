package entities

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/systems/update/combat"
	"game/systems/update/physics"
)

// UpdateProjectiles advances projectile behavior and schedules removal when needed.
// Pure ECS implementation - projectiles are pure data components.
func UpdateProjectiles(world *ecs.World, _ interface{}, dt float64) {
	if world == nil || dt == 0 {
		return
	}
	for _, eid := range world.EntitiesWith((*components.Projectile)(nil)) {
		p := ecs.GetComponent[components.Projectile](world, eid)
		if skipProjectile(world, eid, p) {
			continue
		}

		// Handle bouncing state
		if p.Bouncing {
			updateBouncingProjectile(world, eid, p, dt)
			continue
		}

		// Disable friction for active projectiles
		physics := ecs.GetComponent[components.Physics](world, eid)
		if physics != nil {
			physics.FrictionEnabled = false
		}

		if handleProjectileSpawnGrace(p) {
			continue
		}

		// Check for hit
		if projectileHit(world, eid, p) {
			// Check if we should bounce instead of destroy
			if shouldBounce(world, eid, p) {
				startBouncing(world, eid, p)
			} else {
				cleanupProjectile(world, eid)
			}
		}
	}
}

func skipProjectile(world *ecs.World, eid entities.EntityId, p *components.Projectile) bool {
	if p == nil {
		return true
	}

	// Bouncing projectiles don't need hitbox (it's removed during bounce)
	if p.Bouncing {
		// Only need Physics for bouncing physics
		if !ecs.HasComponent[components.Physics](world, eid) {
			return true
		}
		return false
	}

	// Active (non-bouncing) projectiles need both hitbox and physics
	if !ecs.HasComponent[components.Hitbox](world, eid) {
		return true
	}
	if !ecs.HasComponent[components.Physics](world, eid) {
		return true
	}
	return false
}

func handleProjectileSpawnGrace(p *components.Projectile) bool {
	if p == nil || p.SpawnGrace <= 0 {
		return false
	}
	p.SpawnGrace--
	return true
}

func projectileHit(world *ecs.World, eid entities.EntityId, p *components.Projectile) bool {
	if world == nil || p == nil {
		return false
	}
	t := ecs.GetComponent[components.Transform](world, eid)
	if t == nil {
		return false
	}
	filter := buildProjectileFilter(world, eid, p)
	rect := bump.Rect{W: t.W, H: t.H}
	contactType, contacted := combat.ResolveHitboxArea(world, eid, rect, p.Damage, filter)

	// Check for landing or collision
	grounded, vy := false, 0.0
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics != nil {
		grounded = physics.Grounded
		vy = physics.Velocity.Y
	}
	return len(contacted) > 0 || contactType >= components.Block || (grounded && vy >= 0)
}

// buildProjectileFilter creates a filter to prevent projectiles from hitting their owner.
// Pure ECS: Queries owner's hitbox component directly via EntityId.
func buildProjectileFilter(world *ecs.World, eid entities.EntityId, p *components.Projectile) []*components.Hitbox {
	if world == nil || p == nil || p.Owner == 0 {
		return nil
	}
	// Get owner's hitbox to filter it out
	ownerHitbox := ecs.GetComponent[components.Hitbox](world, p.Owner)
	if ownerHitbox == nil {
		return nil
	}
	return []*components.Hitbox{ownerHitbox}
}

func cleanupProjectile(world *ecs.World, eid entities.EntityId) {
	// Remove from collision space
	if space := physics.GetCollisionSpace(world); space != nil {
		space.Remove(eid)
	}

	// Pure ECS: Components removed automatically when entity destroyed
	// No manual cleanup needed for pure data components

	world.DestroyEntity(eid)
}

// shouldBounce determines if projectile should bounce instead of being destroyed.
// Returns true if projectile hit ground/wall but not an enemy.
func shouldBounce(world *ecs.World, eid entities.EntityId, p *components.Projectile) bool {
	if world == nil || p == nil {
		return false
	}

	// Check what we hit
	t := ecs.GetComponent[components.Transform](world, eid)
	if t == nil {
		return false
	}

	filter := buildProjectileFilter(world, eid, p)
	rect := bump.Rect{W: t.W, H: t.H}
	contactType, contacted := combat.ResolveHitboxArea(world, eid, rect, p.Damage, filter)

	// If we hit an enemy/damageable target, don't bounce (destroy)
	if len(contacted) > 0 {
		return false
	}

	// If we hit ground or wall, bounce
	if contactType >= components.Block {
		return true
	}

	// If we're grounded, bounce
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics != nil {
		if physics.Grounded && physics.Velocity.Y >= 0 {
			return true
		}
	}

	return false
}

// startBouncing transitions projectile into bouncing state.
// Reduces velocity, enables friction, removes hitbox, sets timer.
func startBouncing(world *ecs.World, eid entities.EntityId, p *components.Projectile) {
	if world == nil || p == nil {
		return
	}

	// Mark as bouncing
	p.Bouncing = true
	p.BounceTimer = 3.0 // 3 seconds to roll and fade
	p.Damage = 0        // Can't damage while bouncing

	// Get transform and render for visual adjustment
	t := ecs.GetComponent[components.Transform](world, eid)
	render := ecs.GetComponent[components.Render](world, eid)

	if t != nil {
		// Lower the transform Y by 3 pixels to compensate for visual offset
		// This makes the rock sit properly on the ground instead of floating
		t.Y += 3
	}

	// Remove render Y offset so rock sits flush on ground
	if render != nil {
		render.Y = 0 // Was -3, now 0 to sit on ground
	}

	// Remove hitbox so it can't damage anymore
	if ecs.HasComponent[components.Hitbox](world, eid) {
		ecs.RemoveComponent[components.Hitbox](world, eid)
	}

	// Enable friction and reduce velocity for realistic bounce
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}

	// Reduce horizontal velocity for bounce effect
	physics.Velocity.X *= 0.3 // 30% of original speed

	// Enable friction for gradual slowdown
	physics.FrictionEnabled = true
}

// updateBouncingProjectile handles bouncing projectile behavior.
// Decrements timer and cleans up when expired.
func updateBouncingProjectile(world *ecs.World, eid entities.EntityId, p *components.Projectile, dt float64) {
	if p == nil {
		return
	}

	// Count down bounce timer
	p.BounceTimer -= dt

	if p.BounceTimer <= 0 {
		cleanupProjectile(world, eid)
	}
}
