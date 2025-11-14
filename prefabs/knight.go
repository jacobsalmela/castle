package prefabs

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

// NewKnightPrefab constructs a Knight entity using the shared enemy factory.
//
// Knight is a complex enemy with multiple phases and mechanics:
//  1. Phase 1 (100%-50% health): Defensive stance with shield blocking
//  2. Phase 2 (<50% health): Aggressive with dash attacks
//  3. Shield: Blocks attacks, drains stamina, reflects damage
//
// Balance values are loaded from config.yml (enemy_balance.knight section).
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: Unused (kept for API compatibility)
//   - flipX: Initial sprite orientation (true = facing right)
//
// Returns: EntityId of the created entity, or 0 if world is nil
func NewKnightPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
	// Load config from world
	config := GetEnemyConfig(world, "knight")

	// Create enemy with Knight-specific component
	return NewEnemyPrefab(world, x, y, flipX, config, &components.Knight{
		Paused:       false,
		SecondPhase:  false,
		ShieldActive: false,
		DashCooldown: 0,
	})
}
