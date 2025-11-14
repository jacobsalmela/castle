// Package decision contains Phase 3 systems for AI and player control.
//
// This is the LARGEST phase - it contains all entity decision-making systems.
// All entities determine what they want to do this frame (movement, attacks,
// behaviors) before physics applies those decisions.
//
// Systems in this phase:
//   - Target Detection: Find nearest enemy/player for AI targeting
//   - Facing: Update entity facing direction based on velocity/target
//   - Player: Process player input into movement/attack intents
//   - Enemy AI: All enemy behavior systems (bat, rat, ghoul, skeleman, ent, crawler)
//   - Boss AI: All boss behavior systems (knight, oscar, ferragus, gram)
//   - Combat AI: AI combat decision-making (attack timing, blocking, etc.)
//   - Enemy Helpers: Shared enemy logic (death handling, common behaviors)
//
// Order: This phase runs after initialization, before physics.
//   - Input has been captured (preupdate phase)
//   - Entities are initialized (initialization phase)
//   - Decisions produce INTENTS (what entities want to do)
//   - Physics (next phase) determines what actually happens
//
// Performance: ~2-3ms per frame (AI can be expensive with many entities)
package decision

import (
	"game/ecs"
)

// Update runs all Phase 3 systems: AI decision-making and player control.
//
// This is the entry point for the decision phase. It processes all entity
// decision-making before physics applies movements.
//
// Order within phase:
//  1. Target Detection - Find targets for all AI entities
//  2. Facing - Update entity facing direction
//  3. Player - Process player input into movement/attack intents
//  4. Enemy AI - All enemy decision-making systems (6 simple enemies)
//  5. Boss AI - All boss decision-making systems (4 bosses)
func Update(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	// Phase 3.1: Common decision systems (IMPLEMENTED)
	UpdateTargetDetection(world, dt)
	UpdateFacing(world, dt)

	// Phase 3.2: Player decision system (IMPLEMENTED)
	UpdatePlayer(world, dt)

	// Phase 3.3: Enemy AI systems (IMPLEMENTED - 16 files migrated)
	// Simple enemies (6)
	UpdateBat(world, world, dt)     // Flying enemy with stalk/attack
	UpdateRat(world, world, dt)     // Jump attack ground enemy
	UpdateGhoul(world, dt)          // Rock throwing / melee enemy
	UpdateSkeleman(world, dt)       // Combo attack enemy
	UpdateEnt(world, world, dt)     // Slow melee with cooldowns
	UpdateCrawler(world, world, dt) // Wall-crawling enemy

	// Boss NPCs (4)
	UpdateOscar(world, world, dt)    // Passive NPC with death dialogue
	UpdateFerragus(world, world, dt) // Invulnerable passive NPC
	UpdateGram(world, world, dt)     // Invulnerable passive NPC
	UpdateKnight(world, world, dt)   // Full boss fight with phases
	UpdateAcedian(world, world, dt)
}
