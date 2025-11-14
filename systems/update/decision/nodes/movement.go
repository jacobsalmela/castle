package nodes

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/systems/update/decision/nodes/helpers"
	"math"
)

// ApproachTarget creates an action that moves toward the AI's target.
// Uses the ApproachBehavior component for configuration (speed, range).
// Returns Success when in range, Running while moving, Failure if no target.
func ApproachTarget() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			approach := ecs.GetComponent[components.ApproachBehavior](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			if approach == nil || aiComp == nil || physics == nil || aiComp.TargetID == 0 {
				return ai.Failure
			}

			transform := ecs.GetComponent[components.Transform](world, eid)
			targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
			if transform == nil || targetTransform == nil {
				return ai.Failure
			}

			// Check if we're already in range
			dist := helpers.CalculateDistance(transform, targetTransform)

			// Account for range adjustment
			effectiveMinRange := approach.MinRange + approach.RangeAdjustment
			if dist <= effectiveMinRange {
				return ai.Success
			}

			// Move toward target
			dx, _ := helpers.CalculateDirection(transform, targetTransform)
			physics.Velocity.X += math.Copysign(approach.Speed*dt, dx)

			return ai.Running
		},
	}
}

// BackupFromTarget creates an action that moves away from the AI's target.
// Uses the BackupBehavior component for configuration (speed, max range).
// Returns Success when far enough, Running while moving, Failure if no target.
func BackupFromTarget() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			backup := ecs.GetComponent[components.BackupBehavior](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			if backup == nil || aiComp == nil || physics == nil || aiComp.TargetID == 0 {
				return ai.Failure
			}

			transform := ecs.GetComponent[components.Transform](world, eid)
			targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
			if transform == nil || targetTransform == nil {
				return ai.Failure
			}

			// Check if we're already far enough
			dist := helpers.CalculateDistance(transform, targetTransform)

			if dist >= backup.MaxRange {
				return ai.Success
			}

			// Move away from target (opposite direction)
			dx, _ := helpers.CalculateDirection(transform, targetTransform)
			physics.Velocity.X += math.Copysign(backup.Speed*dt, -dx)

			return ai.Running
		},
	}
}

// FaceTarget creates an action that updates the Facing component to face the target.
// Returns Success after facing target, Failure if no target.
func FaceTarget() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)

			if aiComp == nil || facing == nil || aiComp.TargetID == 0 {
				return ai.Failure
			}

			transform := ecs.GetComponent[components.Transform](world, eid)
			targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
			if transform == nil || targetTransform == nil {
				return ai.Failure
			}

			// Calculate if target is to our right
			targetOnRight := helpers.IsTargetOnRight(transform, targetTransform)

			// FlipX = true when target is on the left (facing left)
			facing.FlipX = !targetOnRight

			return ai.Success
		},
	}
}

// MoveToPosition creates an action that moves to a specific position.
// Position is retrieved from the blackboard using the specified key.
// Returns Success when close to position, Running while moving, Failure if position not found.
func MoveToPosition(blackboardKey string, speed float64, arrivalThreshold float64) *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)
			transform := ecs.GetComponent[components.Transform](world, eid)

			if aiComp == nil || physics == nil || transform == nil || aiComp.Blackboard == nil {
				return ai.Failure
			}

			// Get target position from blackboard
			posVal, exists := aiComp.Blackboard[blackboardKey]
			if !exists {
				return ai.Failure
			}

			// Expect position as [2]float64 or similar
			pos, ok := posVal.([2]float64)
			if !ok {
				return ai.Failure
			}

			// Check if we're close enough
			myX, myY := helpers.GetCenter(transform)
			dx := pos[0] - myX
			dy := pos[1] - myY
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= arrivalThreshold {
				return ai.Success
			}

			// Move toward position
			physics.Velocity.X += math.Copysign(speed*dt, dx)

			return ai.Running
		},
	}
}

// Jump creates an action that makes the entity jump.
// Returns Success after applying jump force, Failure if not grounded or no physics.
func Jump(jumpForce float64) *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			physics := ecs.GetComponent[components.Physics](world, eid)
			if physics != nil && physics.Grounded {
				physics.Velocity.Y = jumpForce
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			physics := ecs.GetComponent[components.Physics](world, eid)
			if physics == nil {
				return ai.Failure
			}
			// Success after applying jump force
			return ai.Success
		},
	}
}

// StopMovement creates an action that sets velocity to zero.
// Returns Success after stopping.
func StopMovement() *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			physics := ecs.GetComponent[components.Physics](world, eid)
			if physics != nil {
				physics.Velocity.X = 0
				physics.Velocity.Y = 0
			}
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			return ai.Success
		},
	}
}
