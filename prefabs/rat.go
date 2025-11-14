package prefabs

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

// NewRatPrefab constructs a Rat entity using the shared enemy factory.
//
// Rats are small, fast ground enemies with patrol and chase behavior:
//  1. Idle: Stand still, scan patrol area for targets
//  2. Patrol: Simple back-and-forth movement
//  3. Chase: Pursue player when detected
//  4. Attack: Melee strike at close range
//
// Balance values are loaded from config.yml (enemy_balance.rat section).
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: Size parameters (unused, dimensions defined by config)
//   - flipX: Initial facing direction (true = left, false = right)
//
// Returns: EntityId of the created entity, or 0 if world is nil
func NewRatPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
	// Load config from world
	config := GetEnemyConfig(world, "rat")

	// Animation behavior (not in config - specific to rat)
	config.AnimLayer = 5
	config.AnimFSMInitial = "Idle"
	config.AnimFSMTransitions = map[string]string{
		"Attack": "Idle",
	}

	// NOTE: Rat uses behavior tree system (rat_bt.go) instead of behavior components
	// ApproachBehavior and BackupBehavior are not needed for BT implementation
	config.ApproachBehavior = nil
	config.BackupBehavior = nil

	// Create enemy with Rat-specific component
	return NewEnemyPrefab(world, x, y, flipX, config, &components.Rat{
		Paused: false,
		// RemovalTarget set automatically by factory
	})
}
