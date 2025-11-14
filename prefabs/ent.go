package prefabs

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

const (
	// Visual properties
	entAnimFile   = "ent"
	entWidth      = 12
	entHeight     = 16
	entOffsetX    = -8
	entOffsetY    = -4
	entOffsetFlip = 17

	// Physics
	entMaxSpeed = 30

	// Combat stats
	entHealth = 100
	entPoise  = 41
	entExp    = 40

	// Visual effect durations
	entFlashDuration = 0.8 // Flash effect duration in seconds
	entDieDuration   = 1.0 // Death fade duration in seconds
)

// NewEntPrefab constructs an Ent entity using the shared enemy factory.
//
// Ent is a melee enemy with simple AI:
//  1. Idle: Patrols optional view area, detects player
//  2. Approach: Moves toward player until in attack range
//  3. Attack: Swings weapon with cooldown between attacks
//  4. BackUp: Occasionally retreats to maintain spacing
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: Unused (kept for API compatibility)
//   - flipX: Initial sprite orientation (true = facing right)
//
// Returns: EntityId of the created entity, or 0 if world is nil
func NewEntPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
	config := EnemyConfig{
		// Animation
		AnimFile:   entAnimFile,
		OffsetX:    entOffsetX,
		OffsetY:    entOffsetY,
		OffsetFlip: entOffsetFlip,

		// Animation behavior
		AnimLayer:      5,
		AnimFSMInitial: "Idle",
		AnimFSMTransitions: map[string]string{
			"Attack": "Idle",
		},

		// Dimensions
		Width:  entWidth,
		Height: entHeight,

		// Physics
		Weight:          0, // Default weight
		GravityEnabled:  true,
		FrictionEnabled: true,

		// Stats
		Health: entHealth,
		Poise:  entPoise,
		Exp:    entExp,

		// Visual effects
		FlashDuration: entFlashDuration,
		FlashColor:    [3]float32{1, 1, 1}, // White
		DieDuration:   entDieDuration,

		// Detection (360° vision, 1.5x entity size in all directions)
		DetectionFront: 18.0, // 1.5 * entWidth
		DetectionBack:  18.0,
		DetectionUp:    24.0, // 1.5 * entHeight
		DetectionDown:  24.0,

		// Optional behavior components
		ApproachBehavior: &components.ApproachBehavior{
			Speed:           65.0,
			MaxSpeed:        30.0, // Slow ent
			MinRange:        20.0,
			RangeAdjustment: 0.0,
		},
		BackupBehavior: &components.BackupBehavior{
			Speed:    65.0,
			MaxRange: 24.0,
		},
		MeleeAttackBehavior: &components.MeleeAttackBehavior{
			Damage:       40.0,
			PushForce:    10.0,
			ReactForce:   12.0,
			AnimationTag: "Attack",
		},
	}

	// Create enemy with Ent-specific component
	return NewEnemyPrefab(world, x, y, flipX, config, &components.Ent{
		Paused:         false,
		AttackCooldown: 0,
	})
}
