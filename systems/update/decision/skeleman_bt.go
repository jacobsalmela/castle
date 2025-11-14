package decision

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/systems/update/entities/animation"
	"game/systems/update/combat"
	"game/systems/update/decision/nodes"
)

// buildSkelemanBehaviorTree creates the behavior tree for skeleman enemies.
// This replicates the legacy action queue behavior using the new BT system.
func buildSkelemanBehaviorTree() ai.Node {
	// Pre-wrap actions with human-friendly names for debug visualization
	idle := ai.WrapForDebug(nodes.Idle(), "Idle", 2)
	approach := ai.WrapForDebug(newSkelemanApproach(), "Approach", 2)
	pauseBrief := ai.WrapForDebug(nodes.Wait(0.1), "Wait(0.1s)", 2)

	// Action choices with descriptive names
	attackShort := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newSkelemanAttackShort(), "AttackShort", 4),
			ai.WrapForDebug(nodes.Wait(0.5), "Recovery(0.5s)", 4),
		},
	}, "AttackShortSequence", 3)

	attackLong := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newSkelemanAttackLong(), "AttackLong", 4),
			ai.WrapForDebug(nodes.Wait(0.5), "Recovery(0.5s)", 4),
		},
	}, "AttackLongSequence", 3)

	jumpAttack := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newSkelemanJumpAttack(), "JumpAttack", 4),
			ai.WrapForDebug(nodes.Wait(0.5), "Recovery(0.5s)", 4),
		},
	}, "JumpAttackSequence", 3)

	backup := ai.WrapForDebug(&ai.Timeout{
		Duration: 1.0,
		Child:    newSkelemanBackup(),
	}, "Backup(1.0s)", 3)

	idlePause := ai.WrapForDebug(nodes.Wait(0.8), "IdlePause(0.8s)", 3)

	// Random action selector with weights matching legacy system
	randomAction := ai.WrapForDebug(nodes.NewRandomSelector(
		nodes.WeightedChoice{Weight: 2.0, Node: attackShort}, // ~31% chance
		nodes.WeightedChoice{Weight: 2.0, Node: attackLong},  // ~31% chance
		nodes.WeightedChoice{Weight: 1.0, Node: jumpAttack},  // ~15% chance
		nodes.WeightedChoice{Weight: 0.5, Node: backup},      // ~8% chance
		nodes.WeightedChoice{Weight: 1.0, Node: idlePause},   // ~15% chance
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
	return ai.WrapForDebug(tree, "SkelemanBT", 0)
}

// newSkelemanApproach creates a custom approach action for the skeleman.
// This moves the skeleman toward its target using physics velocity until within range.
func newSkelemanApproach() *ai.Action {
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
			minRangeSq := skeleApproachMinRange * skeleApproachMinRange
			if distance <= minRangeSq {
				return ai.Success
			}

			// Apply velocity toward target
			if dx > 0 {
				physics.Velocity.X += skeleSpeed * dt
			} else {
				physics.Velocity.X += -skeleSpeed * dt
			}

			return ai.Running
		},
	}
}

// newSkelemanBackup creates a custom backup action for the skeleman.
// This moves the skeleman away from its target using physics velocity.
func newSkelemanBackup() *ai.Action {
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
			maxRangeSq := skeleBackRange * skeleBackRange
			if distance >= maxRangeSq {
				return ai.Success
			}

			// Apply velocity away from target (opposite direction)
			if dx > 0 {
				physics.Velocity.X += -skeleSpeed * dt
			} else {
				physics.Velocity.X += skeleSpeed * dt
			}

			return ai.Running
		},
	}
}

// newSkelemanAttackShort creates a short attack action for the skeleman.
// This triggers the AttackShort animation and handles hitbox registration.
func newSkelemanAttackShort() *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			skeleman := ecs.GetComponent[components.Skeleman](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)

			if skeleman == nil || anim == nil || facing == nil || aiComp == nil {
				return
			}

			// Start attack animation
			animation.SetAnimationState(anim, "AttackShort")

			// Pause skeleman during attack
			animation.SetStateEffect(anim, func() func() {
				skeleman.Paused = true
				return func() { skeleman.Paused = false }
			}, "AttackShort")

			// Register hitbox slice for damage
			skelemanRegisterAttackSlice(world, eid, facing, anim, skeleAttackDamage)

			// Stop movement and face target
			skeleSetVelocity(world, eid, 0, 0)
			skeleEnsureFacingTarget(world, eid, facing, aiComp)
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			anim := ecs.GetComponent[components.Animation](world, eid)

			// Check if attack animation ended
			if anim == nil || anim.State != "AttackShort" {
				return ai.Success
			}

			return ai.Running
		},
	}
}

// newSkelemanAttackLong creates a long attack action for the skeleman.
// This triggers the AttackLong animation and handles hitbox registration.
func newSkelemanAttackLong() *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			skeleman := ecs.GetComponent[components.Skeleman](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)

			if skeleman == nil || anim == nil || facing == nil || aiComp == nil {
				return
			}

			// Start attack animation
			animation.SetAnimationState(anim, "AttackLong")

			// Pause skeleman during attack
			animation.SetStateEffect(anim, func() func() {
				skeleman.Paused = true
				return func() { skeleman.Paused = false }
			}, "AttackLong")

			// Register hitbox slice for damage
			skelemanRegisterAttackSlice(world, eid, facing, anim, skeleAttackDamage)

			// Stop movement and face target
			skeleSetVelocity(world, eid, 0, 0)
			skeleEnsureFacingTarget(world, eid, facing, aiComp)
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			anim := ecs.GetComponent[components.Animation](world, eid)

			// Check if attack animation ended
			if anim == nil || anim.State != "AttackLong" {
				return ai.Success
			}

			return ai.Running
		},
	}
}

// newSkelemanJumpAttack creates a jump attack action for the skeleman.
// This triggers the AttackShort animation while jumping toward the target.
func newSkelemanJumpAttack() *ai.Action {
	prevMaxVelocityX := 0.0
	prevMaxVelocityY := 0.0
	speed := 0.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			skeleman := ecs.GetComponent[components.Skeleman](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			if skeleman == nil || anim == nil || facing == nil || aiComp == nil || physics == nil {
				return
			}

			// Save previous max velocity
			prevMaxVelocityX = physics.MaxVelocity.X
			prevMaxVelocityY = physics.MaxVelocity.Y
			if prevMaxVelocityX == 0 {
				prevMaxVelocityX = skeleMaxSpeed
			}
			if prevMaxVelocityY == 0 {
				prevMaxVelocityY = skeleMaxSpeed
			}

			// Start attack animation
			animation.SetAnimationState(anim, "AttackShort")

			// Pause skeleman during attack
			animation.SetStateEffect(anim, func() func() {
				skeleman.Paused = true
				return func() { skeleman.Paused = false }
			}, "AttackShort")

			// Register hitbox slice for damage
			skelemanRegisterAttackSlice(world, eid, facing, anim, skeleAttackDamage)

			// Stop movement and face target
			skeleSetVelocity(world, eid, 0, 0)
			skeleEnsureFacingTarget(world, eid, facing, aiComp)

			// Start jump
			speed = skeleStartJump(world, eid, skeleman, facing)
			physics.MaxVelocity.X = skeleMaxSpeed * 2
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			skeleman := ecs.GetComponent[components.Skeleman](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)

			// Update jump physics
			if skeleUpdateJump(world, eid, skeleman, facing, dt, speed, anim) {
				return ai.Success // Landed and animation finished
			}

			return ai.Running
		},
		OnEnd: func(world *ecs.World, eid entities.EntityId) {
			// Restore previous max velocity
			physics := ecs.GetComponent[components.Physics](world, eid)
			if physics != nil {
				physics.MaxVelocity.X = prevMaxVelocityX
				physics.MaxVelocity.Y = prevMaxVelocityY
			}
		},
	}
}

// skelemanRegisterAttackSlice registers the hitbox slice callback for the skeleman's attack.
// Used by both BT and legacy systems.
func skelemanRegisterAttackSlice(world *ecs.World, eid entities.EntityId, facing *components.Facing, anim *components.Animation, damage float64) {
	if anim == nil || world == nil || facing == nil {
		return
	}
	hitbox := ecs.GetComponent[components.Hitbox](world, eid)
	var contactedPrev []*components.Hitbox
	animation.RegisterSliceCallback(anim, components.HitboxSliceName, facing.FlipX, false, func(x, y, w, h float64, firstFrame bool) {
		if firstFrame {
			contactedPrev = contactedPrev[:0]
		}
		rect := bump.Rect{X: x, Y: y, W: w, H: h}
		contact, contacted := combat.ResolveHitboxArea(world, eid, rect, damage, skeleAttackFilters(hitbox, contactedPrev))
		contactedPrev = skeleCollectContacts(contactedPrev, contacted)
		if len(contacted) > 0 {
			skeleApplyAttackImpulse(world, eid, contact, facing)
		}
	})
}
