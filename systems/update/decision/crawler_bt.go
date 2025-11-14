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

// buildCrawlerBehaviorTree creates the behavior tree for crawler enemies.
// This replicates the legacy action queue behavior using the new BT system.
func buildCrawlerBehaviorTree() ai.Node {
	// Pre-wrap actions with human-friendly names for debug visualization
	idle := ai.WrapForDebug(nodes.Idle(), "Idle", 2)
	approach := ai.WrapForDebug(newCrawlerApproach(), "Approach", 2)
	pauseBrief := ai.WrapForDebug(nodes.Wait(0.1), "Wait(0.1s)", 2)

	// Action choices with descriptive names
	attack := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newCrawlerAttack(), "MeleeAttack", 4),
			ai.WrapForDebug(nodes.Wait(0.5), "Recovery(0.5s)", 4),
		},
	}, "AttackSequence", 3)

	backup := ai.WrapForDebug(&ai.Timeout{
		Duration: crawlerBackOffTime,
		Child:    newCrawlerBackup(),
	}, "Backup(1.5s)", 3)

	idlePause := ai.WrapForDebug(nodes.Wait(crawlerWaitTime), "IdlePause(0.6s)", 3)

	// Random action selector
	randomAction := ai.WrapForDebug(nodes.NewRandomSelector(
		nodes.WeightedChoice{Weight: 1.0, Node: attack},    // 33% chance
		nodes.WeightedChoice{Weight: 1.0, Node: backup},    // 33% chance
		nodes.WeightedChoice{Weight: 1.0, Node: idlePause}, // 33% chance
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
	return ai.WrapForDebug(tree, "CrawlerBT", 0)
}

// newCrawlerApproach creates a custom approach action for the crawler.
// This moves the crawler toward its target using physics velocity until within range.
func newCrawlerApproach() *ai.Action {
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
			minRangeSq := crawlerApproachMinRange * crawlerApproachMinRange
			if distance <= minRangeSq {
				return ai.Success
			}

			// Apply velocity toward target
			if dx > 0 {
				physics.Velocity.X += crawlerSpeed * dt
			} else {
				physics.Velocity.X += -crawlerSpeed * dt
			}

			return ai.Running
		},
	}
}

// newCrawlerBackup creates a custom backup action for the crawler.
// This moves the crawler away from its target using physics velocity.
func newCrawlerBackup() *ai.Action {
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

			// Check if we're already far enough (40.0 max range from prefab)
			const maxBackupRange = 40.0
			dy := (transform.Y + transform.H/2) - (targetTransform.Y + targetTransform.H/2)
			distance := dx*dx + dy*dy
			maxRangeSq := maxBackupRange * maxBackupRange
			if distance >= maxRangeSq {
				return ai.Success
			}

			// Apply velocity away from target (opposite direction)
			if dx > 0 {
				physics.Velocity.X += -crawlerSpeed * dt
			} else {
				physics.Velocity.X += crawlerSpeed * dt
			}

			return ai.Running
		},
	}
}

// newCrawlerAttack creates a melee attack action for the crawler.
// This triggers the attack animation and handles hitbox registration.
func newCrawlerAttack() *ai.Action {
	elapsed := 0.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			elapsed = 0.0

			crawler := ecs.GetComponent[components.Crawler](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)

			if crawler == nil || anim == nil || facing == nil || aiComp == nil {
				return
			}

			// Start attack animation
			animation.SetAnimationState(anim, "Attack")

			// Pause crawler during attack
			animation.SetStateEffect(anim, func() func() {
				crawler.Paused = true
				return func() { crawler.Paused = false }
			}, "Attack")

			// Register hitbox slice for damage
			crawlerRegisterAttackSlice(world, eid, facing, anim)
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			elapsed += dt

			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)

			// Check if attack animation ended
			if anim == nil || anim.State != "Attack" {
				return ai.Success
			}

			// Update facing direction toward target during attack
			if facing != nil && aiComp != nil && aiComp.TargetID != 0 {
				crawlerDirectionTowardsTarget(world, eid, aiComp, facing)
			}

			// Attack completes when animation finishes
			// Animation system will transition back to Idle automatically
			return ai.Running
		},
	}
}

// crawlerRegisterAttackSlice registers the hitbox slice callback for the crawler's attack.
// Used by both BT and legacy systems.
func crawlerRegisterAttackSlice(world *ecs.World, eid entities.EntityId, facing *components.Facing, anim *components.Animation) {
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
		contact, contacted := combat.ResolveHitboxArea(world, eid, rect, crawlerAttackDamage, BuildAttackFilters(hitbox, contactedPrev))
		contactedPrev = CollectUniqueContacts(contactedPrev, contacted)
		if len(contacted) > 0 {
			ApplyMeleeImpulse(world, eid, facing, contact, crawlerAttackPush, crawlerAttackReact)
		}
	})
}
