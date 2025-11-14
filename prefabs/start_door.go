package prefabs

import (
	"time"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/pkg/tilemap"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// Effect properties
	StartDoorSpawnDelay     = 2 * time.Second // Delay before smoke and removal
	StartDoorSmokeCount     = 10              // Number of smoke particles to spawn
	StartDoorShakeDuration  = 0.1             // Camera shake duration in seconds
	StartDoorShakeMagnitude = 0.1             // Camera shake magnitude
)

// NewStartDoorPrefab constructs a StartDoor entity.
//
// StartDoors are temporary visual barriers at spawn points that:
//  1. Display as a colored rectangle for 2 seconds
//  2. Spawn smoke particles with camera shake
//  3. Self-destruct after effects complete
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: Door dimensions in pixels
//   - p: Tilemap properties (unused, for interface compatibility)
//
// Returns: EntityId of the created door, or 0 if world is nil
func NewStartDoorPrefab(world *ecs.World, x, y, w, h float64, _ *tilemap.Properties) entities.EntityId {
	if world == nil {
		return 0
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Create colored background image using config background color
	image := ebiten.NewImage(int(w), int(h))
	image.Fill(cfg.Screen.BackgroundColor.ToColor())

	// Create the entity
	entityID := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions in world space
	transform := &components.Transform{
		X: x,
		Y: y,
		W: w,
		H: h,
	}
	world.AddComponent(entityID, transform)

	// === VISUAL COMPONENT ===
	// Render as solid color rectangle
	render := &components.Render{
		Image: image,
		Layer: 0, // Default layer
	}
	world.AddComponent(entityID, render)

	// === BEHAVIOR COMPONENT (Pure ECS) ===
	// Lifecycle handled by systems/update/entities/start_door_lifecycle.go
	startDoor := &components.StartDoor{
		SpawnTime:    time.Now(),
		DestroyDelay: StartDoorSpawnDelay.Seconds(),
	}
	world.AddComponent(entityID, startDoor)

	return entityID
}
