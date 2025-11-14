package decision

import (
	"game/components"
	"game/ecs"
)

// UpdateFacing updates entity facing direction based on AI target position.
func UpdateFacing(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	// Query entities that have Facing, AI, and Transform components
	entities := world.EntitiesWith(
		(*components.Facing)(nil),
		(*components.AI)(nil),
		(*components.Transform)(nil),
	)

	for _, eid := range entities {
		facing := ecs.GetComponent[components.Facing](world, eid)
		ai := ecs.GetComponent[components.AI](world, eid)
		transform := ecs.GetComponent[components.Transform](world, eid)

		// Skip if any component is missing or AI has no target
		if facing == nil || ai == nil || transform == nil || ai.TargetID == 0 {
			continue
		}

		// Get target's Transform via ECS
		targetTransform := ecs.GetComponent[components.Transform](world, ai.TargetID)
		if targetTransform == nil {
			continue
		}

		// Calculate center positions
		actorCenterX := transform.X + transform.W/2
		targetCenterX := targetTransform.X + targetTransform.W/2

		// Update facing direction: true = facing right (target is to the right)
		facing.FlipX = targetCenterX > actorCenterX
	}
}
