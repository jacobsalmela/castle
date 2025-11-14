package systems

import (
	"game/ecs"
	"game/systems/update/posttick"
	"game/systems/update/pretick"
	"game/systems/update/tick"
)

// ╔════════════════════════════════════════════════════════════════════════════════╗
// ║                    SYSTEMS PACKAGE: UPDATE ORCHESTRATION                       ║
// ║                                                                                ║
// ║  This is called from game/game_update.go every frame (60 TPS).                 ║
// ║                                                                                ║
// ║  Flow:                                                                         ║
// ║  systems.Update()                                                              ║
// ║    ├─ pretick.Update()            (performance, save handling)                 ║
// ║    ├─ tick.Update()               (main ECS update - 10 phases, 40+ systems)   ║
// ║    └─ posttick.Update()           (lighting, input, reset)                     ║
// ║                                                                                ║
// ║  All logic lives in systems/update/* - this file just orchestrates.            ║
// ╚════════════════════════════════════════════════════════════════════════════════╝

// Update runs the complete game update cycle.
// This is a pure orchestrator that delegates to update subpackages.
//
// Parameters:
//   - world: ECS world instance
//   - onReset: Callback for scene reset (used by transitions)
//   - onSave: Callback for saving game state
func Update(world *ecs.World, onReset func(), onSave func() error) error {
	if world == nil {
		return nil
	}

	// PHASE 1: Pre-tick (performance monitoring, save handling)
	if err := pretick.Update(world, onSave); err != nil {
		return err
	}

	// PHASE 2: Main ECS tick (all game systems)
	dt := 1.0 / 60
	tick.Update(world, dt, onReset)

	// PHASE 3: Post-tick (lighting, input, reset)
	return posttick.Update(world, dt)
}
