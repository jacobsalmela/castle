// Package initialization contains Phase 2 systems for deferred entity setup.
// This phase handles entity initialization that was deferred during creation
// and sets up the camera to follow the player.
//
// Systems in this phase:
//   - Init Queue: Processes deferred entity initialization (enemies, interactives)
//   - Camera Follow: Updates camera position to track player
//
// Order: This phase runs after input (preupdate), before AI decision-making (decision).
//   - Entities created last frame are fully initialized
//   - Camera is positioned before AI makes decisions (affects AI targeting)
//
// Performance: ~0.1ms per frame (only runs when entities are queued)
package initialization

import (
	"game/ecs"
)

// Update runs all Phase 2 systems: deferred initialization and camera setup.
//
// This is the entry point for the initialization phase. It ensures all entities
// are fully initialized and the camera is positioned before game logic runs.
//
// Order within phase:
//  1. Init queue - Process deferred entity initialization
//  2. Camera follow - Update camera position to track player
func Update(world *ecs.World) {
	if world == nil {
		return
	}

	RunInitQueue(world)
	RunCameraFollow(world, world)
}
