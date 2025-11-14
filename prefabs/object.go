package prefabs

import (
	"log"
	"math"
	"math/rand/v2"
	"strconv"

	"game/assets"
	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/pkg/config"
	"game/pkg/tilemap"
	"game/resources"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// Object collision properties
	objectCollisionTag    = "object"
	objectShrinkCollision = 1 // Shrink hitbox by this many pixels on each side
)

// NewObjectPrefab constructs a destructible object entity from tilemap data.
//
// Objects are multi-tile destructible environmental elements (crates, barrels, pots, etc)
// that snap to the tile grid. When destroyed, they spawn particles (smoke, debris) and
// reward flakes based on their reward property.
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Top-left position in world coordinates
//   - w, h: Object dimensions (will be snapped to tile grid)
//   - props: Tile properties from map (contains "reward" custom property)
//
// Returns: EntityId of the created entity, or 0 if world is nil
func NewObjectPrefab(world *ecs.World, x, y, w, h float64, props *tilemap.Properties) entities.EntityId {
	if world == nil {
		return 0
	}

	// Get resources
	space := ecs.Resource[bump.Space](world)
	mapRef := ecs.Resource[resources.MapRef](world)
	if mapRef == nil || mapRef.Map == nil {
		log.Panic("object: no active map available for tile lookup")
	}
	worldMap := mapRef.Map

	// Construct tile images from map
	tileImage, normalImage := constructObjectTileImages(worldMap, space, x, y, w, h)

	// Calculate render offset (for multi-tile objects)
	dx := x - math.Floor(x/TileSize)*TileSize
	dy := y - math.Floor(y/TileSize)*TileSize

	// Parse reward from properties
	reward := 0
	if props != nil && props.Custom != nil {
		if rewardStr, ok := props.Custom["reward"]; ok {
			reward, _ = strconv.Atoi(rewardStr)
		}
	}

	// Create entity
	entityID := world.NewEntity()

	// === SPATIAL COMPONENT ===
	transform := &components.Transform{
		X: x,
		Y: y,
		W: w,
		H: h,
	}
	world.AddComponent(entityID, transform)

	// === VISUAL COMPONENT ===
	// Render tile composite image
	render := &components.Render{
		Image:       tileImage,
		NormalImage: normalImage,
		X:           -dx,
		Y:           -dy,
		Layer:       *tilemap.LayerIndex,
	}
	world.AddComponent(entityID, render)

	// === PHYSICS COMPONENT ===
	// Objects can be pushed but have physics
	physics := spatial.NewPhysics()
	physics.GravityEnabled = true
	physics.FrictionEnabled = true
	physics.Weight = 1.0
	world.AddComponent(entityID, physics)

	// === COLLISION COMPONENT ===
	// Solid collider that blocks movement
	collider := &components.Collider{
		Tags:      []string{objectCollisionTag},
		QueryTags: []string{"body", "map", "solid", objectCollisionTag},
		Solid:     true,
		Immovable: false, // Can be pushed
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0,
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(entityID, collider)

	// === HITBOX COMPONENT ===
	// Can be hit by player attacks
	hitbox := NewHitbox(
		objectShrinkCollision,
		objectShrinkCollision,
		w-objectShrinkCollision*2,
		h-objectShrinkCollision*2,
	)
	world.AddComponent(entityID, hitbox)

	// === BEHAVIOR COMPONENT ===
	// Object-specific state
	object := &components.Object{
		TileImage:   tileImage,
		NormalImage: normalImage,
		OffsetX:     -dx,
		OffsetY:     -dy,
		Reward:      reward,
	}
	world.AddComponent(entityID, object)

	// === TEAM COMPONENT ===
	world.AddComponent(entityID, &components.Team{Type: components.TeamNeutral})

	return entityID
}

// constructObjectTileImages builds composite images from map tiles.
// Objects can span multiple tiles, so we stitch them together into a single image.
func constructObjectTileImages(worldMap *tilemap.Map, space *bump.Space, x, y, w, h float64) (*ebiten.Image, *ebiten.Image) {
	// Snap to tile grid
	x = math.Floor(x/TileSize) * TileSize
	y = math.Floor(y/TileSize) * TileSize
	w = math.Ceil(w/TileSize) * TileSize
	h = math.Ceil(h/TileSize) * TileSize

	// Create composite images
	image := ebiten.NewImage(int(w), int(h))
	normalImage := ebiten.NewImage(int(w), int(h))

	// Iterate over tiles in the object's footprint
	for ty := y; ty < y+h; ty += TileSize {
		for tx := x; tx < x+w; tx += TileSize {
			// Get tile from map (removeTiles=true to extract from map)
			tiles, err := tilemap.TilesFromPosition(worldMap, tx, ty, true, space)
			if err != nil {
				log.Printf("object: Failed to get tiles from position (%f, %f): %v", tx, ty, err)
				continue
			}

			// Build draw options with proper transforms
			op := &ebiten.DrawImageOptions{}
			tile := tiles[config.PipelineScreenTag]

			var sx, sy, dx, dy float64 = 1, 1, 0, 0
			if tile.FlipR {
				op.GeoM.Rotate(math.Pi / 2)
				sx = -1
			}
			if tile.FlipX {
				sx, dx = -1, TileSize
				if tile.FlipR {
					sx = 1
				}
			}
			if tile.FlipY {
				sy, dy = -1, TileSize
			}

			op.GeoM.Scale(sx, sy)
			op.GeoM.Translate(tx-x+dx, ty-y+dy)

			// Draw to composite images
			image.DrawImage(tile.Image, op)
			normalImage.DrawImage(tiles[config.PipelineNormalMapTag].Image, op)
		}
	}

	return image, normalImage
}

// ===== PARTICLE PREFABS =====
// These are spawned when objects are destroyed

const (
	debrisDuration     = 5.0 // seconds debris lives
	debrisSize         = 3   // pixels per side
	debrisLayer        = 1   // render layer
	debrisRotationRate = 4.0 // rotation speed multiplier
)

var debrisImage *ebiten.Image

func init() {
	debrisImage = assets.GetSpriteImage("debris")
}

// NewDebrisPrefab spawns a debris particle from a destroyed object.
func NewDebrisPrefab(world *ecs.World, from entities.EntityId) entities.EntityId {
	if world == nil || from == 0 {
		return 0
	}

	// Get source entity transform
	t := ecs.GetComponent[components.Transform](world, from)
	if t == nil {
		return 0
	}

	// Spawn at center of source
	x := t.X + t.W/2
	y := t.Y + t.H/2

	// Create entity
	eid := world.NewEntity()

	// === SPATIAL COMPONENT ===
	transform := &components.Transform{
		X: x,
		Y: y,
		W: debrisSize,
		H: debrisSize,
	}
	world.AddComponent(eid, transform)

	// === VISUAL COMPONENT ===
	render := &components.Render{
		Image: debrisImage,
		Layer: debrisLayer,
	}
	world.AddComponent(eid, render)

	// === PHYSICS COMPONENT ===
	physics := spatial.NewPhysics()
	// Random initial velocity (from flake constants in original code)
	vx := -50.0 + rand.Float64()*100.0 // flakeSpawnMinX to flakeSpawnMaxX
	vy := -30.0 + rand.Float64()*40.0  // flakeSpawnMinY to flakeSpawnMaxY
	physics.SetVelocity(vx, vy)
	physics.GravityEnabled = true
	physics.FrictionEnabled = true
	physics.Weight = 1.0
	world.AddComponent(eid, physics)

	// === COLLISION COMPONENT ===
	collider := &components.Collider{
		Tags:      []string{},
		QueryTags: []string{"map"},
		Solid:     false, // Ghost collision
		Immovable: false,
		Width:     0, // Use Transform size
		Height:    0,
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(eid, collider)

	// === BEHAVIOR COMPONENT ===
	// Debris rotation and lifetime tracking
	debris := &components.Debris{
		Timer:         debrisDuration * (0.5 + rand.Float64()),
		RotationSpeed: (rand.Float64() - 0.5) * 2 * debrisRotationRate * math.Pi,
		ImageIndex:    0,
	}
	world.AddComponent(eid, debris)

	return eid
}
