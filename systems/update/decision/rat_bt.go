package decision

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/systems/update/entities/animation"
	"game/systems/update/decision/nodes"
)

// buildRatBehaviorTree creates the behavior tree for rat enemies.
// This replicates the legacy action queue behavior using the new BT system.
func buildRatBehaviorTree() ai.Node {
	// Pre-wrap actions with human-friendly names for debug visualization
	idle := ai.WrapForDebug(nodes.Idle(), "Idle", 2)
	approach := ai.WrapForDebug(newRatApproach(), "Approach", 2)
	pauseBrief := ai.WrapForDebug(nodes.Wait(0.1), "Wait(0.1s)", 2)

	// Action choices with descriptive names
	jumpAttack := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newRatJumpAttack(), "JumpAttack", 4),
			ai.WrapForDebug(nodes.Wait(0.5), "Recovery(0.5s)", 4),
		},
	}, "JumpAttackSequence", 3)

	backup := ai.WrapForDebug(&ai.Timeout{
		Duration: 1.0,
		Child:    newRatBackup(),
	}, "Backup(1s)", 3)

	idlePause := ai.WrapForDebug(nodes.Wait(0.5), "IdlePause(0.5s)", 3)

	// Random action selector
	randomAction := ai.WrapForDebug(nodes.NewRandomSelector(
		nodes.WeightedChoice{Weight: 2.0, Node: jumpAttack}, // 57% chance
		nodes.WeightedChoice{Weight: 0.5, Node: backup},     // 14% chance
		nodes.WeightedChoice{Weight: 1.0, Node: idlePause},  // 29% chance
	), "ChooseAction", 2)

	// Assemble the tree
	tree := &ai.Repeat{
		Count: 0, // Infinite loop
		Child: &ai.Sequence{
			Children: []ai.Node{
				idle,
				approach,
				pauseBrief,
				randomAction,
			},
		},
	}

	// Wrap tree with debug tracking for visualization (only in debug builds)
	return ai.WrapForDebug(tree, "RatBT", 0)
}

// newRatBackup creates a custom backup action for the rat.
// This moves the rat away from its target using physics velocity.
// Unlike nodes.BackupFromTarget(), this doesn't require a BackupBehavior component.
func newRatBackup() *ai.Action {
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

			// Check if we're already far enough (maxTargetRange)
			dy := (transform.Y + transform.H/2) - (targetTransform.Y + targetTransform.H/2)
			distance := dx*dx + dy*dy
			maxRangeSq := ratMaxTargetRange * ratMaxTargetRange
			if distance >= maxRangeSq {
				return ai.Success
			}

			// Apply velocity away from target (opposite direction)
			if dx > 0 {
				physics.Velocity.X += -ratSpeed * dt
			} else {
				physics.Velocity.X += ratSpeed * dt
			}

			return ai.Running
		},
	}
}

// newRatApproach creates a custom approach action for the rat.
// This moves the rat toward its target using physics velocity until within range.
// Unlike nodes.ApproachTarget(), this doesn't require an ApproachBehavior component.
func newRatApproach() *ai.Action {
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
			minRangeSq := ratApproachMinRange * ratApproachMinRange
			if distance <= minRangeSq {
				return ai.Success
			}

			// Apply velocity toward target
			if dx > 0 {
				physics.Velocity.X += ratSpeed * dt
			} else {
				physics.Velocity.X += -ratSpeed * dt
			}

			return ai.Running
		},
	}
}

func newRatJumpAttack() *ai.Action {
	// State variables (like the legacy action)
	jumped := false
	lifted := false
	prevMaxVelocity := 0.0
	jumpFrame := 0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			// Reset state
			jumped = false
			lifted = false

			rat := ecs.GetComponent[components.Rat](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			if rat == nil || anim == nil || physics == nil {
				return
			}

			// Save previous max velocity
			prevMaxVelocity = physics.MaxVelocity.X
			if prevMaxVelocity == 0 {
				cfg := ecs.Resource[config.Config](world)
				if cfg != nil {
					prevMaxVelocity = cfg.Body.MaxX
				}
			}

			// Calculate jump trigger frame
			jumpFrame = ratJumpTriggerFrame(anim)

			// Start attack animation
			animation.SetAnimationState(anim, "Attack")

			// Pause rat during attack
			animation.SetStateEffect(anim, func() func() {
				rat.Paused = true
				return func() { rat.Paused = false }
			}, "Attack")

			// Register hitbox slice for damage
			ratRegisterAttackSlice(world, eid, rat, facing, anim)

			// Register frame callback for jump trigger
			animation.RegisterFrameCallback(anim, jumpFrame, func() {
				phys := ecs.GetComponent[components.Physics](world, eid)
				face := ecs.GetComponent[components.Facing](world, eid)
				if phys == nil {
					return
				}

				jumped = true
				phys.MaxVelocity.X = ratMaxSpeed * 2
				phys.Velocity.Y = -ratSpeed
				phys.Grounded = false

				if face != nil && face.FlipX {
					phys.Velocity.X += ratMaxSpeed * 2
				} else {
					phys.Velocity.X -= ratMaxSpeed * 2
				}
			})
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			rat := ecs.GetComponent[components.Rat](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			// Check if attack animation ended
			if anim == nil || anim.State != "Attack" {
				return ai.Success
			}

			if physics == nil {
				return ai.Failure
			}

			// Track when rat lifts off ground
			if !physics.Grounded {
				lifted = true
			}

			// Wait for both jump and liftoff before applying movement
			if !jumped || !lifted {
				return ai.Running
			}

			// Apply movement while in air
			if rat != nil && !rat.Paused {
				move := ratSpeed * dt
				if facing != nil && facing.FlipX {
					move = -move
				}
				physics.Velocity.X += move
			}

			// Check if landed and can transition to idle
			if physics.Grounded && anim.Data != nil && animation.HasState(anim, components.IdleTag) {
				animation.SetAnimationState(anim, components.IdleTag)
				return ai.Success
			}

			return ai.Running
		},
		OnEnd: func(world *ecs.World, eid entities.EntityId) {
			// Restore previous max velocity
			physics := ecs.GetComponent[components.Physics](world, eid)
			if physics != nil {
				physics.MaxVelocity.X = prevMaxVelocity
			}

			// Reset state
			jumped = false
			lifted = false
		},
	}
}
