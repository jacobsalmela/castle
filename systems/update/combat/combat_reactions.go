package combat

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/prefabs"
	"game/resources"
	"game/systems/update/state"
)

// ApplyCombatDefenderStats processes hit events for non-player entities.
// Applies damage to Health/Poise/Stamina components and spawns damage number VFX.
// This is the default combat handler for enemies, NPCs, and environmental entities.
// NOTE: Skips player entities - they have dedicated handlers (ApplyPlayerHurt, ApplyPlayerBlock).
func ApplyCombatDefenderStats(world *ecs.World, events []resources.HitEvent) {
	if world == nil || len(events) == 0 {
		return
	}
	for _, evt := range events {
		if evt.Handled {
			continue
		}

		// Skip player entities - they have dedicated combat handlers
		if ecs.HasComponent[components.Player](world, evt.Target) {
			continue
		}

		// Apply damage to components (Health, Poise, Stamina)
		applyCombat(world, evt)
	}
}

// applyCombat applies combat damage to Health/Poise/Stamina components.
// For invulnerable entities (no Health), still applies poise damage and shows hit reactions.
func applyCombat(world *ecs.World, evt resources.HitEvent) {
	health := ecs.GetComponent[components.Health](world, evt.Target)
	contact := components.ContactType(evt.Contact)

	if health != nil {
		applyHealthDamage(world, evt.Target, evt.Damage, contact, health)
	} else {
		applyPoiseDamageOnly(world, evt.Target, evt.Damage, contact)
	}

	resetHeadHealthTimer(world, evt.Target)
}

// applyHealthDamage applies damage to entities with Health component.
// Handles both blocked hits (chip damage + stamina) and direct hits (full damage + poise).
func applyHealthDamage(world *ecs.World, target entities.EntityId, damage float64, contact components.ContactType, health *components.Health) {
	actualDamage := calculateDamage(damage, contact)
	state.AddHealth(world, target, -actualDamage)

	if contact >= components.Block {
		applyStaminaCost(world, target, damage)
	} else {
		applyPoiseCost(world, target, damage)
	}

	spawnDamageNumber(world, target, actualDamage, health.Max)
}

// applyPoiseDamageOnly applies poise damage to invulnerable entities (no Health component).
// Used for entities that can be staggered but not killed.
func applyPoiseDamageOnly(world *ecs.World, target entities.EntityId, damage float64, contact components.ContactType) {
	if !ecs.HasComponent[components.Poise](world, target) {
		return
	}

	poiseDamage := calculateDamage(damage, contact)
	state.AddPoise(world, target, -poiseDamage)
}

// calculateDamage returns the actual damage based on contact type.
// Blocked hits deal reduced damage (chip damage = damage/10).
func calculateDamage(baseDamage float64, contact components.ContactType) float64 {
	if contact >= components.Block {
		return baseDamage / 10 // Chip damage for blocked hits
	}
	return baseDamage // Full damage for direct hits
}

// applyStaminaCost deducts stamina from the target when blocking.
func applyStaminaCost(world *ecs.World, target entities.EntityId, cost float64) {
	if ecs.HasComponent[components.Stamina](world, target) {
		state.AddStamina(world, target, -cost)
	}
}

// applyPoiseCost deducts poise from the target when hit directly.
func applyPoiseCost(world *ecs.World, target entities.EntityId, cost float64) {
	if ecs.HasComponent[components.Poise](world, target) {
		state.AddPoise(world, target, -cost)
	}
}

// resetHeadHealthTimer resets the visibility timer for healthbar/poisebar display.
func resetHeadHealthTimer(world *ecs.World, target entities.EntityId) {
	if timer := ecs.GetComponent[components.HeadHealthTimer](world, target); timer != nil {
		timer.Timer = 3.0 // Standard visibility duration
	}
}

// spawnDamageNumber creates a damage number VFX centered on the entity's sprite.
func spawnDamageNumber(world *ecs.World, targetEID entities.EntityId, damage, maxHealth float64) {
	if world == nil || damage <= 0 {
		return
	}

	// Get target's transform to center the damage number
	transform := ecs.GetComponent[components.Transform](world, targetEID)
	if transform == nil {
		return
	}

	// Calculate sprite center position
	centerX := transform.X + transform.W/2
	centerY := transform.Y + transform.H/2

	// Spawn damage number VFX
	damageNumID := prefabs.NewDamageNumberPrefab(world, centerX, centerY, damage, maxHealth)
	if damageNumID != 0 {
		world.QueueInit(damageNumID)
	}
}

// ApplyCombatKnockback processes hit events for non-player entities.
// Applies knockback impulses scaled inversely with entity weight.
// This is the default knockback handler for enemies, NPCs, and environmental entities.
// NOTE: Skips player entities - they have dedicated knockback handling in ApplyPlayerHurt.
func ApplyCombatKnockback(world *ecs.World, events []resources.HitEvent) {
	if world == nil || len(events) == 0 {
		return
	}

	for _, evt := range events {
		if !shouldApplyKnockback(world, evt) {
			continue
		}

		applyKnockbackForce(world, evt)
	}
}

// shouldApplyKnockback determines if knockback should be applied to the target entity.
// Returns false for handled events, player entities, immovable entities, or entities without physics.
func shouldApplyKnockback(world *ecs.World, evt resources.HitEvent) bool {
	if evt.Handled {
		return false
	}

	if ecs.HasComponent[components.Player](world, evt.Target) {
		return false // Player has dedicated handler
	}

	physics := ecs.GetComponent[components.Physics](world, evt.Target)
	if physics == nil || physics.Weight == 0 {
		return false
	}

	collider := ecs.GetComponent[components.Collider](world, evt.Target)
	if collider != nil && collider.Immovable {
		return false
	}

	return true
}

// applyKnockbackForce calculates and applies knockback to the target entity.
// Force is scaled by weight (lighter = more knockback) and reduced for blocked hits.
func applyKnockbackForce(world *ecs.World, evt resources.HitEvent) {
	physics := ecs.GetComponent[components.Physics](world, evt.Target)

	weight := physics.Weight
	if weight <= 0 {
		weight = 1.0 // Prevent division by zero
	}

	force := calculateKnockbackForce(evt.Damage, weight, components.ContactType(evt.Contact))
	direction := calculateKnockbackDirection(evt.AttackRect, evt.TargetRect)

	physics.Velocity.X += force * direction
}

// calculateKnockbackForce computes knockback force based on damage, weight, and contact type.
// Base knockback multiplier converts damage to velocity impulse, scaled by weight.
// Typical: 20 damage * 2.0 / 0.6 weight = 66 velocity (light enemies)
//          20 damage * 2.0 / 1.0 weight = 40 velocity (medium enemies)
//          20 damage * 2.0 / 2.0 weight = 20 velocity (heavy enemies)
func calculateKnockbackForce(damage, weight float64, contact components.ContactType) float64 {
	const baseMultiplier = 2.0
	force := (damage * baseMultiplier) / weight

	if contact >= components.Block {
		force *= 0.5 // Reduced knockback for blocked hits
	}

	return force
}

// calculateKnockbackDirection determines the direction of knockback based on attacker/target positions.
// Returns 1.0 for rightward push, -1.0 for leftward push.
func calculateKnockbackDirection(attackRect, targetRect bump.Rect) float64 {
	dx := (targetRect.X + targetRect.W/2) - (attackRect.X + attackRect.W/2)
	if dx == 0 {
		dx = 1 // Prevent zero direction
	}

	if dx > 0 {
		return 1.0 // Push right
	}
	return -1.0 // Push left
}
