package combat

import (
	"game/components"
	"game/ecs"
	"game/pkg/config"
	"game/resources"
)

// ApplyPlayerHurt processes Hit combat events for player entities.
// Migrated from entity/actor/actor.go Control.Hurt() method.
//
// Hurt mechanics:
// - ShieldDown: exit blocking state
// - Poise damage: -damage poise
// - Health damage: -damage health
// - Knockback: reactForce (directional based on attacker position)
// - Stagger: if poise depleted or consuming, stagger with increased force
// - Flash timer: visual feedback (handled by existing flash system)
func ApplyPlayerHurt(world *ecs.World, events []resources.HitEvent) {
	if world == nil || len(events) == 0 {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		return
	}

	for _, evt := range events {
		if evt.Handled {
			continue
		}

		// Only process Hit events
		contact := components.ContactType(evt.Contact)
		if contact != components.Hit {
			continue
		}

		// Only process events for player defenders
		player := ecs.GetComponent[components.Player](world, evt.Target)
		if player == nil {
			continue
		}

		// Check if player is invulnerable (debug/testing mode)
		if cfg.Stats.PlayerInvulnerable {
			continue
		}

		// Get required components
		health := ecs.GetComponent[components.Health](world, evt.Target)
		poise := ecs.GetComponent[components.Poise](world, evt.Target)
		playerAnim := ecs.GetComponent[components.Animation](world, evt.Target)
		physics := ecs.GetComponent[components.Physics](world, evt.Target)
		hitbox := ecs.GetComponent[components.Hitbox](world, evt.Target)

		if health == nil || poise == nil || playerAnim == nil || physics == nil || hitbox == nil {
			continue
		}

		// ShieldDown: exit blocking state
		shieldDown(playerAnim, hitbox)

		// Apply poise and health damage
		poise.Current -= evt.Damage
		if poise.Current < 0 {
			poise.Current = 0
		}

		health.Current -= evt.Damage
		if health.Current < 0 {
			health.Current = 0
		}

		// Immediately disable gravity when player dies to prevent falling through ground
		if health.Current <= 0 {
			physics.GravityEnabled = false
			physics.Velocity.X = 0
			physics.Velocity.Y = 0
		}

		// Calculate knockback force (directional)
		force := cfg.Actor.ReactForce

		// Direction based on position difference
		dx := (evt.TargetRect.X + evt.TargetRect.W/2) - (evt.AttackRect.X + evt.AttackRect.W/2)
		if dx < 0 {
			force *= -1
		}

		// Apply knockback
		physics.Velocity.X += force

		// Stagger if poise depleted or in consume animation
		if poise.Current <= 0 || playerAnim.State == components.ConsumeTag {
			staggerForce := force * 2 * (evt.Damage / health.Max)
			facing := ecs.GetComponent[components.Facing](world, evt.Target)
			stagger(playerAnim, physics, hitbox, facing, staggerForce, false, 1.0)
		}

		// Flash timer is handled by ApplyCombatFlash system

		// AI targeting would go here for enemies, but player doesn't need it
	}
}
