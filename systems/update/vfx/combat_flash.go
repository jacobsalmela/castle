package vfx

import (
	"game/components"
	"game/ecs"
	"game/resources"
)

// ApplyCombatFlash processes combat events and applies flash effects to hit entities.
// This system replaces the flash timer logic that was previously in HitFunc callbacks.
//
// For each unhandled HitEvent, it triggers the flash effect on the target entity
// if the entity has a VisualEffects component.
//
// This system should run AFTER combat detection (which generates events) but
// BEFORE rendering (which reads the flash timer).
func ApplyCombatFlash(world *ecs.World, events []resources.HitEvent) {
	if world == nil {
		return
	}

	// Process all combat events from this frame
	for i := range events {
		hit := &events[i]
		// Skip events already handled by entity-specific systems
		if hit.Handled {
			continue
		}

		// Apply flash effect to target only if not already flashing
		// This prevents rapid re-triggers when hit multiple times in quick succession
		if vfx := ecs.GetComponent[components.VisualEffects](world, hit.Target); vfx != nil {
			// inline IsFlashing and TriggerFlash
			if vfx.FlashTimer <= 0 {
				vfx.FlashTimer = vfx.FlashDuration
			}
		}
	}
}

// UpdateVisualEffects decrements flash timers for all entities with VisualEffects.
// This system should run every frame to animate the flash effect countdown.
//
// This is a separate system from ApplyCombatFlash because:
// 1. Flash timers need to update every frame (not just when hit)
// 2. Other systems may trigger flashes (not just combat)
// 3. Separation of concerns: event processing vs. state animation
func UpdateVisualEffects(world *ecs.World, dt float64) {
	entities := world.EntitiesWith((*components.VisualEffects)(nil))
	for _, eid := range entities {
		if vfx := ecs.GetComponent[components.VisualEffects](world, eid); vfx != nil {
			// inline Update - decrement flash timer
			if vfx.FlashTimer > 0 {
				vfx.FlashTimer -= dt
				if vfx.FlashTimer < 0 {
					vfx.FlashTimer = 0
				}
			}
		}
	}
}
