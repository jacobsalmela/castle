package preupdate

import (
	"game/ecs"
	"game/resources"
)

// Update handles Phase 1: Time control, input capture, and debug input.
// This phase processes time manipulation (slow-mo, pause) and captures
// player/debug inputs before any game logic runs.
//
// Returns:
//   - dt: Adjusted delta time (may be modified by time control)
//   - skip: True if the update cycle should be skipped (world frozen)
func Update(world *ecs.World, dt float64) (float64, bool) {
	if world == nil {
		return dt, true
	}

	// Update deterministic game time for ECS systems
	if timeCtrl := ecs.Resource[resources.TimeControl](world); timeCtrl != nil {
		timeCtrl.ElapsedTime += dt
	}

	// Capture player input (keyboard/gamepad polling)
	UpdateInput(world, nil, dt)

	// Capture debug input (keyboard shortcuts for debug toggles)
	UpdateDebugInput(world)

	// Update animated map tiles
	world.UpdateMap(dt)

	return dt, false
}
