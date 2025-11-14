package nodes

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/systems/update/decision/nodes/helpers"
	"game/systems/update/entities/animation"
)

// PlayAnimation creates an action that plays an animation and waits for it to complete.
// Returns Success when animation is no longer playing the specified state.
// Returns Running while animation is playing.
// Returns Failure if animation component not found.
func PlayAnimation(animationState string) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			anim := ecs.GetComponent[components.Animation](world, eid)
			if anim != nil && animation.HasState(anim, animationState) {
				animation.SetAnimationState(anim, animationState)
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			anim := ecs.GetComponent[components.Animation](world, eid)
			if anim == nil {
				return ai.Failure
			}

			// Check if animation has finished (transitioned to different state)
			if anim.State != animationState {
				return ai.Success
			}

			return ai.Running
		},
	}
}

// PlayAnimationOnce creates an action that plays an animation once and completes immediately.
// Useful for triggering an animation without waiting for it to finish.
func PlayAnimationOnce(animationState string) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			anim := ecs.GetComponent[components.Animation](world, eid)
			if anim != nil && animation.HasState(anim, animationState) {
				animation.SetAnimationState(anim, animationState)
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}

// MeleeAttack creates an action that performs a melee attack.
// Uses the MeleeAttackBehavior component for configuration.
// Faces target, plays attack animation, and waits for completion.
func MeleeAttack(attackAnimationName string) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			// Face the target
			aiComp := ecs.GetComponent[components.AI](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			if aiComp != nil && facing != nil && aiComp.TargetID != 0 {
				transform := ecs.GetComponent[components.Transform](world, eid)
				targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
				if transform != nil && targetTransform != nil {
					targetOnRight := helpers.IsTargetOnRight(transform, targetTransform)
					facing.FlipX = !targetOnRight
				}
			}

			// Start attack animation
			anim := ecs.GetComponent[components.Animation](world, eid)
			if anim != nil && animation.HasState(anim, attackAnimationName) {
				animation.SetAnimationState(anim, attackAnimationName)
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			anim := ecs.GetComponent[components.Animation](world, eid)
			if anim == nil {
				return ai.Failure
			}

			// Wait for animation to complete
			if anim.State != attackAnimationName {
				return ai.Success
			}

			return ai.Running
		},
	}
}

// Block creates an action that enters a blocking state.
// duration: How long to block (in seconds). Use 0 for indefinite.
// blockAnimationName: Name of the block animation state.
func Block(blockAnimationName string, duration float64) *ai.Action {
	elapsed := 0.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			elapsed = 0
			anim := ecs.GetComponent[components.Animation](world, eid)
			if anim != nil && animation.HasState(anim, blockAnimationName) {
				animation.SetAnimationState(anim, blockAnimationName)
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			elapsed += dt

			if duration > 0 && elapsed >= duration {
				return ai.Success
			}

			// If duration is 0, keep blocking (return Running forever)
			return ai.Running
		},
		OnEnd: func(world *ecs.World, eid entities.EntityId) {
			elapsed = 0
		},
	}
}

// ApplyDamageToTarget creates an action that deals damage to the AI's target.
// This is a simple direct damage application (not using hitboxes).
func ApplyDamageToTarget(damage float64) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			if aiComp == nil || aiComp.TargetID == 0 {
				return
			}

			targetHealth := ecs.GetComponent[components.Health](world, aiComp.TargetID)
			if targetHealth != nil {
				targetHealth.Current -= damage
				if targetHealth.Current < 0 {
					targetHealth.Current = 0
				}
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}

// SetBlackboardValue creates an action that sets a value in the AI's blackboard.
func SetBlackboardValue(key string, value any) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			if aiComp != nil {
				if aiComp.Blackboard == nil {
					aiComp.Blackboard = make(map[string]any)
				}
				aiComp.Blackboard[key] = value
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}

// ClearBlackboardValue creates an action that removes a key from the AI's blackboard.
func ClearBlackboardValue(key string) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			if aiComp != nil && aiComp.Blackboard != nil {
				delete(aiComp.Blackboard, key)
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}
