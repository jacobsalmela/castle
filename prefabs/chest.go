package prefabs

import (
	"log"

	"game/assets"
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/tilemap"
)

const (
	// Visual properties
	ChestSpriteW = 14 // Sprite width in pixels (exported for system use)
	ChestSpriteH = 9  // Sprite height in pixels (exported for system use)

	// Collision properties
	chestHitboxW = 14 // Hitbox width (same as sprite)
	chestHitboxH = 9  // Hitbox height (same as sprite)

	// Render settings
	chestRenderLayer = -1 // Render layer (below player)

	// Animation timing (exported for system use)
	ChestSemiOpenDelay = 0.5 // Seconds before transitioning to semi-open
	ChestFullOpenDelay = 1.0 // Seconds before transitioning to fully open

	// Gameplay constants
	chestDefaultReward = 100 // Default number of flake particles spawned
)

// NewChestPrefab constructs a chest entity.
func NewChestPrefab(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	if world == nil {
		return 0
	}

	// Extract flip flag from properties (if provided)
	flipX := false
	imageOffsetX := 0.0
	if p != nil && p.FlipX {
		flipX = true
		// Adjust position and offset for flipped sprite
		imageOffsetX = ChestSpriteW - TileSize*2
		x -= ChestSpriteW - TileSize
	}

	// Create entity
	entityID := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions
	transform := &components.Transform{
		X: x,
		Y: y,
		W: ChestSpriteW,
		H: ChestSpriteH,
	}
	world.AddComponent(entityID, transform)

	// === VISUAL COMPONENT ===
	// Render with chest sprite (Animation will update the frame)
	render := &components.Render{
		X:     imageOffsetX,
		Image: assets.GetSpriteImage("chest"),
		FlipX: flipX,
		Layer: chestRenderLayer,
	}
	world.AddComponent(entityID, render)

	// === ANIMATION COMPONENT ===
	// Chest uses Aseprite animation with three states:
	//   - "idle": Closed chest (frame 0, initial state)
	//   - "activate": Opening animation (frames 1-2)
	//   - "open": Fully open chest (frame 2, stays here)
	anim := &components.Animation{
		FilesName:  "chest",
		State:      "idle", // Start in idle state (closed chest)
		FSMInitial: "idle",
		FSMTransitions: map[string]string{
			"activate": "open", // After opening animation, transition to open state
			"open":     "open", // Stay in open state (prevents returning to idle)
		},
		OX: 0, OY: 0, // No offsets needed for chest
		OXFlip: 0, OYFlip: 0,
		Layer: chestRenderLayer,
	}
	if err := InitializeAnimation(anim); err != nil {
		// Fallback: chest will still render but without animation
		// This allows the game to run even if chest.json is missing
		log.Printf("Warning: Failed to initialize chest animation: %v", err)
	} else {
		world.AddComponent(entityID, anim)
	}

	// === NEW COLLIDER COMPONENT (Phase 2) ===
	// Ghost collider - chests don't block movement, just need hitbox detection
	collider := &components.Collider{
		Tags:      []string{},
		QueryTags: []string{},
		Solid:     false, // Ghost collider doesn't block movement
		Immovable: false,
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0, // Use Transform size
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(entityID, collider)

	// === COMBAT COMPONENT ===
	// Hitbox for player attack detection
	hitbox := NewHitbox(0, 0, chestHitboxW, chestHitboxH) // Offset-based hitbox (relative to transform)
	world.AddComponent(entityID, hitbox)

	// === BEHAVIOR COMPONENT ===
	// Chest state and reward configuration
	chest := &components.Chest{
		Opened:         false,
		Reward:         chestDefaultReward,
		AnimationStage: 0, // 0=closed
		AnimationTimer: 0.0,
	}
	world.AddComponent(entityID, chest)

	// === TEAM COMPONENT ===
	// Mark as neutral (can be interacted with by any team)
	team := &components.Team{Type: components.TeamNeutral}
	world.AddComponent(entityID, team)

	return entityID
}
