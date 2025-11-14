package prefabs

import (
	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// NewBlockPrefab constructs a pushable block that uses the size from Tiled.
// The block is solid, can be pushed/hit, and can be jumped on like a platform.
//
// Phase 13: Pure ECS prefab - returns entities.EntityId, assembles components directly.
func NewBlockPrefab(world *ecs.World, x, y, w, h float64) entities.EntityId {
	if world == nil {
		return 0
	}

	// Create entity
	eid := world.NewEntity()

	// Create Transform component
	transform := &components.Transform{X: x, Y: y, W: w, H: h}
	world.AddComponent(eid, transform)

	// Create Render component with simple colored block
	img := ebiten.NewImage(int(w), int(h))
	img.Fill(color.RGBA{200, 200, 200, 255}) // Light gray
	render := &components.Render{Image: img, Layer: 3}
	world.AddComponent(eid, render)

	// Create Hitbox component
	hitbox := NewHitbox(0, 0, w, h) // Offset-based hitbox (relative to transform)
	world.AddComponent(eid, hitbox)

	// === PHYSICS COMPONENT ===
	// Movable physics - block can be pushed, affected by gravity
	physics := spatial.NewPhysics()
	physics.Weight = 1.0 // Affected by gravity
	physics.GravityEnabled = true
	physics.FrictionEnabled = true
	world.AddComponent(eid, physics)

	// === COLLISION COMPONENT ===
	// Solid movable collider - player can stand on it and push it
	collider := &components.Collider{
		Tags:      []string{"body", "solid"},
		QueryTags: []string{"body", "map", "solid"},
		Solid:     true,
		Immovable: false, // Can be pushed
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0, // Use Transform size
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(eid, collider)

	return eid
}
