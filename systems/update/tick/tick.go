package tick

import (
	"fmt"
	"game/ecs"
	"game/pkg/perfmon"
	"game/resources"
	"game/systems/update/cleanup"
	"game/systems/update/combat"
	"game/systems/update/decision"
	"game/systems/update/entities"
	"game/systems/update/initialization"
	"game/systems/update/physics"
	"game/systems/update/preupdate"
	"game/systems/update/state"
	"game/systems/update/ui"
	"game/systems/update/vfx"
)

// Update runs all ECS update systems in the correct order for one game tick.
// This is the main update pipeline organized by phases.
//
// Parameters:
//   - world: ECS world instance
//   - dt: Delta time in seconds
//   - onReset: Callback function for scene reset (used by transitions)
//
// Performance:
// Each system is profiled individually. Enable profiler with Cmd+Shift+0 to see
// performance breakdown. Report prints every 5 seconds when profiler is enabled.
func Update(world *ecs.World, dt float64, onReset func()) {
	if world == nil {
		return
	}

	// Profile each system to find bottlenecks
	prof := perfmon.GlobalProfiler

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 1: TIME & INPUT
	// Process time control and capture player/debug input before any game logic.
	// ═══════════════════════════════════════════════════════════════════════════════
	dt, skip := preupdate.Update(world, dt)
	if skip {
		return // World is frozen, skip all updates
	}

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 2: INITIALIZATION
	// Process deferred entity initialization and setup camera.
	// ═══════════════════════════════════════════════════════════════════════════════
	initialization.Update(world)

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 3: DECISION (AI & Logic)
	// All entities make decisions about what they want to do this frame.
	// This must happen BEFORE physics so AI can influence movement.
	// ═══════════════════════════════════════════════════════════════════════════════
	decision.Update(world, dt)

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 4: PHYSICS
	// Apply movement, collision detection, and resolve positions.
	// ═══════════════════════════════════════════════════════════════════════════════
	physics.Update(world, dt)

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 5: COMBAT RESOLUTION
	// Process attacks, apply damage, update health/poise/stamina.
	// This happens AFTER physics so positions are correct for hitbox checks.
	// ═══════════════════════════════════════════════════════════════════════════════
	combat.Update(world, dt)

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 6: STATE UPDATES
	// Update timers, stats, animations, and other state that responds to combat.
	// ═══════════════════════════════════════════════════════════════════════════════
	state.Update(world, dt)

	// Player reactions to combat
	combat.UpdatePlayerHeal(world)

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 7: ENTITY UPDATES
	// Animation systems, interactive objects, hazards, and projectiles.
	// ═══════════════════════════════════════════════════════════════════════════════
	entities.Update(world, dt)

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 8: VISUAL EFFECTS
	// Update particle systems and visual effects.
	// ═══════════════════════════════════════════════════════════════════════════════
	vfx.Update(world, dt)

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 9: UI
	// Update UI elements and interactive objects.
	// ═══════════════════════════════════════════════════════════════════════════════
	ui.Update(world, dt)

	// ═══════════════════════════════════════════════════════════════════════════════
	// PHASE 10: CLEANUP & TRANSITIONS
	// Handle scene transitions and cleanup at the end of the frame.
	// ═══════════════════════════════════════════════════════════════════════════════
	cleanup.Update(world, dt, onReset)

	// Print profiling report every 5 seconds (if profiler enabled via Cmd+Shift+0)
	debugState := ecs.Resource[resources.DebugState](world)
	if debugState != nil && debugState.IsEnabled(resources.DebugCategoryProfiler) && prof.ShouldReport() {
		fmt.Println(prof.Report())
		prof.Reset()
	}
}
