package prefabs

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

const (
	// Visual properties
	skelemanAnimFile = "skeleman" // Animation file base name
	skelemanWidth    = 8          // Collision width in pixels
	skelemanHeight   = 12         // Collision height in pixels

	// Sprite offset configuration
	skelemanOffsetX    = -12 // Sprite offset when facing right
	skelemanOffsetY    = -5  // Vertical sprite offset
	skelemanOffsetFlip = 20  // Sprite offset when facing left

	// Physics properties
	skelemanWeight = 0.85 // Body weight for knockback calculations

	// Combat stats
	skelemanHealth = 110 // Hit points
	skelemanPoise  = 30  // Knockback resistance
	skelemanExp    = 25  // Experience points on death

	// Visual effect durations
	skelemanFlashDuration = 0.8 // Flash effect duration in seconds
	skelemanDieDuration   = 1.0 // Death fade duration in seconds
)

// NewSkelemanPrefab constructs a Skeleman entity using the shared enemy factory.
func NewSkelemanPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
	config := EnemyConfig{
		// Animation
		AnimFile:   skelemanAnimFile,
		OffsetX:    skelemanOffsetX,
		OffsetY:    skelemanOffsetY,
		OffsetFlip: skelemanOffsetFlip,

		// Dimensions
		Width:  skelemanWidth,
		Height: skelemanHeight,

		// Physics
		Weight:          skelemanWeight,
		GravityEnabled:  true,
		FrictionEnabled: true,

		// Stats
		Health: skelemanHealth,
		Poise:  skelemanPoise,
		Exp:    skelemanExp,

		// Visual effects (red flash)
		FlashDuration: skelemanFlashDuration,
		FlashColor:    [3]float32{222, 0, 0}, // Red
		DieDuration:   skelemanDieDuration,

		// Detection (360° vision, 1.5x entity size in all directions)
		DetectionFront: 12.0, // 1.5 * skelemanWidth
		DetectionBack:  12.0,
		DetectionUp:    18.0, // 1.5 * skelemanHeight
		DetectionDown:  18.0,

		// Optional behavior components
		ApproachBehavior: &components.ApproachBehavior{
			Speed:           100.0,
			MaxSpeed:        50.0,
			MinRange:        20.0,
			RangeAdjustment: 0.0,
		},
		BackupBehavior: &components.BackupBehavior{
			Speed:    100.0,
			MaxRange: 30.0,
		},
		MeleeAttackBehavior: &components.MeleeAttackBehavior{
			Damage:       18.0,
			PushForce:    10.0,
			ReactForce:   10.0,
			AnimationTag: "Attack",
		},
	}

	// Create enemy with Skeleman-specific component
	return NewEnemyPrefab(world, x, y, flipX, config, &components.Skeleman{
		Paused: false,
		// RemovalTarget set automatically by factory
	})
}
