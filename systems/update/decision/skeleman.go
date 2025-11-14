package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/systems/update/entities/animation"
	"game/systems/update/combat"
)

const (
	skeleDieDuration      = 1.0
	skeleFlashDuration    = 0.05
	skeleSpeed            = 100.0
	skeleMaxSpeed         = 50.0
	skeleAttackDamage     = 18.0
	skeleAttackReact      = 10.0
	skeleAttackPush       = 10.0
	skeleBackRange        = 30.0
	skeleJumpAccel        = 100.0
	skeleApproachMinRange = 20.0 // Minimum target range for approach behavior
)

type skeleState interface {
	QueueRemoval(entities.EntityId)
}

func UpdateSkeleman(world *ecs.World, dt float64) {
	entities := world.EntitiesWith((*components.Skeleman)(nil), (*components.Physics)(nil))
	for _, eid := range entities {
		skele := ecs.GetComponent[components.Skeleman](world, eid)
		physics := ecs.GetComponent[components.Physics](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		facing := ecs.GetComponent[components.Facing](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		ai := ecs.GetComponent[components.AI](world, eid)
		visualEffects := ecs.GetComponent[components.VisualEffects](world, eid)
		deathState := ecs.GetComponent[components.DeathState](world, eid)

		// Initialize skeleman-specific kinematics (one-time setup)
		skeleInitKinematics(world, eid)

		// Initialize AI behavior tree when target is first detected
		if ai != nil && ai.TargetID != 0 && ai.BehaviorTree == nil {
			ai.BehaviorTree = buildSkelemanBehaviorTree()
		}

		updateSkeleman(world, eid, skele, physics, health, facing, anim, ai, visualEffects, deathState, dt)
	}
}

func skeleInitKinematics(world *ecs.World, eid entities.EntityId) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	physics.MaxVelocity.X = skeleMaxSpeed
}

func updateSkeleman(world *ecs.World, eid entities.EntityId, skele *components.Skeleman, physics *components.Physics, health *components.Health, facing *components.Facing, anim *components.Animation, ai *components.AI, visualEffects *components.VisualEffects, deathState *components.DeathState, dt float64) {
	if skele == nil || physics == nil {
		return
	}

	// Use generic enemy update (handles death, flash, AI, animations)
	// Note: Passing nil for state since skeleman doesn't use QueueRemoval
	if UpdateGenericEnemy(world, nil, eid, &skele.Paused, skele.RemovalTarget, dt) {
		return // Enemy is dead
	}

	// Skeleman-specific logic: None currently
	// Skeleman's unique behavior (combo attacks) is handled in AI action system
}

func skeleRegisterAttackSlice(world *ecs.World, eid entities.EntityId, facing *components.Facing, anim *components.Animation, damage float64) {
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

func skeleApplyAttackImpulse(world *ecs.World, eid entities.EntityId, contact components.ContactType, facing *components.Facing) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	force := skeleAttackPush
	if contact >= components.Block {
		force = skeleAttackReact
	}
	if facing != nil && facing.FlipX {
		force *= -1
	}
	physics.Velocity.X += force
}

func skeleAttackFilters(self *components.Hitbox, contacted []*components.Hitbox) []*components.Hitbox {
	count := len(contacted)
	if self != nil {
		count++
	}
	filters := make([]*components.Hitbox, 0, count)
	if self != nil {
		filters = append(filters, self)
	}
	filters = append(filters, contacted...)
	return filters
}

func skeleCollectContacts(existing []*components.Hitbox, contacted []*components.Hitbox) []*components.Hitbox {
	for _, target := range contacted {
		if target == nil {
			continue
		}
		seen := false
		for _, stored := range existing {
			if stored == target {
				seen = true
				break
			}
		}
		if !seen {
			existing = append(existing, target)
		}
	}
	return existing
}

func skeleStartJump(world *ecs.World, eid entities.EntityId, skele *components.Skeleman, facing *components.Facing) float64 {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return 0
	}
	speed := skeleJumpAccel

	physics.MaxVelocity.X = skeleMaxSpeed * 2
	physics.Velocity.Y = -skeleJumpAccel
	physics.Grounded = false
	if facing != nil && facing.FlipX {
		physics.Velocity.X += skeleMaxSpeed * 2
	} else {
		physics.Velocity.X -= skeleMaxSpeed * 2
		speed *= -1
	}
	return speed
}

func skeleUpdateJump(world *ecs.World, eid entities.EntityId, skele *components.Skeleman, facing *components.Facing, dt, speed float64, anim *components.Animation) bool {
	if skele == nil {
		return true
	}
	grounded := false
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return true
	}
	if !skele.Paused {
		physics.Velocity.X += speed * dt
	}
	grounded = physics.Grounded
	return grounded && (anim == nil || anim.State != "AttackShort")
}

func skeleSetVelocity(world *ecs.World, eid entities.EntityId, vx, vy float64) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	physics.Velocity.X = vx
	physics.Velocity.Y = vy
}

func skeleEnsureFacingTarget(world *ecs.World, eid entities.EntityId, facing *components.Facing, ai *components.AI) float64 {
	if world == nil || eid == 0 || ai == nil || ai.TargetID == 0 {
		return 0
	}
	t := ecs.GetComponent[components.Transform](world, eid)
	if t == nil {
		return 0
	}
	// Get target's Transform via ECS directly
	targetTransform := ecs.GetComponent[components.Transform](world, ai.TargetID)
	if targetTransform == nil {
		return 0
	}
	targetMid := targetTransform.X + targetTransform.W/2
	selfMid := t.X + t.W/2
	facingRight := targetMid > selfMid
	if facing != nil {
		facing.FlipX = facingRight
	}
	if facingRight {
		return 1
	}
	return -1
}
