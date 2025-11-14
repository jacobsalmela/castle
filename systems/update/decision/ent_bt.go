package decision

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/systems/update/decision/nodes"
)

// buildEntBehaviorTree creates the behavior tree for ent enemies.
// This replicates the legacy action queue behavior using the new BT system.
func buildEntBehaviorTree() ai.Node {
	// Pre-wrap actions with human-friendly names for debug visualization
	idle := ai.WrapForDebug(nodes.Idle(), "Idle", 2)
	approach := ai.WrapForDebug(newEntApproach(), "Approach", 2)
	pauseBrief := ai.WrapForDebug(nodes.Wait(0.1), "Wait(0.1s)", 2)

	// Action choices with descriptive names
	// Attack sequence includes cooldown check, attack, and post-attack pause
	attackSequence := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newEntAttackWithCooldown(), "AttackWithCooldown", 4),
			ai.WrapForDebug(nodes.Wait(0.8), "PostAttackPause(0.8s)", 4),
		},
	}, "AttackSequence", 3)

	backup := ai.WrapForDebug(&ai.Timeout{
		Duration: 1.0,
		Child:    newEntBackup(),
	}, "Backup(1.0s)", 3)

	idlePause := ai.WrapForDebug(nodes.Wait(0.5), "IdlePause(0.5s)", 3)

	// Random action selector with weights matching legacy system
	randomAction := ai.WrapForDebug(nodes.NewRandomSelector(
		nodes.WeightedChoice{Weight: 2.0, Node: attackSequence}, // ~57% chance
		nodes.WeightedChoice{Weight: 0.5, Node: backup},         // ~14% chance
		nodes.WeightedChoice{Weight: 1.0, Node: idlePause},      // ~29% chance
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
	return ai.WrapForDebug(tree, "EntBT", 0)
}

// newEntApproach creates a custom approach action for the ent.
// This moves the ent toward its target using physics velocity until within range.
func newEntApproach() *ai.Action {
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
			minRangeSq := entApproachMinRange * entApproachMinRange
			if distance <= minRangeSq {
				return ai.Success
			}

			// Apply velocity toward target
			if dx > 0 {
				physics.Velocity.X += entApproachSpeed * dt
			} else {
				physics.Velocity.X += -entApproachSpeed * dt
			}

			return ai.Running
		},
	}
}

// newEntBackup creates a custom backup action for the ent.
// This moves the ent away from its target using physics velocity.
func newEntBackup() *ai.Action {
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
			maxRangeSq := entBackRange * entBackRange
			if distance >= maxRangeSq {
				return ai.Success
			}

			// Apply velocity away from target (opposite direction)
			if dx > 0 {
				physics.Velocity.X += -entApproachSpeed * dt
			} else {
				physics.Velocity.X += entApproachSpeed * dt
			}

			return ai.Running
		},
	}
}

// newEntAttackWithCooldown creates an attack action that respects attack cooldown.
// This is a Selector that either waits for cooldown or performs the attack.
func newEntAttackWithCooldown() ai.Node {
	// Selector: Try cooldown check first, if it succeeds (cooldown active), wait
	// Otherwise, fall through to perform attack
	return &ai.Selector{
		Children: []ai.Node{
			newEntCooldownWait(), // If on cooldown, wait
			newEntMeleeAttack(),  // Otherwise, attack
		},
	}
}

// newEntCooldownWait creates an action that waits for attack cooldown to expire.
// Returns Success if cooldown is active (to trigger wait), Failure if ready to attack.
func newEntCooldownWait() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			ent := ecs.GetComponent[components.Ent](world, eid)
			if ent == nil {
				return ai.Failure
			}

			// If cooldown is active, keep running (wait for it)
			if ent.AttackCooldown > 0 {
				return ai.Running
			}

			// Cooldown expired, signal failure so Selector moves to attack
			return ai.Failure
		},
	}
}

// newEntMeleeAttack creates a melee attack action for the ent.
// This triggers the attack animation and handles hitbox registration via generic system.
func newEntMeleeAttack() *ai.Action {
	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			ent := ecs.GetComponent[components.Ent](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)
			melee := ecs.GetComponent[components.MeleeAttackBehavior](world, eid)

			if ent == nil || facing == nil || anim == nil || aiComp == nil || melee == nil {
				return
			}

			// Set attack cooldown
			ent.AttackCooldown = entAttackCooldown

			// Use the generic melee attack system to set up the attack
			// This handles animation, pausing, and hitbox registration
			EnterMeleeAttack(world, eid, melee, facing, anim, aiComp, &ent.Paused)
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
