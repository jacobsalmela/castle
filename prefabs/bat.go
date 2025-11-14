package prefabs

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

// NewBatPrefab constructs a Bat entity using the shared enemy factory.
//
// Bats are flying enemies with a 3-phase lifecycle:
//  1. Idle: Hover in place, scan patrol area for targets
//  2. Stalk: Move toward/away from target maintaining attack range
//  3. Attack: Dive at target with hitbox active
//
// Balance values are loaded from config.yml (enemy_balance.bat section).
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: UNUSED - Kept for tilemap.EntityConstructor interface compatibility
//   - flipX: Initial sprite facing direction (true = left, false = right)
//
// Returns: EntityId of the created bat, or 0 if world is nil
func NewBatPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
	// Load config from world
	config := GetEnemyConfig(world, "bat")

	// Animation behavior (not in config - specific to bat)
	config.AnimLayer = 5
	config.AnimFSMInitial = "Idle"
	config.AnimFSMTransitions = map[string]string{
		"Attack": "Idle",
	}

	// Create enemy with Bat-specific component
	return NewEnemyPrefab(world, x, y, flipX, config, &components.Bat{
		Paused: false,
		// RemovalTarget set automatically by factory
	})
}
