package pretick

import (
	"game/ecs"
	"game/pkg/perfmon"
	"game/resources"
)

// Update handles pre-tick operations before the main ECS update loop.
// This includes:
//   - Performance monitoring updates
//   - Save game signal processing
//
// Returns an error if save operation fails.
func Update(world *ecs.World, onSave func() error) error {
	// === PERFORMANCE MONITORING ===
	perfmon.GlobalPerfMonitor.Update()

	// === SAVE/LOAD ===
	// Handle save game signal BEFORE main ECS systems run
	// This ensures save happens before any reset transition is triggered
	if err := handleSaveGame(world, onSave); err != nil {
		return err
	}

	return nil
}

// handleSaveGame processes save game signals from the ECS world.
func handleSaveGame(world *ecs.World, onSave func() error) error {
	if world == nil || onSave == nil {
		return nil
	}

	signals := ecs.Resource[resources.GameSignals](world)
	if signals == nil || !signals.ConsumeSave() {
		return nil
	}

	return onSave()
}
