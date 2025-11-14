package prefabs

import (
	"game/assets"
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/tilemap"
)

const (
	// Visual properties
	spikeSpriteOffsetX = -2 // Horizontal offset for sprite alignment (pixels)
	spikeSpriteW       = 16 // Sprite width in pixels
	spikeSpriteH       = 16 // Sprite height in pixels

	// Collision properties
	spikeHitboxOffsetX = 1            // Horizontal inset for hitbox (pixels)
	spikeHitboxW       = TileSize - 3 // Hitbox width (narrower than tile)
	spikeHitboxH       = TileSize     // Hitbox height (full tile height)

	// Render settings
	spikeRenderLayer = 1 // Render above background tiles
)

// NewSpikePrefab constructs a spike hazard entity.
//
// Spikes are static environmental hazards with a single-phase lifecycle:
//  1. Passive: Wait for entity contact
//  2. Damage: Apply damage on contact with cooldown timer
//
// Spikes use the hitbox system to detect collisions and apply periodic
// damage to entities that touch them. A per-target cooldown prevents
// instant death from continuous contact.
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Tiled object position (top-left corner in world coordinates)
//   - w, h: Tiled object dimensions (unused, spike uses fixed sizes)
//   - p: Tiled properties (optional, used to extract flip flags)
//
// Returns: EntityId of the created spike, or 0 if world is nil
func NewSpikePrefab(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	if world == nil {
		return 0
	}

	// Extract flip flags from properties (if provided)
	flipX := false
	flipY := false
	if p != nil {
		flipX = p.FlipX
		flipY = p.FlipY
	}

	// Calculate hitbox position (slightly inset from tile edges)
	hitboxX, hitboxY := calculateSpikeHitboxPosition(x, y)

	// Create entity
	entityID := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Collision area for damage detection
	transform := &components.Transform{
		X: hitboxX,
		Y: hitboxY,
		W: spikeHitboxW,
		H: spikeHitboxH,
	}
	world.AddComponent(entityID, transform)

	// === VISUAL COMPONENT ===
	// Render spike sprite with optional flipping
	render := &components.Render{
		X:     spikeSpriteOffsetX,
		Image: assets.GetSpriteImage("spike"),
		FlipX: flipX,
		FlipY: flipY,
		Layer: spikeRenderLayer,
	}
	world.AddComponent(entityID, render)

	// === COMBAT COMPONENT ===
	// Hitbox for damage application (offset-based, relative to transform)
	hitbox := NewHitbox(0, 0, spikeHitboxW, spikeHitboxH)
	world.AddComponent(entityID, hitbox)

	// === COLLISION COMPONENT ===
	// Static ghost collider - doesn't block movement, just detects contact
	collider := &components.Collider{
		Tags:      []string{"hazard", "spike"},
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

	// === BEHAVIOR COMPONENT ===
	// Spike marker for damage system
	spike := &components.Spike{}
	world.AddComponent(entityID, spike)

	// Initialize hitbox system
	world.QueueInitWithID(entityID, uint(entityID))

	return entityID
}

// calculateSpikeHitboxPosition computes the hitbox position with horizontal inset.
// Spikes have a narrower hitbox than the full tile to avoid edge overlap issues.
func calculateSpikeHitboxPosition(tiledX, tiledY float64) (float64, float64) {
	return tiledX + spikeHitboxOffsetX, tiledY
}
