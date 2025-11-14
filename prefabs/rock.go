package prefabs

import (
	"image/color"
	"math"
	"time"

	"game/assets"
	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
)

const (
	// Visual properties
	rockSize        = 7   // pixels per side
	rockRenderLayer = 100 // high layer (above most entities)
	rockOffsetX     = 3   // horizontal render offset
	rockOffsetY     = -3  // vertical render offset (raised)

	// Color properties
	rockColorR = 255 // red channel
	rockColorG = 40  // green channel
	rockColorB = 40  // blue channel
	rockColorA = 255 // alpha channel

	// Physics properties
	rockDamage      = 5.0  // damage dealt on contact
	rockWeight      = 0.6  // gravity multiplier
	rockMinVel      = 40.0 // minimum horizontal velocity
	rockMaxVel      = 70.0 // maximum horizontal velocity
	rockInitVY      = 45.0 // initial upward velocity
	rockWidthBuffer = 10.0 // target width compensation for trajectory

	// Animation timing
	rockRollingTime = 200 * time.Millisecond // rolling animation duration

	// Spawn configuration
	rockSpawnGrace = 8 // frames to skip owner collision after spawn
)

// SpawnRockWithDirection constructs a rock projectile with explicit direction and owner.
// This is the preferred spawn method during the Pure ECS migration.
//
// Parameters:
//   - world: ECS world instance
//   - x, y: Spawn position
//   - ownerID: EntityId of the throwing entity
//   - directionX: Horizontal direction (-1 for left, +1 for right)
//
// Returns: EntityId of the created projectile, or 0 if world is nil
func SpawnRockWithDirection(world *ecs.World, x, y float64, ownerID entities.EntityId, directionX float64) entities.EntityId {
	if world == nil {
		return 0
	}

	// Create entity
	eid := world.NewEntity()

	// === SPATIAL COMPONENT ===
	transform := &components.Transform{X: x, Y: y, W: rockSize, H: rockSize}
	world.AddComponent(eid, transform)

	// === VISUAL COMPONENT ===
	render := &components.Render{
		Image:       assets.GetSpriteImage("rock"),
		Layer:       rockRenderLayer,
		RollingTime: rockRollingTime,
		ColorScale:  color.RGBA{rockColorR, rockColorG, rockColorB, rockColorA},
		X:           rockOffsetX,
		Y:           rockOffsetY,
	}
	world.AddComponent(eid, render)

	// === PHYSICS COMPONENT ===
	// Movable physics - rock affected by gravity, no friction during flight
	physics := spatial.NewPhysics()
	physics.Weight = rockWeight
	physics.GravityEnabled = true
	physics.FrictionEnabled = false
	physics.Velocity = components.Vec2{X: directionX * rockMaxVel, Y: -rockInitVY}
	physics.MaxVelocity = components.Vec2{X: rockMaxVel, Y: rockMaxVel}
	world.AddComponent(eid, physics)

	// === COLLISION COMPONENT ===
	// Solid collider for damage detection
	collider := &components.Collider{
		Tags:      []string{"projectile", "body"},
		QueryTags: []string{"enemy", "map", "solid"},
		Solid:     true,
		Immovable: false,
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0, // Use Transform size
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(eid, collider)

	// === COMBAT COMPONENT ===
	// Hitbox offset to match visual offset (sprite is raised by rockOffsetY=-3)
	hitbox := NewHitbox(0, rockOffsetY, rockSize, rockSize)
	world.AddComponent(eid, hitbox)

	// === BEHAVIOR COMPONENT ===
	projectile := &components.Projectile{
		Owner:      ownerID,
		Damage:     rockDamage,
		SpawnGrace: rockSpawnGrace,
	}
	world.AddComponent(eid, projectile)

	return eid
}

// SpawnRockProjectile constructs a rock projectile entity.
func SpawnRockProjectile(world *ecs.World, x, y float64, owner entities.EntityId) entities.EntityId {
	if world == nil {
		return 0
	}

	// Calculate initial velocity and direction
	vx, vy := initialProjectileVelocity(x, y, owner)
	dirSign := projectileDirection(x, owner)

	// Create entity
	eid := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions
	transform := &components.Transform{
		X: x,
		Y: y,
		W: rockSize,
		H: rockSize,
	}
	world.AddComponent(eid, transform)

	// === VISUAL COMPONENT ===
	// Render with rolling animation
	render := &components.Render{
		Image:       assets.GetSpriteImage("rock"),
		Layer:       rockRenderLayer,
		RollingTime: rockRollingTime,
		ColorScale:  color.RGBA{rockColorR, rockColorG, rockColorB, rockColorA},
		X:           rockOffsetX,
		Y:           rockOffsetY,
	}
	world.AddComponent(eid, render)

	// === PHYSICS COMPONENT ===
	// Movable physics - rock with gravity, no friction
	physics := spatial.NewPhysics()
	physics.Weight = rockWeight
	physics.GravityEnabled = true
	physics.FrictionEnabled = false
	physics.Velocity = components.Vec2{X: vx, Y: -vy}
	physics.MaxVelocity = components.Vec2{X: rockMaxVel, Y: 0} // Only clamp X
	world.AddComponent(eid, physics)

	// === COLLISION COMPONENT ===
	// Solid collider for damage detection
	collider := &components.Collider{
		Tags:      []string{"projectile", "body"},
		QueryTags: []string{"enemy", "map", "solid"},
		Solid:     true,
		Immovable: false,
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0, // Use Transform size
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(eid, collider)

	// === COMBAT COMPONENT ===
	// Hitbox for damage detection (offset-based, relative to transform)
	hitbox := NewHitbox(0, 0, rockSize, rockSize)
	world.AddComponent(eid, hitbox)

	// === BEHAVIOR COMPONENT ===
	// Projectile data with owner tracking
	ownerID := entities.EntityId(0)

	projectile := &components.Projectile{
		Owner:      ownerID,
		Damage:     rockDamage,
		SpawnGrace: rockSpawnGrace,
	}
	world.AddComponent(eid, projectile)

	// Apply directional velocity
	speed := math.Min(math.Max(math.Abs(vx), rockMinVel), rockMaxVel)
	physics.Velocity.X = dirSign * speed

	return eid
}

// initialProjectileVelocity calculates the initial X and Y velocity for a projectile.
// Returns (vx, vy) where vy is the upward component.
func initialProjectileVelocity(x, y float64, owner entities.EntityId) (float64, float64) {
	vy := rockInitVY
	vx := rockMaxVel
	// Owner-based trajectory calculation removed - will be added when Actor interface migrated
	return vx, vy
}

// projectileDirection determines the horizontal direction (-1 left, +1 right).
// Returns 1 (right) by default.
func projectileDirection(x float64, owner entities.EntityId) float64 {
	// Owner-based direction removed - will be added when Actor interface migrated
	return 1
}
