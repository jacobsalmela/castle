package combat

// Package combat contains Phase 5 systems for combat resolution.
// This phase processes attacks, applies damage, updates health/poise,
// and handles combat reactions (knockback, stagger, death).
//
// Systems in this phase:
//   - Attack Intent Consume: Consume attack intents and spawn active attacks
//   - Attack In Progress: Process active attack hitboxes
//   - Combat: Core combat resolution (damage calculation, health/poise updates)
//   - Combat Events: Process combat events (hit, blocked, killed, etc.)
//   - Combat Reactions: Apply combat reactions (knockback, stagger)
//   - Stagger: Stagger state management and recovery
//   - Block Knockback: Knockback when player blocks attacks
//   - Player Hurt: Player damage reactions and invincibility frames
//
// Order: This phase runs after physics, before state updates.
//   - Entity positions are finalized from physics
//   - Hitboxes can accurately detect collisions
//   - Damage/healing is applied
//   - State updates (next phase) respond to combat results
//
// Performance: ~1-2ms per frame (combat calculations can be expensive)

import (
	"game/ecs"
)

// Update runs all Phase 5 systems: combat resolution and reactions.
//
// This is the entry point for the combat phase. It processes all attack
// intents, applies damage, and handles combat reactions.
//
// Order within phase:
//  1. Attack Intent Consume - Convert attack intents to active attacks
//  2. Attack In Progress - Process active attack hitboxes
//  3. Combat Events - Process combat events (damage, blocks, etc.)
//  4. Combat Reactions - Apply knockback, stagger, player hurt (integrated in ProcessCombatEvents)
func Update(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	// Consume attack intents and register active attacks
	AttackIntentConsume(world, dt)
	AttackInProgress(world, dt)

	// Process combat events (damage, knockback, etc.)
	ProcessCombatEvents(world)
}
