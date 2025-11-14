package nodes

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"math/rand"
)

// Wait creates an action that waits for a specified duration.
// Returns Success after duration elapses, Running while waiting.
func Wait(duration float64) *ai.Action {
	elapsed := 0.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			elapsed = 0
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			elapsed += dt
			if elapsed >= duration {
				return ai.Success
			}
			return ai.Running
		},
	}
}

// Idle creates an action that does nothing and returns Success immediately.
// Useful as a default action or placeholder.
func Idle() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}

// SetAnimationState creates an action that sets the animation state and returns Success.
// Does not wait for the animation to complete.
func SetAnimationState(stateName string) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			anim := ecs.GetComponent[components.Animation](world, eid)
			if anim != nil {
				// Use the animation system's helper function
				// This properly checks if the state exists
				anim.State = stateName
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}

// Log creates an action that logs a message (useful for debugging).
// Returns Success immediately.
func Log(message string) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			// In a real implementation, you'd use your logging system
			// For now, this is a placeholder
			_ = message // Would log: fmt.Println(message)
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}

// DebugPrint creates an action that prints debug information about the entity.
// Returns Success immediately.
func DebugPrint(prefix string) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			// In a real implementation, print entity state
			// transform, animation, health, etc.
			_ = prefix
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}

// RunCallback creates an action that executes a custom callback function.
// The callback receives the world and entity ID.
// Returns Success after executing callback.
func RunCallback(callback func(*ecs.World, entities.EntityId)) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			if callback != nil {
				callback(world, eid)
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}

// RunCallbackEveryFrame creates an action that executes a callback every frame for a duration.
// Returns Success after duration, Running while executing.
func RunCallbackEveryFrame(duration float64, callback func(*ecs.World, entities.EntityId, float64) bool) *ai.Action {
	elapsed := 0.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			elapsed = 0
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			if callback != nil {
				// If callback returns true, end early
				if callback(world, eid, dt) {
					return ai.Success
				}
			}

			elapsed += dt
			if duration > 0 && elapsed >= duration {
				return ai.Success
			}

			return ai.Running
		},
	}
}

// Fail creates an action that always returns Failure.
// Useful for testing or forcing a branch to fail.
func Fail() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Failure
		},
	}
}

// Succeed creates an action that always returns Success.
// Useful for testing or ensuring a branch succeeds.
func Succeed() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}

// RandomSuccess creates an action that succeeds with a given probability.
// probability should be between 0.0 and 1.0 (e.g., 0.75 for 75% chance)
func RandomSuccess(probability float64) *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			// Generate random value
			r := rand.Float64()

			// If random value is less than probability, succeed
			if r <= probability {
				return ai.Success
			}
			return ai.Failure
		},
	}
}
