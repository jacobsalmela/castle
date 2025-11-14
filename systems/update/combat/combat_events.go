package combat

import (
	"game/ecs"
	"game/resources"
)

// ProcessCombatEvents drains queued hit events and dispatches them to ECS combat reaction systems.
func ProcessCombatEvents(world *ecs.World) {
	if world == nil {
		return
	}
	queue := ecs.Resource[resources.EventQueue](world)
	if queue == nil || !queue.HasHits() {
		return
	}
	events := queue.DrainHits()

	ApplyCombatFlash(world, events)
	UpdateCombatAI(world, events)
	ApplyPlayerCombatEffects(world, events) // Visual effects (shake, freeze)
	ApplyPlayerHurt(world, events)          // Player hit mechanics (poise, stagger)
	ApplyPlayerBlock(world, events)         // Player block mechanics (stamina, chip damage)
	ApplyCombatDefenderStats(world, events)
	ApplyCombatKnockback(world, events)
	// ApplyBlockKnockback(world, events)
	// ApplyDoorOpen(world, events)
	// ApplyFakeWallOpen(world, events)
	// ApplyChestOpen(world, events)
}
