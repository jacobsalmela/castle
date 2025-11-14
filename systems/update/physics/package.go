// Package physics contains Phase 4 systems for movement and collision.
// This phase applies forces, handles collision detection and resolution,
// and updates entity positions based on physics calculations.
//
// Systems in this phase:
//   - Physics: Core physics simulation (gravity, velocity, collision detection/resolution)
//   - Body Physics: Extended physics for complex bodies (multiple hitboxes, etc.)
//   - Jump: Player jump mechanics and aerial control
//
// Velocity Helpers:
// This package also provides velocity manipulation utilities used during
// the AI decision phase (Phase 3) to set movement intent before physics
// integration (Phase 4):
//   - SnapshotMaxVelocity: Read current max velocity settings
//   - AddVelocity/AddVelocityX/AddVelocityY: Apply velocity changes
//   - SetVelocity/SetMaxVelocity: Set absolute velocity values
//   - FacingMultiplier: Get directional multiplier from facing component
//
// Order: This phase runs after AI decision-making, before combat resolution.
//   - Phase 3 (Decision): AI uses velocity helpers to set movement intent
//   - Phase 4 (Physics): Physics applies forces and resolves collisions
//   - Phase 5 (Combat): Combat uses final positions for hitbox checks
//
// Performance: ~1-2ms per frame (collision detection can be expensive)
package physics

import (
	"game/ecs"
)

// Update runs all Phase 4 systems: physics simulation and collision.
//
// This is the entry point for the physics phase. It applies movement,
// handles collision detection/resolution, and updates entity positions.
//
// Order within phase:
//  1. Physics - Apply forces, detect collisions, resolve positions
//  2. Body Physics - Extended physics for complex collision shapes
//  3. Jump - Player jump mechanics (handled within physics currently)
func Update(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	// Run main physics system (includes body physics, hitboxes, and projectiles)
	RunPhysics(world, dt)
}
