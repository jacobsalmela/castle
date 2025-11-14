package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"math"
)

// UpdateAISystem processes all AI entities, handling target validation and behavior tree execution.
// This is called from systems/update/tick/tick.go during the DECISION phase.
func UpdateAISystem(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	ais := world.EntitiesWith((*components.AI)(nil))
	for _, eid := range ais {
		ai := ecs.GetComponent[components.AI](world, eid)
		if ai == nil {
			continue
		}

		PruneDeadTarget(world, ai)
		TickBehaviorTree(world, eid, ai, dt)
	}
}

// PruneDeadTarget clears the AI's TargetID if the target's health is zero or below.
func PruneDeadTarget(world *ecs.World, ai *components.AI) {
	if world == nil || ai == nil || ai.TargetID == 0 {
		return
	}

	health := ecs.GetComponent[components.Health](world, ai.TargetID)
	if health != nil && health.Current <= 0 {
		ai.TargetID = 0
	}
}

// GetTargetTransform queries the Transform component for the AI's current target.
// Returns nil if no target or target has no Transform.
func GetTargetTransform(world *ecs.World, ai *components.AI) *components.Transform {
	if world == nil || ai == nil || ai.TargetID == 0 {
		return nil
	}
	return ecs.GetComponent[components.Transform](world, ai.TargetID)
}

// GetPosition returns the position of an entity by querying its Transform component.
func GetPosition(world *ecs.World, eid entities.EntityId) (float64, float64) {
	if world == nil || eid == 0 {
		return 0, 0
	}
	transform := ecs.GetComponent[components.Transform](world, eid)
	if transform == nil {
		return 0, 0
	}
	return transform.X, transform.Y
}

// IsInTargetRange returns true if the AI's target is within [minDist, maxDist].
// If maxDist <= 0, only minDist is checked (infinite max range).
func IsInTargetRange(world *ecs.World, eid entities.EntityId, ai *components.AI, minDist, maxDist float64) bool {
	if world == nil || eid == 0 || ai == nil || ai.TargetID == 0 {
		return false
	}

	targetTransform := GetTargetTransform(world, ai)
	if targetTransform == nil {
		return false
	}

	x, y := GetPosition(world, eid)
	tx, ty := targetTransform.X, targetTransform.Y

	// Calculate Euclidean distance
	dx, dy := x-tx, y-ty
	dist := math.Sqrt(dx*dx + dy*dy)

	inRange := dist >= minDist
	withinMax := maxDist <= 0 || dist <= maxDist

	return inRange && withinMax
}

// TickBehaviorTree executes the entity's behavior tree for this frame.
// This is the NEW system for AI behavior.
// The tree is ticked every frame and returns Success, Failure, or Running.
// Trees automatically re-evaluate from the root each tick, providing natural reactivity.
func TickBehaviorTree(world *ecs.World, eid entities.EntityId, ai *components.AI, dt float64) {
	if world == nil || ai == nil || ai.BehaviorTree == nil {
		return
	}

	// Initialize blackboard if needed
	if ai.Blackboard == nil {
		ai.Blackboard = make(map[string]any)
	}

	// Tick the tree from the root
	ai.BehaviorTree.Tick(world, eid, dt)

	// Note: We don't need to check the return status at this level.
	// The tree handles its own state internally.
	// Selector/Sequence nodes handle Success/Failure/Running appropriately.
}
