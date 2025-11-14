package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

// UpdateGenericEnemy handles common enemy update logic that's shared across all enemies.
// This includes death handling, flash effects, AI updates, and animation switching.
//
// Parameters:
//   - world: ECS world instance
//   - state: Game state interface with QueueRemoval method (can be nil for enemies without removal)
//   - eid: Entity ID of the enemy
//   - paused: Pointer to enemy's paused state flag
//   - removalTarget: Entity to remove when enemy dies (often eid itself)
//   - dt: Delta time in seconds
//
// Returns true if the enemy is dead and being removed (caller should return early).
// Returns false if the enemy is alive and should continue with enemy-specific logic.
//
// Usage pattern:
//
//	func updateEnemy(world *ecs.World, state enemyState, eid entities.EntityId, ..., dt float64) {
//	    if UpdateGenericEnemy(world, state, eid, &enemy.Paused, enemy.RemovalTarget, dt) {
//	        return // Enemy is dead, skip enemy-specific logic
//	    }
//
//	    // Enemy-specific logic here (cooldowns, special mechanics, etc.)
//	}
func UpdateGenericEnemy(world *ecs.World, state interface{ QueueRemoval(entities.EntityId) }, eid entities.EntityId, paused *bool, removalTarget entities.EntityId, dt float64) bool {
	// Get required components
	health := ecs.GetComponent[components.Health](world, eid)
	anim := ecs.GetComponent[components.Animation](world, eid)
	visualEffects := ecs.GetComponent[components.VisualEffects](world, eid)
	deathState := ecs.GetComponent[components.DeathState](world, eid)
	ai := ecs.GetComponent[components.AI](world, eid)
	transform := ecs.GetComponent[components.Transform](world, eid)

	// Handle death animation and cleanup
	if deathState != nil {
		if HandleDeath(world, state, eid, health, deathState, anim, removalTarget, paused, dt) {
			return true // Enemy is dead and being removed
		}
	}

	// Update flash effect using shared VisualEffects component (Pure ECS: inline Update)
	if visualEffects != nil {
		if visualEffects.FlashTimer > 0 {
			visualEffects.FlashTimer -= dt
			if visualEffects.FlashTimer < 0 {
				visualEffects.FlashTimer = 0
			}
		}
	}

	// Apply flash color to animation
	ApplyFlashColorToAnimation(anim, visualEffects, deathState, health)

	// Update AI system
	UpdateAI(world, eid, ai, anim, transform, dt)

	// Update walk/idle animation based on movement (ground enemies)
	// Air-based enemies (like bat) should skip this or use UpdateAirGroundState
	physics := ecs.GetComponent[components.Physics](world, eid)
	UpdateWalkIdleAnimation(anim, physics)

	return false // Enemy is alive, continue with specific logic
}

// ApplyFlashColorToAnimation applies the flash effect color to the animation.
// This handles the flash-when-hit visual effect and resets color when not flashing.
func ApplyFlashColorToAnimation(anim *components.Animation, visualEffects *components.VisualEffects, deathState *components.DeathState, health *components.Health) {
	if anim == nil {
		return
	}

	// Apply flash color when flashing (Pure ECS: inline IsFlashing)
	if visualEffects != nil && visualEffects.FlashTimer > 0 {
		// Convert flash color to UberColor format (0-65535 range, multiplied by 4 for brightness)
		flashColor := visualEffects.FlashColor
		r := uint32(flashColor[0] * 0xffff * 4)
		g := uint32(flashColor[1] * 0xffff * 4)
		b := uint32(flashColor[2] * 0xffff * 4)
		anim.ColorScale = components.UberColor{R: r, G: g, B: b, A: 0xffff * 4}
		return
	}

	// Reset to normal color when not flashing
	// Don't reset during death animation (HandleDeath manages color during fade)
	isDying := deathState != nil && health != nil && health.Current <= 0
	if !isDying {
		anim.ColorScale = components.UberColor{R: 0xffff, G: 0xffff, B: 0xffff, A: 0xffff}
	}
}
