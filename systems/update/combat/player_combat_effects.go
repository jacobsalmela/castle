package combat

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/resources"
)

// ApplyPlayerCombatEffects processes combat events and applies player-specific effects.
// This system replaces the camera shake and world freeze logic that was previously
// in the player's HitFunc callback.
//
// When the player is hit, this system:
// - Shakes the camera for visual impact
// - Freezes the game world briefly for "hit stop" effect
//
// This system should run AFTER combat detection but BEFORE rendering.
func ApplyPlayerCombatEffects(world *ecs.World, events []resources.HitEvent) {
	if world == nil {
		return
	}

	// Find the player entity
	playerEid := findPlayerEntity(world)
	if playerEid == 0 {
		return // No player in world
	}

	// Process all combat events from this frame
	for i := range events {
		hit := &events[i]
		// Skip events already handled by entity-specific systems
		if hit.Handled {
			continue
		}

		// Only process hits on the player
		if hit.Target != playerEid {
			continue
		}

		// Apply effects based on contact type
		switch components.ContactType(hit.Contact) {
		case components.Hit:
			// Full hit - strong camera shake and freeze
			if camera := ecs.Resource[resources.Camera](world); camera != nil {
				camera.Shake(0.5, 1)
			}
			if tc := ecs.Resource[resources.TimeControl](world); tc != nil {
				tc.RequestFreeze(0.1)
			}
		case components.Block:
			// Block - lighter effects (no camera shake, shorter freeze)
			if tc := ecs.Resource[resources.TimeControl](world); tc != nil {
				tc.RequestFreeze(0.05)
			}
		case components.ParryBlock:
			// Parry - medium effects
			if camera := ecs.Resource[resources.Camera](world); camera != nil {
				camera.Shake(0.3, 0.5)
			}
			if tc := ecs.Resource[resources.TimeControl](world); tc != nil {
				tc.RequestFreeze(0.08)
			}
		}
	}
}

// findPlayerEntity searches for the player entity in the ECS world.
// Returns 0 if no player entity is found.
func findPlayerEntity(world *ecs.World) entities.EntityId {
	// Look for entity with Player component
	ents := world.EntitiesWith((*components.Player)(nil))
	if len(ents) > 0 {
		return ents[0]
	}
	return 0
}
