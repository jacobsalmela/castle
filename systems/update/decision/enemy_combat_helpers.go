package decision

import (
	"math"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/systems/update/entities/animation"
)

// enemy_combat_helpers.go
// Generic helper functions extracted from individual enemy systems.
// These utilities are used by all enemies for common combat patterns.

// BuildAttackFilters constructs a filter list for hitbox contact resolution.
// Includes self hitbox and previously contacted hitboxes to prevent multi-hitting.
//
// Parameters:
//   - self: The attacker's hitbox (to exclude from hits)
//   - contacted: Previously contacted hitboxes this attack (prevents duplicates)
//
// Returns: Combined filter list for ResolveHitboxArea
func BuildAttackFilters(self *components.Hitbox, contacted []*components.Hitbox) []*components.Hitbox {
	count := len(contacted)
	if self != nil {
		count++
	}
	filters := make([]*components.Hitbox, 0, count)
	if self != nil {
		filters = append(filters, self)
	}
	filters = append(filters, contacted...)
	return filters
}

// CollectUniqueContacts merges new contacted hitboxes with existing list.
// Deduplicates to ensure each target is only counted once per attack.
//
// Parameters:
//   - existing: Currently tracked contacted hitboxes
//   - newContacts: Newly contacted hitboxes to add
//
// Returns: Combined list with no duplicates
func CollectUniqueContacts(existing []*components.Hitbox, newContacts []*components.Hitbox) []*components.Hitbox {
	for _, target := range newContacts {
		if target == nil {
			continue
		}
		if isAlreadyContacted(existing, target) {
			continue
		}
		existing = append(existing, target)
	}
	return existing
}

// isAlreadyContacted checks if a hitbox is already in the contacted list.
func isAlreadyContacted(contacted []*components.Hitbox, target *components.Hitbox) bool {
	for _, stored := range contacted {
		if stored == target {
			return true
		}
	}
	return false
}

// ApplyMeleeImpulse applies recoil force to attacker based on contact result.
// Uses different force values for successful hits vs blocked attacks.
//
// Parameters:
//   - world: ECS world instance
//   - eid: Entity performing the attack
//   - facing: Attacker's facing direction
//   - contact: Result of attack (Hit, Block, etc.)
//   - pushForce: Force applied on successful hit
//   - reactForce: Force applied when attack is blocked
func ApplyMeleeImpulse(world *ecs.World, eid entities.EntityId, facing *components.Facing, contact components.ContactType, pushForce, reactForce float64) {
	force := pushForce
	if contact >= components.Block {
		force = reactForce
	}

	if facing != nil && facing.FlipX {
		force *= -1
	}

	physics := ecs.GetComponent[components.Physics](world, eid)
	physics.Velocity.X += force
}

// SetBodyVelocity sets both X and Y velocity.
// Safe wrapper that handles nil checks.
//
// Parameters:
//   - world: ECS world
//   - eid: Entity ID
//   - vx: X velocity to set
//   - vy: Y velocity to set
func SetBodyVelocity(world *ecs.World, eid entities.EntityId, vx, vy float64) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	physics.Velocity.X = vx
	physics.Velocity.Y = vy
}

// UpdateWalkIdleAnimation switches between Walk and Idle based on velocity.
// Common pattern for ground-based enemies.
//
// Parameters:
//   - anim: Animation component to update
//   - physics: Physics component to check velocity
func UpdateWalkIdleAnimation(anim *components.Animation, physics *components.Physics) {
	if anim == nil || physics == nil {
		return
	}

	moving := math.Abs(physics.Velocity.X) > 0.1

	if moving && anim.State == components.IdleTag {
		animation.SetAnimationState(anim, components.WalkTag)
	} else if !moving && anim.State == components.WalkTag {
		animation.SetAnimationState(anim, components.IdleTag)
	}
}
