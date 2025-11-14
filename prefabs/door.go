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
	DoorWidth        = 3  // Door collision width in pixels
	DoorRenderLayer  = -1 // Render behind player
	DoorTileSize     = 8  // Door uses 8x8 tiles
	DoorSpriteHeight = 24 // Door sprite is 24 pixels tall (top 8px + base 16px)

	// Collision configuration
	DoorCollisionTag = "solid" // Tag for solid collision
)

// NewDoorPrefab constructs a Door entity.
//
// Doors are interactive barriers that can be opened by player attacks.
// They have a 2-state lifecycle:
//  1. Closed: Solid collision, visible hitbox, can be hit
//  2. Open: No collision, no hitbox, triggers door chain
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: Door dimensions (width enforced to DoorWidth, height minimum DoorSpriteHeight)
//   - p: Tilemap properties (FlipX for direction, Custom["open"] for pre-opened state)
//
// Returns: EntityId of the created door, or 0 if world is nil
func NewDoorPrefab(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	if world == nil {
		return 0
	}

	// Force door height to match sprite height (3 tiles = 24px)
	// Door objects in Tiled may be taller, but visual/collision should match sprite
	h = DoorSpriteHeight

	// Calculate image offset for flipped doors
	imageOffset := 0.0
	opensFromRight := false
	if p != nil && p.FlipX {
		imageOffset = -TileSize + DoorWidth
		x -= imageOffset
		opensFromRight = true
	}

	// Determine if door starts open
	startOpen := false
	if p != nil && p.Custom != nil {
		startOpen = p.Custom["open"] == "true"
	}

	// Create the entity
	entityID := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions in world space
	transform := &components.Transform{
		X: x,
		Y: y,
		W: DoorWidth,
		H: h,
	}
	world.AddComponent(entityID, transform)

	// === VISUAL COMPONENT ===
	// Render with sprite (Animation will update the frame)
	render := &components.Render{
		X:     imageOffset,
		FlipX: opensFromRight,
		Layer: DoorRenderLayer,
		Image: assets.GetSpriteImage("door"),
	}
	world.AddComponent(entityID, render)

	// === ANIMATION COMPONENT ===
	// Door uses Aseprite animation with two states:
	//   - "idle": Closed door (frame 0, initial state)
	//   - "open": Open door (frame 1, final state after activation)
	anim := &components.Animation{
		FilesName:  "door",
		State:      "idle", // Start in idle state (closed door)
		FSMInitial: "idle",
		FSMTransitions: map[string]string{
			"activate": "open", // Opening animation transitions to open state
			"open":     "open", // Stay in open state (prevents returning to idle)
		},
		OX: 0, OY: 0, // No offsets needed for door
		OXFlip: 0, OYFlip: 0,
		Layer: DoorRenderLayer,
	}
	// Always add the animation component (needed for OpenDoor system)
	world.AddComponent(entityID, anim)

	// Initialize the animation data
	if err := InitializeAnimation(anim); err != nil {
		// Log warning but continue - door will still be functional
		log.Printf("Warning: Failed to initialize door animation: %v", err)
	}

	// === NEW COLLIDER COMPONENT (Phase 2) ===
	// Static collider - door is immovable solid barrier (until opened)
	collider := &components.Collider{
		Tags:      []string{DoorCollisionTag}, // Solid collision
		QueryTags: []string{},
		Solid:     !startOpen, // Start non-solid if pre-opened
		Immovable: true,       // Static object
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0, // Use Transform size
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(entityID, collider)

	// Register in collision space (done automatically by physics system)

	// === HITBOX COMPONENT ===
	// Hitbox so the door can be hit by attacks (offset-based, relative to transform)
	hitbox := NewHitbox(0, 0, DoorWidth, h)
	world.AddComponent(entityID, hitbox)

	// === TEAM COMPONENT ===
	// Team affiliation (neutral, not player or enemy)
	world.AddComponent(entityID, &components.Team{Type: components.TeamNeutral})

	// === BEHAVIOR COMPONENT ===
	// Door-specific state
	door := &components.Door{
		Opened:         startOpen, // Set initial state
		OpensFromRight: opensFromRight,
		Height:         h,
		NeedsInit:      startOpen, // System will apply open state (Pure ECS)
	}
	world.AddComponent(entityID, door)

	return entityID
}

func init() {
	// Door registration happens in game/game_vars.go line 151 via entity map:
	// 151: func(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	//     return prefabs.NewDoorPrefab(world, x, y, w, h, p)
	// }
}
