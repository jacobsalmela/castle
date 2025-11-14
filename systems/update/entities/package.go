package entities

// Package entities contains Phase 7 systems: entity-specific updates and interactions.
//
// This phase handles entity lifecycle, animations, and interactive objects.
// Systems in this phase:
//   - Animation Updates: Frame advancement, state transitions, callbacks
//   - Animation→Render Sync: Updates Render components from Animation data
//   - Interactive Objects: Doors, chests, graves, fake walls
//   - Hazards: Spikes
//   - Projectiles: Projectile updates and bouncing
//
// Order: This phase runs AFTER state updates, BEFORE VFX
//   - State updates have been applied (health, timers, etc.)
//   - Animation state changes are reflected here
//   - Visual effects spawn based on entity behaviors (next phase)
//
// Performance: ~0.5-1ms per frame (animation systems are relatively fast)

import (
	"game/ecs"
	"game/resources"
	"game/systems/update/entities/animation"
)

// Update runs all Phase 7 systems: entity updates and interactions.
//
// Order within phase:
//  1. Animation.Update - Advance animation frames, state transitions, sync to Render
//  2. ApplyDoorOpen - Process door opening events
//  3. ApplyChestOpen - Process chest opening events
//  4. UpdateChestAnimation - Advance chest opening animations
//  5. ApplyFakeWallOpen - Process fake wall crumbling events
//  6. UpdateGrave - Process grave interactions (save/rest)
//  7. UpdateSpike - Apply spike hazard damage
//  8. UpdateProjectiles - Update projectile behavior and bouncing
//
// All systems in this phase handle entity-specific logic and interactions.
func Update(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	// Animation systems (orchestrated by animation package)
	animation.Update(world, dt)

	// Get combat events for interactive objects
	eventQueue := ecs.Resource[resources.EventQueue](world)
	var hitEvents []resources.HitEvent
	if eventQueue != nil {
		hitEvents = eventQueue.Hits
	}

	// Interactive objects (doors, chests, fake walls)
	InitializeDoors(world)
	ApplyDoorOpen(world, hitEvents)
	ApplyChestOpen(world, hitEvents)
	UpdateChestAnimation(world, dt)
	ApplyFakeWallOpen(world, hitEvents)
	UpdateFakeWallTimers(world, dt)

	// Grave interactions (save/rest)
	UpdateGrave(world, dt)

	// Hazards and projectiles
	UpdateSpike(world, nil, dt)
	UpdateProjectiles(world, nil, dt)
	UpdateObject(world, hitEvents, dt)

	// Lifecycle systems
	UpdateStartDoorLifecycle(world, dt)
}
