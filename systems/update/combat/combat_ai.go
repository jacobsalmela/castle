package combat

import (
	"game/components"
	"game/ecs"
	"game/resources"
)

// UpdateCombatAI processes combat events and updates AI targets when enemies take damage.
// This system replaces the AI target logic that was previously in HitFunc callbacks.
//
// When an enemy with AI is hit, this system sets the attacker as the AI's target
// (if the AI doesn't already have a target). This makes enemies aggro on whoever
// attacks them.
//
// This system should run AFTER combat detection (which generates events) but
// BEFORE AI behavior systems (which read the target).
func UpdateCombatAI(world *ecs.World, events []resources.HitEvent) {
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

		// Pure ECS: Any entity with AI component that gets hit will aggro on the attacker
		// No need for enemy-type-specific functions - the component presence is enough
		if ai := ecs.GetComponent[components.AI](world, hit.Target); ai != nil && ai.TargetID == 0 {
			ai.TargetID = hit.Attacker
		}
	}
}
