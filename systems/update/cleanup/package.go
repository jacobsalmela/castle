// Package cleanup contains Phase 10 systems for scene transitions and cleanup.
// This phase handles scene transitions (level changes, respawns) that should
// execute after all other game logic has completed for the frame.
//
// Systems in this phase:
//   - Transitions: Scene transition effects, level loading, respawn handling
//
// Order: This phase MUST run last - it can reset the world state.
//   - All gameplay has completed for this frame
//   - Transitions can safely trigger scene resets
//   - Deferred reset execution prevents race conditions
//
// Performance: ~0.1ms per frame (most frames have no transitions)
package cleanup

import (
	"game/ecs"
)

// Update runs all Phase 10 systems: scene transitions and end-of-frame cleanup.
//
// This is the entry point for the cleanup phase. It handles game transitions
// (restart, death) and any end-of-frame cleanup logic.
//
// Order within phase:
//  1. Transitions - Process restart and death transitions
//  2. (Future) Entity removal/cleanup systems
func Update(world *ecs.World, dt float64, onReset func()) {
	if world == nil {
		return
	}

	RunTransitions(world, dt, onReset)
}
