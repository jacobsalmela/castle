package prefabs

import (
	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/pkg/tilemap"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// ═══════════════════════════════════════════════════════════════════════════════
// TORCH PREFAB - Interactive Light Source
// ═══════════════════════════════════════════════════════════════════════════════

const (
	// Visual properties
	torchSpriteWidth  = 16 // pixels per frame
	torchSpriteHeight = 16 // pixels per frame

	// Animation timing
	torchAnimInterval = 0.15 // seconds between flame frames

	// Light properties
	torchLightRadius    = 96   // pixels - matches existing lighting system
	torchLightIntensity = 1.0  // full brightness
	torchPulseSpeed     = 10.0 // matches existing shader pulsation (Time*10)

	// Collision properties
	torchCollisionTag = "world" // Tagged as static world object
)

// NewTorchPrefab constructs a torch entity with dynamic lighting.
//
// The torch is an interactive light source that:
//  1. Renders animated flame sprite
//  2. Emits dynamic light via LightEmitter component
//  3. Can be destroyed to extinguish the light
//  4. Optionally blocks player movement (if w/h provided)
//
// The lighting system (systems/draw/lighting/lighting.go) queries all LightEmitter
// components each frame to build the shader light list. When the torch is destroyed
// or its Active flag is set to false, it stops emitting light.
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates (bottom-left from Tiled)
//   - w, h: Collision box size (0 = no collision)
//   - props: Tiled properties (currently unused, reserved for future custom properties)
//
// Returns: EntityId of the created torch entity, or 0 if world is nil
func NewTorchPrefab(world *ecs.World, x, y, w, h float64, props *tilemap.Properties) entities.EntityId {
	if world == nil {
		return 0
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Create the entity
	entityID := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions in world space
	transform := &components.Transform{
		X: x,
		Y: y,
		W: torchSpriteWidth,
		H: torchSpriteHeight,
	}
	world.AddComponent(entityID, transform)

	// === VISUAL COMPONENT ===
	// Simple colored square as placeholder until torch sprite asset is added
	// The main visual effect comes from the lighting shader
	placeholder := ebiten.NewImage(torchSpriteWidth, torchSpriteHeight)
	placeholder.Fill(color.RGBA{255, 200, 100, 255}) // Orange glow color

	render := &components.Render{
		Image: placeholder,
		Layer: *tilemap.LayerIndex, // Foreground layer
	}
	world.AddComponent(entityID, render)

	// === LIGHT EMITTER COMPONENT ===
	// Dynamic lighting for the shader - use config values for defaults
	torchSource := cfg.Lighting.GetLightSource("torch")
	lightEmitter := &components.LightEmitter{
		EntityID:   entityID,
		Radius:     torchSource.Radius,
		Active:     true,
		Intensity:  torchSource.Intensity,
		PulseSpeed: torchPulseSpeed,
		PulsePhase: rand.Float64() * 10.0, // Random phase offset for variety
	}
	world.AddComponent(entityID, lightEmitter)

	// === COLLISION COMPONENT (optional) ===
	// If w/h provided, make torch block player movement
	if w > 0 && h > 0 {
		collider := &components.Collider{
			Tags:      []string{torchCollisionTag},
			QueryTags: []string{},
			Solid:     true,
			Immovable: true, // Static object
			OffsetX:   0,
			OffsetY:   0,
			Width:     0, // Use Transform size
			Height:    0, // Use Transform size
			FilterOut: []entities.EntityId{},
		}
		world.AddComponent(entityID, collider)

		physics := spatial.NewPhysicsStatic()
		world.AddComponent(entityID, physics)

		// Override transform size with Tiled object dimensions
		transform.W = w
		transform.H = h
	}

	// === TEAM COMPONENT ===
	// Mark as neutral object
	world.AddComponent(entityID, &components.Team{Type: components.TeamNeutral})

	return entityID
}

// ExtinguishTorch deactivates a torch's light emission.
// This is called when the torch is destroyed or disabled.
// The lighting system will skip this torch in its light list query.
func ExtinguishTorch(world *ecs.World, eid entities.EntityId) {
	if world == nil || eid == 0 {
		return
	}

	emitter := ecs.GetComponent[components.LightEmitter](world, eid)
	if emitter != nil {
		emitter.Active = false
	}

	// Optional: hide the sprite
	render := ecs.GetComponent[components.Render](world, eid)
	if render != nil {
		render.Image = nil // Hide sprite
	}
}

// IsTorchActive checks if a torch is currently emitting light.
func IsTorchActive(world *ecs.World, eid entities.EntityId) bool {
	if world == nil || eid == 0 {
		return false
	}

	emitter := ecs.GetComponent[components.LightEmitter](world, eid)
	if emitter == nil {
		return false
	}

	return emitter.Active
}
