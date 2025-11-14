package prefabs

import (
	"image/color"
	"log"

	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/pkg/config"
	"game/pkg/tilemap"
	"game/resources"
)

const (
	// Visual properties
	FakeWallLayer        = 200        // Very high layer to ensure it renders above all map tiles
	FakeWallColorR       = uint8(255) // Red component of fake wall tint
	FakeWallColorG       = uint8(220) // Green component of fake wall tint
	FakeWallColorB       = uint8(220) // Blue component of fake wall tint
	FakeWallColorA       = uint8(255) // Alpha component of fake wall tint
	FakeWallCollisionTag = "fakeWall" // Tag for fake wall collision
)

// NewFakeWallPrefab constructs a FakeWall entity.
//
// Fake walls appear as normal wall tiles but crumble when hit by the player.
// They copy the visual appearance from the map tile at their position and
// trigger a chain reaction to adjacent fake walls when destroyed.
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Position in world coordinates (tile position)
//   - w, h: Ignored (fake walls are always TileSize × TileSize)
//   - p: Tile properties (unused for fake walls)
//
// Returns: EntityId of the created entity, or 0 if world is nil
func NewFakeWallPrefab(world *ecs.World, x, y, _, _ float64, _ *tilemap.Properties) entities.EntityId {
	if world == nil {
		return entities.EntityId(0)
	}

	// Get collision space and map for tile lookup
	space := ecs.Resource[bump.Space](world)

	// Get map from resources
	mapRef := ecs.Resource[resources.MapRef](world)
	if mapRef == nil || mapRef.Map == nil {
		log.Panic("fake wall: no active map available for tile lookup")
	}
	worldMap := mapRef.Map

	// Get tile images from map at this position
	tiles, err := tilemap.TilesFromPosition(worldMap, x, y, true, space)
	if err != nil {
		log.Panic("fake wall: Failed to get tiles from position: ", err)
	}

	// Create the entity
	entityID := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions in world space
	transform := &components.Transform{
		X: x,
		Y: y,
		W: TileSize,
		H: TileSize,
	}
	world.AddComponent(entityID, transform)

	// === VISUAL COMPONENT ===
	// Render with tile image from map (with red tint to distinguish)
	render := &components.Render{
		Image:       tiles[config.PipelineScreenTag].Image,
		NormalImage: tiles[config.PipelineNormalMapTag].Image,
		Layer:       FakeWallLayer,
		ColorScale:  color.RGBA{R: FakeWallColorR, G: FakeWallColorG, B: FakeWallColorB, A: FakeWallColorA},
	}
	world.AddComponent(entityID, render)

	// === PHYSICS COMPONENT ===
	// Static physics - fake wall never moves
	physics := spatial.NewPhysicsStatic()
	world.AddComponent(entityID, physics)

	// === COLLISION COMPONENT ===
	// Solid collider until crumbled - blocks movement
	collider := &components.Collider{
		Tags:      []string{"solid", FakeWallCollisionTag},
		QueryTags: []string{},
		Solid:     true,
		Immovable: true, // Static wall
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0, // Use Transform size
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(entityID, collider)

	// === HITBOX COMPONENT ===
	// Can be hit by player attacks
	hitbox := NewHitbox(0, 0, TileSize, TileSize)
	world.AddComponent(entityID, hitbox)

	// === BEHAVIOR COMPONENT ===
	// Fake wall state
	fakeWall := &components.FakeWall{
		Opened: false,
	}
	world.AddComponent(entityID, fakeWall)

	// === TEAM COMPONENT ===
	// Neutral team affiliation
	world.AddComponent(entityID, &components.Team{Type: components.TeamNeutral})

	return entityID
}
