package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/systems/update/entities/animation"
)

const (
	batDieDuration   = 1.0
	batFlashDuration = 1.9
	batSpeed         = 60.0
	batMaxSpeed      = 40.0
	batAttackDamage  = 15.0
	batAttackReact   = 10.0
	batAttackPush    = 10.0
	batViewHeight    = 80.0
	batAttackTime    = 0.7
	batBackOffFactor = 0.25
	batRangeClose    = 20.0
	batRangeMid      = 25.0
	batRangeFar      = 28.0
)

type batState interface {
	QueueRemoval(entities.EntityId)
}

func UpdateBat(world *ecs.World, state batState, dt float64) {
	if world == nil || state == nil || dt == 0 {
		return
	}
	for _, eid := range world.EntitiesWith((*components.Bat)(nil)) {
		bat := ecs.GetComponent[components.Bat](world, eid)
		hitbox := ecs.GetComponent[components.Hitbox](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		facing := ecs.GetComponent[components.Facing](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		ai := ecs.GetComponent[components.AI](world, eid)
		visualEffects := ecs.GetComponent[components.VisualEffects](world, eid)
		deathState := ecs.GetComponent[components.DeathState](world, eid)

		if bat == nil || anim == nil || hitbox == nil || health == nil || facing == nil || ai == nil {
			continue
		}

		// Initialize bat-specific kinematics (one-time setup)
		batInitKinematics(world, eid)

		// Initialize AI behavior tree when target is first detected
		if ai.TargetID != 0 && ai.BehaviorTree == nil {
			ai.BehaviorTree = buildBatBehaviorTree()
		}

		ctx := &EnemyUpdateContext{
			World:         world,
			EID:           eid,
			DT:            dt,
			Health:        health,
			Facing:        facing,
			Animation:     anim,
			AI:            ai,
			VisualEffects: visualEffects,
			DeathState:    deathState,
		}
		updateBat(ctx, bat, state)
	}
}

func batInitKinematics(world *ecs.World, eid entities.EntityId) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	// Only initialize once (when weight is still non-zero from prefab defaults)
	// This prevents resetting velocity every frame, which would break BT movement
	if physics.Weight == 0 {
		return
	}
	physics.Weight = 0
	physics.Velocity.X = 0
	physics.Velocity.Y = 0
}

func updateBat(ctx *EnemyUpdateContext, bat *components.Bat, state batState) {
	if bat == nil || ctx.Facing == nil {
		return
	}

	// Use generic enemy update (handles death, flash)
	if UpdateGenericEnemy(ctx.World, state, ctx.EID, &bat.Paused, bat.RemovalTarget, ctx.DT) {
		return // Enemy is dead
	}

	// Bat-specific logic: Custom animation (uses AI target instead of velocity)
	batUpdateAnimation(ctx.Animation, ctx.AI)
}

func batUpdateAnimation(anim *components.Animation, ai *components.AI) {
	if anim == nil || ai == nil {
		return
	}
	if ai.TargetID != 0 {
		if anim.State == components.IdleTag {
			animation.SetAnimationState(anim, components.WalkTag)
		}
	} else if anim.State == components.WalkTag {
		animation.SetAnimationState(anim, components.IdleTag)
	}
}
