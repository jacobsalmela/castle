package prefabs

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

const (
	// Visual properties
	ghoulAnimFile = "ghoul" // Animation file base name
	ghoulWidth    = 9       // Collision width in pixels
	ghoulHeight   = 13      // Collision height in pixels

	// Sprite offset configuration
	ghoulOffsetX    = -6.5 // Sprite offset when facing right
	ghoulOffsetY    = -4   // Vertical sprite offset
	ghoulOffsetFlip = 14   // Sprite offset when facing left

	// Physics properties
	ghoulMaxSpeed = 40.0 // Maximum horizontal velocity (pixels/second)
	ghoulWeight   = 0.85 // Body weight for knockback calculations

	// Combat stats
	ghoulHealth = 70   // Hit points
	ghoulPoise  = 21.0 // Knockback resistance
	ghoulExp    = 20   // Experience points awarded on death

	// Default behavior
	ghoulDefaultRocks = 0 // Default rock count (0 = melee only)

	// Visual effect durations
	ghoulFlashDuration = 0.8 // Flash effect duration in seconds
	ghoulDieDuration   = 1.0 // Death fade duration in seconds
)

// NewGhoulPrefab constructs a ghoul enemy entity using the shared enemy factory.
//
// Ghouls are intelligent enemies with two AI modes:
//  1. Aggressive (default): Approaches player, performs melee attacks and jump-attacks
//  2. Poacher (if rocks > 0): Maintains distance, throws rocks, backs away from player
//
// Behavior phases (Aggressive mode):
//  1. Idle: Patrols view area, detects player
//  2. Approach: Moves toward player until in attack range
//  3. Attack: Performs short/long attack or jump-attack
//  4. Repeat
//
// Behavior phases (Poacher mode):
//  1. Idle: Patrols view area, detects player
//  2. BackUp: Moves away from player to maintain range
//  3. Throw: Throws rock at player
//  4. Recover: Pauses after throw
//  5. Repeat (or defend if out of rocks)
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: UNUSED (kept for tilemap compatibility, uses ghoulWidth/ghoulHeight)
//   - flipX: Initial facing direction (true = right, false = left)
//
// Returns: EntityId of the created ghoul, or 0 if world is nil
func NewGhoulPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
	config := EnemyConfig{
		// Animation
		AnimFile:   ghoulAnimFile,
		OffsetX:    ghoulOffsetX,
		OffsetY:    ghoulOffsetY,
		OffsetFlip: ghoulOffsetFlip,

		// Dimensions
		Width:  ghoulWidth,
		Height: ghoulHeight,

		// Physics
		Weight:          ghoulWeight,
		GravityEnabled:  true,
		FrictionEnabled: true,
		MaxVelocityX:    ghoulMaxSpeed,

		// Stats
		Health: int(ghoulHealth),
		Poise:  int(ghoulPoise),
		Exp:    ghoulExp,

		// Visual effects
		FlashDuration: ghoulFlashDuration,
		FlashColor:    [3]float32{1, 1, 1}, // White
		DieDuration:   ghoulDieDuration,

		// Detection (360° vision, 1.5x entity size in all directions)
		DetectionFront: 13.5, // 1.5 * ghoulWidth
		DetectionBack:  13.5,
		DetectionUp:    19.5, // 1.5 * ghoulHeight
		DetectionDown:  19.5,

		// Optional behavior components
		ApproachBehavior: &components.ApproachBehavior{
			Speed:           80.0,
			MaxSpeed:        60.0,
			MinRange:        30.0, // Longer range than rat
			RangeAdjustment: 0.0,
		},
		BackupBehavior: &components.BackupBehavior{
			Speed:    70.0,
			MaxRange: 40.0,
		},
	}

	// Create enemy with Ghoul-specific component
	return NewEnemyPrefab(world, x, y, flipX, config, &components.Ghoul{
		Rocks:         ghoulDefaultRocks,
		Poacher:       false,
		ThrowCooldown: 0,
		Paused:        false,
	})
}

// SetGhoulRocks sets the number of rocks the ghoul can throw.
func SetGhoulRocks(world *ecs.World, eid entities.EntityId, rocks int) {
	if ghoul := ecs.GetComponent[components.Ghoul](world, eid); ghoul != nil {
		ghoul.Rocks = rocks
	}
}

// SetGhoulPoacher enables poacher AI mode for the ghoul.
func SetGhoulPoacher(world *ecs.World, eid entities.EntityId, poacher bool) {
	if ghoul := ecs.GetComponent[components.Ghoul](world, eid); ghoul != nil {
		ghoul.Poacher = poacher
	}
}
