package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/pkg/config"
	"game/systems/update/entities/animation"
	"game/systems/update/combat"
)

const (
	ratDieDuration      = 1.0
	ratFlashDuration    = 1.0
	ratApproachMinRange = 20.0
	ratMaxTargetRange   = 30.0
	ratRangeAdjustment  = 10.0
	ratSpeed            = 60.0
	ratMaxSpeed         = 40.0
	ratAttackDamage     = 15.0
	ratAttackReactForce = 10.0
	ratAttackPushForce  = 10.0
)

type ratState interface {
	QueueRemoval(entities.EntityId)
}

func UpdateRat(world *ecs.World, state ratState, dt float64) {
	if world == nil || state == nil || dt == 0 {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		return
	}

	for _, eid := range world.EntitiesWith((*components.Rat)(nil)) {
		rat := ecs.GetComponent[components.Rat](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		facing := ecs.GetComponent[components.Facing](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		hitbox := ecs.GetComponent[components.Hitbox](world, eid)
		ai := ecs.GetComponent[components.AI](world, eid)
		visualEffects := ecs.GetComponent[components.VisualEffects](world, eid)
		deathState := ecs.GetComponent[components.DeathState](world, eid)

		if rat == nil || health == nil || facing == nil || anim == nil || hitbox == nil || ai == nil {
			continue
		}

		// Initialize AI behavior tree when target is first detected
		if ai.TargetID != 0 && ai.BehaviorTree == nil {
			ai.BehaviorTree = buildRatBehaviorTree()
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
		updateRat(ctx, rat, state)
	}
}

func updateRat(ctx *EnemyUpdateContext, rat *components.Rat, state ratState) {
	if rat == nil || ctx.Health == nil || ctx.Facing == nil {
		return
	}

	// Use generic enemy update (handles death, flash, AI, animations)
	if UpdateGenericEnemy(ctx.World, state, ctx.EID, &rat.Paused, rat.RemovalTarget, ctx.DT) {
		return // Enemy is dead
	}

	// Rat-specific logic: None currently
	// Rat's unique behavior (jump attack) is handled in AI action system
}

// ratJumpTriggerFrame calculates the frame when the rat should trigger its jump.
// Used by both BT and legacy systems.
func ratJumpTriggerFrame(anim *components.Animation) int {
	if anim == nil || anim.Data == nil {
		return 0
	}
	jump := anim.Data.Animation("Jump")
	attack := anim.Data.Animation("Attack")
	if jump == nil || attack == nil {
		return 0
	}
	if off := jump.From - attack.From; off > 0 {
		return off
	}
	return 0
}

// ratRegisterAttackSlice registers the hitbox slice callback for the rat's attack.
// Used by both BT and legacy systems.
func ratRegisterAttackSlice(world *ecs.World, eid entities.EntityId, rat *components.Rat, facing *components.Facing, anim *components.Animation) {
	if rat == nil || anim == nil || world == nil || facing == nil {
		return
	}
	// Pure ECS: Fetch hitbox component internally
	hitbox := ecs.GetComponent[components.Hitbox](world, eid)
	var contactedPrev []*components.Hitbox
	animation.RegisterSliceCallback(anim, components.HitboxSliceName, facing.FlipX, false, func(x, y, w, h float64, firstFrame bool) {
		if firstFrame {
			contactedPrev = contactedPrev[:0]
		}
		rect := bump.Rect{X: x, Y: y, W: w, H: h}
		contact, contacted := combat.ResolveHitboxArea(world, eid, rect, ratAttackDamage, BuildAttackFilters(hitbox, contactedPrev))
		contactedPrev = CollectUniqueContacts(contactedPrev, contacted)
		if len(contacted) > 0 {
			ApplyMeleeImpulse(world, eid, facing, contact, ratAttackPushForce, ratAttackReactForce)
		}
	})
}

func ratSetVelocityY(world *ecs.World, eid entities.EntityId, value float64) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	physics.Velocity.Y = value
}

func ratProcessJumpKinematics(world *ecs.World, eid entities.EntityId, rat *components.Rat, facing *components.Facing, anim *components.Animation, dt float64, jumped, lifted *bool) bool {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return true
	}
	if !physics.Grounded {
		*lifted = true
	}
	if !*jumped || !*lifted {
		return false
	}
	if rat != nil && !rat.Paused {
		move := ratSpeed * dt
		if facing != nil && facing.FlipX {
			move = -move
		}
		physics.Velocity.X += move
	}
	if rat != nil && anim != nil && anim.Data != nil && physics.Grounded && animation.HasState(anim, components.IdleTag) {
		animation.SetAnimationState(anim, components.IdleTag)
		return true
	}
	return false
}
