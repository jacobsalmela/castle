package combat

import (
	"game/components"
	"game/ecs"
	"game/resources"
)

const (
	blockHorizontalForce = 500.0 // Horizontal push force
)

// ApplyBlockKnockback applies push forces to block entities when hit.
// Blocks get pushed horizontally based on the attack direction.
//
// This system replaces the block prefab's HitFunc callback logic.
//
// Block entities are identified by having Physics but no Stats component.
// When hit, blocks are pushed horizontally based on attacker position.
//
// This system should run during combat event processing, after damage/knockback
// systems but before event cleanup.
func ApplyBlockKnockback(world *ecs.World, events []resources.HitEvent) {
	if world == nil || len(events) == 0 {
		return
	}

	for _, evt := range events {
		// Skip events already handled by entity-specific systems
		if evt.Handled {
			continue
		}

		// Check if target is a block entity
		// Blocks have Physics but NO health components
		physics := ecs.GetComponent[components.Physics](world, evt.Target)
		health := ecs.GetComponent[components.Health](world, evt.Target)

		if physics == nil {
			continue
		}

		// Skip entities with health (they're actors/enemies, not blocks)
		if health != nil {
			continue
		}

		// Apply horizontal push based on attack direction
		// Calculate attacker position relative to target
		attackerCenterX := evt.AttackRect.X + evt.AttackRect.W/2
		targetCenterX := evt.TargetRect.X + evt.TargetRect.W/2
		dx := attackerCenterX - targetCenterX

		force := blockHorizontalForce
		if dx > 0 {
			// Attacker is on the right, push block left
			force *= -1
		}
		// else: Attacker is on the left, push block right

		physics.Velocity.X += force
	}
}
