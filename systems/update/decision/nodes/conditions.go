// Package nodes provides concrete behavior tree node implementations for enemy AI.
// These nodes use the core BT types from components/ai and provide game-specific behaviors.
package nodes

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/systems/update/decision/nodes/helpers"
)

// HasTarget checks if the AI has a valid target.
func HasTarget() *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			return aiComp != nil && aiComp.TargetID != 0
		},
	}
}

// IsInRange checks if the AI's target is within the specified distance range.
// minDist: Minimum distance (inclusive)
// maxDist: Maximum distance (inclusive). Use 0 or negative for infinite max range.
func IsInRange(minDist, maxDist float64) *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			if aiComp == nil || aiComp.TargetID == 0 {
				return false
			}

			transform := ecs.GetComponent[components.Transform](world, eid)
			targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
			if transform == nil || targetTransform == nil {
				return false
			}

			// Calculate distance using helper
			dist := helpers.CalculateDistance(transform, targetTransform)

			// Check range bounds
			return dist >= minDist && (maxDist <= 0 || dist <= maxDist)
		},
	}
}

// IsGrounded checks if the entity's physics component indicates it's on the ground.
func IsGrounded() *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			physics := ecs.GetComponent[components.Physics](world, eid)
			return physics != nil && physics.Grounded
		},
	}
}

// IsNotGrounded checks if the entity is in the air.
func IsNotGrounded() *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			physics := ecs.GetComponent[components.Physics](world, eid)
			return physics != nil && !physics.Grounded
		},
	}
}

// HasHealth checks if the entity's health is above the specified threshold.
func HasHealth(minHealth float64) *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			health := ecs.GetComponent[components.Health](world, eid)
			return health != nil && health.Current >= minHealth
		},
	}
}

// HealthBelow checks if the entity's health is below the specified threshold.
func HealthBelow(threshold float64) *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			health := ecs.GetComponent[components.Health](world, eid)
			return health != nil && health.Current < threshold
		},
	}
}

// HealthPercentBelow checks if health is below a percentage of max health.
// percent should be between 0.0 and 1.0 (e.g., 0.5 for 50%)
func HealthPercentBelow(percent float64) *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			health := ecs.GetComponent[components.Health](world, eid)
			if health == nil || health.Max <= 0 {
				return false
			}
			currentPercent := float64(health.Current) / float64(health.Max)
			return currentPercent < percent
		},
	}
}

// IsAnimationState checks if the entity's animation is in the specified state.
func IsAnimationState(stateName string) *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			anim := ecs.GetComponent[components.Animation](world, eid)
			return anim != nil && anim.State == stateName
		},
	}
}

// IsNotAnimationState checks if the entity's animation is NOT in the specified state.
func IsNotAnimationState(stateName string) *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			anim := ecs.GetComponent[components.Animation](world, eid)
			return anim != nil && anim.State != stateName
		},
	}
}

// TargetFacingMe checks if the target is facing toward this entity.
func TargetFacingMe() *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			if aiComp == nil || aiComp.TargetID == 0 {
				return false
			}

			myTransform := ecs.GetComponent[components.Transform](world, eid)
			targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
			targetFacing := ecs.GetComponent[components.Facing](world, aiComp.TargetID)

			if myTransform == nil || targetTransform == nil || targetFacing == nil {
				return false
			}

			// Calculate if we're to the left or right of target
			imOnRight := helpers.IsTargetOnRight(targetTransform, myTransform)

			// If target is flipped and we're on the left, they're facing us
			// If target is not flipped and we're on the right, they're facing us
			return targetFacing.FlipX == imOnRight
		},
	}
}

// HasStamina checks if the entity has at least the specified stamina.
func HasStamina(minStamina float64) *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			stamina := ecs.GetComponent[components.Stamina](world, eid)
			return stamina != nil && stamina.Current >= minStamina
		},
	}
}

// BlackboardCheck checks a boolean value in the AI's blackboard.
func BlackboardCheck(key string) *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			if aiComp == nil || aiComp.Blackboard == nil {
				return false
			}
			val, exists := aiComp.Blackboard[key]
			if !exists {
				return false
			}
			boolVal, ok := val.(bool)
			return ok && boolVal
		},
	}
}

// BlackboardExists checks if a key exists in the AI's blackboard.
func BlackboardExists(key string) *ai.Condition {
	return &ai.Condition{
		Check: func(world *ecs.World, eid entities.EntityId) bool {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			if aiComp == nil || aiComp.Blackboard == nil {
				return false
			}
			_, exists := aiComp.Blackboard[key]
			return exists
		},
	}
}
