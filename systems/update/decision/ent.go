package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

const (
	entDieDuration      = 1.0
	entFlashDuration    = 0.05
	entApproachSpeed    = 65.0
	entMaxSpeed         = 30.0
	entAttackDamage     = 40.0
	entAttackReactForce = 12.0
	entAttackPushForce  = 10.0
	entBackRange        = 24.0
	entAttackCooldown   = 0.8
	entApproachMinRange = 20.0 // Minimum target range for approach behavior
)

type entState interface {
	QueueRemoval(entities.EntityId)
}

func UpdateEnt(world *ecs.World, state entState, dt float64) {
	if world == nil || state == nil || dt == 0 {
		return
	}
	for _, eid := range world.EntitiesWith((*components.Ent)(nil)) {
		ent := ecs.GetComponent[components.Ent](world, eid)
		hitbox := ecs.GetComponent[components.Hitbox](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		facing := ecs.GetComponent[components.Facing](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		ai := ecs.GetComponent[components.AI](world, eid)
		visualEffects := ecs.GetComponent[components.VisualEffects](world, eid)
		deathState := ecs.GetComponent[components.DeathState](world, eid)

		if ent == nil || hitbox == nil || facing == nil || anim == nil || health == nil || ai == nil {
			continue
		}

		// Initialize ent-specific kinematics (one-time setup)
		entInitKinematics(world, eid)

		// Initialize AI behavior tree when target is first detected
		if ai != nil && ai.TargetID != 0 && ai.BehaviorTree == nil {
			ai.BehaviorTree = buildEntBehaviorTree()
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
		updateEnt(ctx, ent, state)
	}
}

func entInitKinematics(world *ecs.World, eid entities.EntityId) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	if physics.MaxVelocity.X == 0 {
		physics.MaxVelocity.X = entMaxSpeed
	}
}

func updateEnt(ctx *EnemyUpdateContext, ent *components.Ent, state entState) {
	if ent == nil || ctx.Facing == nil {
		return
	}

	// Use generic enemy update (handles death, flash, AI, animations)
	if UpdateGenericEnemy(ctx.World, state, ctx.EID, &ent.Paused, ctx.EID, ctx.DT) {
		return // Enemy is dead
	}

	// Ent-specific logic: Attack cooldown
	entUpdateCooldown(ent, ctx.DT)
}

func entUpdateCooldown(ent *components.Ent, dt float64) {
	if ent == nil || ent.AttackCooldown <= 0 {
		return
	}
	ent.AttackCooldown -= dt
	if ent.AttackCooldown < 0 {
		ent.AttackCooldown = 0
	}
}
