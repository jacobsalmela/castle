package decision

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/systems/update/decision/nodes"
)

// buildGhoulBehaviorTree creates the behavior tree for ghoul enemies.
// Ghouls have two modes: Poacher (ranged) and Aggressive (melee).
func buildGhoulBehaviorTree(world *ecs.World, eid entities.EntityId) ai.Node {
	ghoul := ecs.GetComponent[components.Ghoul](world, eid)
	if ghoul == nil {
		return buildGhoulAggressiveBehaviorTree()
	}

	// Select tree based on role
	if ghoul.Poacher {
		return buildGhoulPoacherBehaviorTree()
	}
	return buildGhoulAggressiveBehaviorTree()
}

// buildGhoulPoacherBehaviorTree creates the behavior tree for poacher ghouls.
// Poachers maintain distance and throw rocks at the player.
func buildGhoulPoacherBehaviorTree() ai.Node {
	// Pre-wrap actions with human-friendly names for debug visualization
	idle := ai.WrapForDebug(nodes.Idle(), "Idle", 2)

	// Poacher behavior: back up if has rocks, throw, recover
	// If out of rocks, wait defensively
	poacherSequence := ai.WrapForDebug(&ai.Selector{
		Children: []ai.Node{
			buildGhoulThrowWithBackup(), // Extracted to reduce nesting
			// No rocks - wait defensively
			ai.WrapForDebug(nodes.Wait(2.0), "OutOfRocks(2.0s)", 3),
		},
	}, "PoacherAction", 2)

	// Assemble the tree
	tree := &ai.Repeat{
		Count: 0, // Infinite loop
		Child: &ai.Sequence{
			Children: []ai.Node{
				idle,
				poacherSequence,
			},
		},
	}

	return ai.WrapForDebug(tree, "GhoulPoacherBT", 0)
}

// buildGhoulThrowWithBackup creates the throw sequence with conditional backup.
// Reduced from 8 levels of nesting to 3 levels.
func buildGhoulThrowWithBackup() ai.Node {
	return ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newGhoulHasRocksCondition(), "HasRocks?", 4),
			buildGhoulConditionalBackup(), // Extracted helper
			ai.WrapForDebug(newGhoulThrowRock(), "ThrowRock", 4),
			ai.WrapForDebug(nodes.Wait(1.5), "ThrowRecover(1.5s)", 4),
		},
	}, "ThrowSequence", 3)
}

// buildGhoulConditionalBackup creates backup behavior for first throw only.
// This backs up on first throw (3 rocks), skips backup on subsequent throws.
func buildGhoulConditionalBackup() ai.Node {
	backupOnFirstThrow := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newGhoulIsFirstThrowCondition(), "FirstThrow?", 6),
			ai.WrapForDebug(&ai.Timeout{
				Duration: 1.0,
				Child:    newGhoulBackup(),
			}, "Backup(1.0s)", 6),
		},
	}, "BackupOnFirstThrow", 5)

	return ai.WrapForDebug(&ai.Selector{
		Children: []ai.Node{
			backupOnFirstThrow,
			// Skip backup if not first throw
			ai.WrapForDebug(nodes.Idle(), "SkipBackup", 5),
		},
	}, "ConditionalBackup", 4)
}

// buildGhoulAggressiveBehaviorTree creates the behavior tree for aggressive ghouls.
// Aggressive ghouls approach and perform melee attacks.
func buildGhoulAggressiveBehaviorTree() ai.Node {
	// Pre-wrap actions with human-friendly names for debug visualization
	idle := ai.WrapForDebug(nodes.Idle(), "Idle", 2)
	approach := ai.WrapForDebug(newGhoulApproach(), "Approach", 2)
	pauseBrief := ai.WrapForDebug(nodes.Wait(0.3), "PreAttackPause(0.3s)", 2)

	// Jump attack sequence (lower probability)
	jumpAttackSequence := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(nodes.Wait(0.3), "PreJumpPause(0.3s)", 4),
			ai.WrapForDebug(newGhoulJumpAttack(), "JumpAttack", 4),
			ai.WrapForDebug(nodes.Wait(1.0), "PostJumpPause(1.0s)", 4),
		},
	}, "JumpAttackSequence", 3)

	// Melee attack choices
	attackShort := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newGhoulAttackShort(), "AttackShort", 5),
			ai.WrapForDebug(nodes.Wait(0.8), "Recovery(0.8s)", 5),
		},
	}, "AttackShortSequence", 4)

	attackLong := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newGhoulAttackLong(), "AttackLong", 5),
			ai.WrapForDebug(nodes.Wait(1.0), "Recovery(1.0s)", 5),
		},
	}, "AttackLongSequence", 4)

	backup := ai.WrapForDebug(&ai.Timeout{
		Duration: 1.0,
		Child:    newGhoulBackup(),
	}, "Backup(1.0s)", 4)

	idlePause := ai.WrapForDebug(nodes.Wait(0.5), "IdlePause(0.5s)", 4)

	// Nested choice: either short/long attack, backup, or idle
	meleeChoice := ai.WrapForDebug(nodes.NewRandomSelector(
		nodes.WeightedChoice{Weight: 2.0, Node: attackShort}, // ~36% chance
		nodes.WeightedChoice{Weight: 2.0, Node: attackLong},  // ~36% chance
		nodes.WeightedChoice{Weight: 0.5, Node: backup},      // ~9% chance
		nodes.WeightedChoice{Weight: 1.0, Node: idlePause},   // ~18% chance
	), "ChooseMeleeAction", 3)

	// Top-level choice: jump attack or melee
	randomAction := ai.WrapForDebug(nodes.NewRandomSelector(
		nodes.WeightedChoice{Weight: 0.2, Node: jumpAttackSequence}, // ~17% chance
		nodes.WeightedChoice{Weight: 1.0, Node: meleeChoice},        // ~83% chance
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

	return ai.WrapForDebug(tree, "GhoulAggressiveBT", 0)
}

// newGhoulApproach creates a custom approach action for the ghoul.
func newGhoulApproach() *ai.Action {
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
			minRangeSq := ghoulApproachMinRange * ghoulApproachMinRange
			if distance <= minRangeSq {
				return ai.Success
			}

			// Apply velocity toward target
			if dx > 0 {
				physics.Velocity.X += ghoulSpeed * dt
			} else {
				physics.Velocity.X += -ghoulSpeed * dt
			}

			return ai.Running
		},
	}
}

// newGhoulBackup creates a custom backup action for the ghoul.
func newGhoulBackup() *ai.Action {
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
			maxRangeSq := ghoulBackRange * ghoulBackRange
			if distance >= maxRangeSq {
				return ai.Success
			}

			// Apply velocity away from target (opposite direction)
			if dx > 0 {
				physics.Velocity.X += -ghoulSpeed * dt
			} else {
				physics.Velocity.X += ghoulSpeed * dt
			}

			return ai.Running
		},
	}
}

// newGhoulHasRocksCondition creates a condition that checks if ghoul has rocks.
func newGhoulHasRocksCondition() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			ghoul := ecs.GetComponent[components.Ghoul](world, eid)
			if ghoul == nil || ghoul.Rocks <= 0 {
				return ai.Failure
			}
			return ai.Success
		},
	}
}

// newGhoulIsFirstThrowCondition creates a condition that checks if this is the first throw.
// Returns Success if rocks >= 3 (first engagement), Failure otherwise.
func newGhoulIsFirstThrowCondition() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			ghoul := ecs.GetComponent[components.Ghoul](world, eid)
			if ghoul == nil || ghoul.Rocks < 3 {
				return ai.Failure
			}
			return ai.Success
		},
	}
}

// newGhoulThrowRock creates a rock throwing action for the ghoul.
func newGhoulThrowRock() *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			ghoul := ecs.GetComponent[components.Ghoul](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)

			if ghoul == nil || facing == nil || anim == nil || aiComp == nil {
				return
			}

			// Face target and start throw
			ghoulEnsureFacingTarget(world, eid, facing, aiComp)
			ghoulStartThrow(world, eid, ghoul, facing, anim)
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			anim := ecs.GetComponent[components.Animation](world, eid)

			// Check if throw animation ended
			if anim == nil || anim.State != "Throw" {
				return ai.Success
			}

			return ai.Running
		},
	}
}

// newGhoulAttackShort creates a short melee attack action for the ghoul.
func newGhoulAttackShort() *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			ghoul := ecs.GetComponent[components.Ghoul](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)

			if ghoul == nil || facing == nil || anim == nil {
				return
			}

			ghoulEnterAttack(world, eid, ghoul, facing, "AttackShort", anim)
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

// newGhoulAttackLong creates a long melee attack action for the ghoul.
func newGhoulAttackLong() *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			ghoul := ecs.GetComponent[components.Ghoul](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)

			if ghoul == nil || facing == nil || anim == nil {
				return
			}

			ghoulEnterAttack(world, eid, ghoul, facing, "AttackLong", anim)
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

// newGhoulJumpAttack creates a jump attack action for the ghoul.
func newGhoulJumpAttack() *ai.Action {
	prevMaxVelocityX := 0.0
	prevMaxVelocityY := 0.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			ghoul := ecs.GetComponent[components.Ghoul](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			if ghoul == nil || facing == nil || anim == nil || physics == nil {
				return
			}

			// Save previous max velocity
			prevMaxVelocityX = physics.MaxVelocity.X
			prevMaxVelocityY = physics.MaxVelocity.Y

			// Start attack and jump
			ghoulEnterAttack(world, eid, ghoul, facing, "AttackShort", anim)
			ghoulExecuteJump(world, eid, facing)
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)

			// Update jump physics
			if ghoulUpdateJump(world, eid, facing, anim, aiComp, dt) {
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
