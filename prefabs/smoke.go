package prefabs

import (
	"image/color"
	"math"
	"math/rand/v2"

	"game/assets"
	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"game/pkg/tween"
)

const (
	// Visual properties
	smokeSize  = 3 // pixels per side
	smokeLayer = 1 // render layer (above background)

	// Animation timing
	smokeAnimDuration = 1.0 // seconds for full drift and fade animation

	// Physics properties
	smokeMaxDistance = 40.0 // maximum drift distance (pixels)

	// Color properties
	smokeMinAlpha      = 100 // minimum alpha at spawn
	smokeMaxAlpha      = 255 // maximum alpha (opaque)
	smokeRotationSpeed = 4.0 // rotation speed multiplier (× Pi radians/sec)
)

// NewSmokeFrom spawns a smoke particle around the given source EntityId.
// Returns the EntityId of the created smoke particle.
//
// Parameters:
//   - world: ECS world instance (required)
//   - from: Source entity to spawn smoke around
//
// Returns: EntityId of created smoke particle, or 0 if parameters invalid
func NewSmokeFrom(world *ecs.World, from entities.EntityId) entities.EntityId {
	if world == nil || from == 0 {
		return 0
	}

	// Get transform from source entity
	t := ecs.GetComponent[components.Transform](world, from)
	if t == nil {
		return 0
	}

	// Randomize spawn within source entity rect
	sx := t.X + t.W*rand.Float64()
	sy := t.Y + t.H*rand.Float64()

	return createSmokeParticle(world, sx, sy)
}

// createSmokeParticle is the internal constructor for smoke particles.
// Extracts common creation logic to avoid duplication.
func createSmokeParticle(world *ecs.World, sx, sy float64) entities.EntityId {
	distance := rand.Float64() * smokeMaxDistance

	// Create entity
	eid := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions
	transform := &components.Transform{
		X: sx,
		Y: sy,
		W: smokeSize,
		H: smokeSize,
	}
	world.AddComponent(eid, transform)

	// === VISUAL COMPONENT ===
	// Render with random rotation and full opacity
	render := &components.Render{
		Image:      assets.GetSpriteImage("smoke"),
		R:          rand.Float64() * 2 * math.Pi,
		Layer:      smokeLayer,
		ColorScale: color.RGBA{smokeMaxAlpha, smokeMaxAlpha, smokeMaxAlpha, smokeMaxAlpha},
	}
	world.AddComponent(eid, render)

	// === NEW PHYSICS COMPONENT (Phase 2) ===
	// Smoke is purely visual - no collision or gravity needed
	physics := spatial.NewPhysics()
	physics.GravityEnabled = false  // Smoke drifts, doesn't fall
	physics.FrictionEnabled = false // No friction on smoke particles
	world.AddComponent(eid, physics)

	// NOTE: Smoke particles do NOT have a Collider component
	// They are purely visual effects and should never block player movement
	// Previously had a "ghost" collider but it was still being registered in
	// collision space with default "body" tag, causing invisible barriers

	// === BEHAVIOR COMPONENT ===
	// Smoke drift and fade animation
	smoke := &components.Smoke{
		Tween:        tween.New(0, 1, smokeAnimDuration, tween.EaseOutCubic),
		StartX:       sx,
		StartY:       sy,
		TargetX:      (rand.Float64() - 0.5) * 2 * distance,
		TargetY:      (rand.Float64() - 0.5) * 2 * distance,
		RotationRate: (rand.Float64() - 0.5) * 2 * smokeRotationSpeed * math.Pi,
	}
	world.AddComponent(eid, smoke)

	return eid
}
