package prefabs

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

const (
	// Visual properties
	crawlerAnimFile = "crawler" // Animation file base name
	crawlerWidth    = 11        // Collision width in pixels
	crawlerHeight   = 8         // Collision height in pixels

	// Sprite offset configuration
	crawlerOffsetX    = -4 // Sprite offset when facing right
	crawlerOffsetY    = -4 // Vertical sprite offset
	crawlerOffsetFlip = 10 // Sprite offset when facing left

	// Physics properties
	crawlerWeight = 0.8 // Weight for gravity/knockback

	// Combat stats
	crawlerHealth = 30 // Hit points
	crawlerPoise  = 15 // Knockback resistance
	crawlerExp    = 10 // Experience points on death

	// Visual effect durations
	crawlerFlashDuration = 0.8 // Flash effect duration in seconds
	crawlerDieDuration   = 1.0 // Death fade duration in seconds
)

// NewCrawlerPrefab constructs a Crawler entity using the shared enemy factory.
//
// Crawlers are ground-based melee enemies with a patrol-and-attack pattern:
//  1. Idle: Stand still, scan patrol area for targets
//  2. Patrol: Walk back and forth within view area
//  3. Chase: Move toward target when detected
//  4. Attack: Melee attack when target is in range
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: UNUSED - Kept for tilemap.EntityConstructor interface compatibility
//   - flipX: Initial sprite facing direction (true = left, false = right)
//
// Returns: EntityId of the created crawler, or 0 if world is nil
func NewCrawlerPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
	config := EnemyConfig{
		// Animation
		AnimFile:   crawlerAnimFile,
		OffsetX:    crawlerOffsetX,
		OffsetY:    crawlerOffsetY,
		OffsetFlip: crawlerOffsetFlip,

		// Animation behavior
		AnimLayer:      5,
		AnimFSMInitial: "Idle",
		AnimFSMTransitions: map[string]string{
			"Attack": "Idle",
		},

		// Dimensions
		Width:  crawlerWidth,
		Height: crawlerHeight,

		// Physics
		Weight:          crawlerWeight,
		GravityEnabled:  true,
		FrictionEnabled: true,
		MaxVelocityX:    80.0, // Match approach behavior max speed

		// Stats
		Health: crawlerHealth,
		Poise:  crawlerPoise,
		Exp:    crawlerExp,

		// Visual effects
		FlashDuration: crawlerFlashDuration,
		FlashColor:    [3]float32{1, 1, 1}, // White
		DieDuration:   crawlerDieDuration,

		// Detection (360° vision, 1.5x entity size in all directions)
		DetectionFront: 16.5, // 1.5 * crawlerWidth
		DetectionBack:  16.5,
		DetectionUp:    12.0, // 1.5 * crawlerHeight
		DetectionDown:  12.0,

		// Optional behavior components
		ApproachBehavior: &components.ApproachBehavior{
			Speed:           100.0,
			MaxSpeed:        80.0,
			MinRange:        20.0,
			RangeAdjustment: 0.0,
		},
		BackupBehavior: &components.BackupBehavior{
			Speed:    100.0,
			MaxRange: 40.0,
		},
		MeleeAttackBehavior: &components.MeleeAttackBehavior{
			Damage:       15.0,
			PushForce:    10.0,
			ReactForce:   10.0,
			AnimationTag: "Attack",
		},
	}

	// Create enemy with Crawler-specific component
	return NewEnemyPrefab(world, x, y, flipX, config, &components.Crawler{
		AiMode: "", // Set via SetCrawlerAiMode if needed
		Paused: false,
		// RemovalTarget set automatically by factory
	})
}

// SetCrawlerAiMode sets the AI behavior mode for a crawler entity.
// Common modes: "patrol", "wall" (wall-climbing)
func SetCrawlerAiMode(world *ecs.World, eid entities.EntityId, aiMode string) {
	if crawler := ecs.GetComponent[components.Crawler](world, eid); crawler != nil {
		crawler.AiMode = aiMode
	}
}
