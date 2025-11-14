package prefabs

import (
	"math/rand"
	"time"

	"game/assets"
	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"game/pkg/tilemap"
)

const (
	// Visual properties
	flakeSpriteSize = 3 // pixels per side (flake is square)

	// Animation timing
	flakeSpriteFlipInterval = 0.2 // seconds between sprite frame changes
	flakeHomingDuration     = 0.8 // seconds for tween from current position to target

	// Spawn timing: delay before homing behavior activates
	flakeHomingDelayMin = 500 * time.Millisecond
	flakeHomingDelayMax = 1000 * time.Millisecond

	// Initial velocity range (pixels per second)
	// Flakes spawn with small horizontal spread and slight downward velocity
	flakeVelocityXMin = -50.0
	flakeVelocityXMax = 50.0
	flakeVelocityYMin = -30.0 // Slight upward velocity (negative Y is up)
	flakeVelocityYMax = 10.0  // To slight downward velocity

	// Collision tags
	flakeCollisionTag      = "loot" // Flakes are tagged as loot for filtering
	flakeCollisionQueryTag = "map"  // Flakes query against map geometry
)

// NewFlakePrefab constructs a flake VFX particle.
//
// Flakes have a three-phase lifecycle:
//  1. Spawn: Appears at the given position with random velocity
//  2. Float: Flies randomly for 0.5-1.0 seconds (configurable)
//  3. Home: Tweens toward the target entity (if target != 0)
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates (center of the flake)
//   - from: Source entity that spawned the flake (0 for none, used for filtering)
//   - target: Target entity to home to (0 = no homing, just float forever)
//
// Returns: EntityId of the created flake, or 0 if world is nil
func NewFlakePrefab(world *ecs.World, x, y float64, from, target entities.EntityId) entities.EntityId {
	if world == nil {
		return 0
	}

	// Create the flake entity
	flakeEntity := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions in world space
	transform := &components.Transform{
		X: x,
		Y: y,
		W: flakeSpriteSize,
		H: flakeSpriteSize,
	}
	world.AddComponent(flakeEntity, transform)

	// === VISUAL COMPONENT ===
	// Render the flake sprite with animation (frame 0 to start)
	render := &components.Render{
		Image: assets.GetSpriteImage("flake"),
		Layer: *tilemap.LayerIndex, // Draw on tile layer (above ground, below UI)
	}
	world.AddComponent(flakeEntity, render)

	// === NEW PHYSICS COMPONENT (Phase 2) ===
	// Pure ECS movement - flakes bounce and spread before homing
	physics := spatial.NewPhysics()
	velocityX := calculateRandomVelocity(flakeVelocityXMin, flakeVelocityXMax)
	velocityY := calculateRandomVelocity(flakeVelocityYMin, flakeVelocityYMax)
	physics.SetVelocity(velocityX, velocityY)
	physics.GravityEnabled = true  // Flakes fall with gravity
	physics.Weight = 0.5           // Light gravity (half strength)
	physics.Bounciness = 0.4       // Bounce off ground
	physics.FrictionEnabled = true // Slow down when bouncing
	physics.Friction = 0.95        // High friction to settle quickly
	world.AddComponent(flakeEntity, physics)

	// === NEW COLLIDER COMPONENT (Phase 2) ===
	// Solid collider - flakes bounce off ground tiles
	collider := &components.Collider{
		Tags:      []string{flakeCollisionTag},      // Identify as loot
		QueryTags: []string{flakeCollisionQueryTag}, // Collide with map geometry
		Solid:     true,                             // Physical collision (not ghost)
		Immovable: false,
		OffsetX:   0,
		OffsetY:   0,
		Width:     flakeSpriteSize,
		Height:    flakeSpriteSize,
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(flakeEntity, collider)

	// === FLAKE BEHAVIOR COMPONENT ===
	// Controls animation and homing logic (handled by systems/update/flake.go)
	flakeComponent := &components.Flake{
		From:        from,
		Target:      target,
		RandTargetW: rand.Float64(),                           // Random offset within target width (0.0-1.0)
		RandTargetH: rand.Float64(),                           // Random offset within target height (0.0-1.0)
		Timer:       rand.Float64() * flakeSpriteFlipInterval, // Stagger animation start
		ImageIndex:  0,                                        // Start with first frame
		StartX:      x,                                        // Record spawn position for tween
		StartY:      y,
	}

	// === HOMING BEHAVIOR SCHEDULE (Pure ECS) ===
	// Set the time when homing should start - system will check this each frame
	if shouldEnableHoming(from, target) {
		homingDelayMs := calculateHomingDelay()
		flakeComponent.HomingStartTime = time.Now().Add(homingDelayMs)
	}

	world.AddComponent(flakeEntity, flakeComponent)

	return flakeEntity
}

// calculateRandomVelocity generates a random value within the given range.
// Used for initial flake velocity to create varied floating behavior.
func calculateRandomVelocity(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// calculateHomingDelay returns a random duration between min and max homing delay.
// This creates visual variety in when flakes start moving toward their target.
func calculateHomingDelay() time.Duration {
	minMs := float64(flakeHomingDelayMin)
	maxMs := float64(flakeHomingDelayMax)
	randomMs := minMs + rand.Float64()*(maxMs-minMs)
	return time.Duration(randomMs)
}

// shouldEnableHoming determines if the flake should home to its target.
// Returns false if:
//   - from and target are the same entity (flake spawned by its own target)
//   - target is 0 (no target specified, just float forever)
func shouldEnableHoming(from, target entities.EntityId) bool {
	return from != target && target != 0
}
