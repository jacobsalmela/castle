package ai

import (
	"game/entities"
)

// AI is a pure data component for AI behavior.
// All logic is handled by systems/update/decision/ai.go.
type AI struct {
	// TargetID is the ECS entity ID of the current target (typically the player).
	TargetID entities.EntityId

	// BehaviorTree is the root node of this entity's behavior tree.
	BehaviorTree Node

	// Blackboard is a key-value store for sharing data between behavior tree nodes.
	// Use this for temporary state that doesn't belong in components.
	// Example: storing "lastSeenPosition", "alertLevel", etc.
	Blackboard map[string]any
}
