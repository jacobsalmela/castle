package decision

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/systems/update/entities/animation"
)

// Constants for knight behavior shared across phases
const (
	knightBTApproachSpeed    = 100.0
	knightBTBackSpeed        = 80.0
	knightBTApproachMinRange = 20.0 // Minimum target range for approach behavior
	knightBTBackRange        = 40.0 // Maximum backup distance
)

// newKnightApproach creates a custom approach action for the knight.
//
// The knight moves toward the target until within minimum range.
// Applies velocity acceleration in the direction of the target.
func newKnightApproach() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			if aiComp == nil || aiComp.TargetID == 0 {
				return ai.Failure
			}

			transform := ecs.GetComponent[components.Transform](world, eid)
			targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
			if transform == nil || targetTransform == nil {
				return ai.Failure
			}

			physics := ecs.GetComponent[components.Physics](world, eid)
			if physics == nil {
				return ai.Failure
			}

			// Calculate distance to target (center to center)
			myCenter := transform.X + transform.W/2
			targetCenter := targetTransform.X + targetTransform.W/2
			dx := targetCenter - myCenter

			// Calculate actual distance
			dy := (transform.Y + transform.H/2) - (targetTransform.Y + targetTransform.H/2)
			distance := dx*dx + dy*dy // squared distance for efficiency

			// Check if within approach range (use squared distance to avoid sqrt)
			minRangeSq := knightBTApproachMinRange * knightBTApproachMinRange
			if distance <= minRangeSq {
				return ai.Success
			}

			// Apply velocity toward target
			if dx > 0 {
				physics.Velocity.X += knightBTApproachSpeed * dt
			} else {
				physics.Velocity.X += -knightBTApproachSpeed * dt
			}

			return ai.Running
		},
	}
}

// newKnightBackup creates a custom backup action for the knight.
//
// The knight moves away from the target until beyond backup range.
// Applies velocity acceleration in the opposite direction of the target.
func newKnightBackup() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			aiComp := ecs.GetComponent[components.AI](world, eid)
			if aiComp == nil || aiComp.TargetID == 0 {
				return ai.Failure
			}

			transform := ecs.GetComponent[components.Transform](world, eid)
			targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
			if transform == nil || targetTransform == nil {
				return ai.Failure
			}

			physics := ecs.GetComponent[components.Physics](world, eid)
			if physics == nil {
				return ai.Failure
			}

			// Calculate direction to target
			myCenter := transform.X + transform.W/2
			targetCenter := targetTransform.X + targetTransform.W/2
			dx := targetCenter - myCenter

			// Check if we're already far enough
			dy := (transform.Y + transform.H/2) - (targetTransform.Y + targetTransform.H/2)
			distance := dx*dx + dy*dy
			maxRangeSq := knightBTBackRange * knightBTBackRange
			if distance >= maxRangeSq {
				return ai.Success
			}

			// Apply velocity away from target (opposite direction)
			if dx > 0 {
				physics.Velocity.X += -knightBTBackSpeed * dt
			} else {
				physics.Velocity.X += knightBTBackSpeed * dt
			}

			return ai.Running
		},
	}
}

// newKnightAttack creates a melee attack action for the knight.
//
// Plays the attack animation with hitbox callbacks for damage detection.
// Returns Success when the attack animation completes.
func newKnightAttack() *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			knight := ecs.GetComponent[components.Knight](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)

			if knight == nil || facing == nil || anim == nil {
				return
			}

			KnightEnterAttack(world, eid, knight, facing, "Attack", anim)
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			anim := ecs.GetComponent[components.Animation](world, eid)

			// Check if attack animation ended
			if anim == nil || anim.State != "Attack" {
				return ai.Success
			}

			return ai.Running
		},
	}
}

// newKnightThrow creates a rock throwing action for the knight.
//
// Plays the throw animation and spawns a rock projectile at a specific frame.
// Uses closure-captured state to ensure rock is spawned exactly once.
// Returns Success when the throw animation completes.
func newKnightThrow() *ai.Action {
	fired := false

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			fired = false

			knight := ecs.GetComponent[components.Knight](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)

			if knight == nil || facing == nil || anim == nil {
				return
			}

			// Face target and start throw animation
			KnightEnsureFacingTarget(world, eid, knight, facing)
			KnightEnterAttack(world, eid, knight, facing, "Attack", anim)

			// Register frame callback for projectile spawning
			animation.RegisterFrameCallback(anim, knightThrowFrame, func() {
				if fired {
					return
				}
				KnightSpawnRock(world, eid, knight)
				fired = true
			})
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			anim := ecs.GetComponent[components.Animation](world, eid)

			// Check if throw animation ended
			if anim == nil || anim.State != "Attack" {
				return ai.Success
			}

			return ai.Running
		},
	}
}
