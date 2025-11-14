package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
)

const (
	crawlerDieDuration      = 1.0
	crawlerFlashDuration    = 0.05
	crawlerSpeed            = 100.0
	crawlerAttackDamage     = 15.0
	crawlerAttackReact      = 10.0
	crawlerAttackPush       = 10.0
	crawlerBackOffTime      = 1.5
	crawlerWaitTime         = 0.6
	crawlerApproachMinRange = 20.0 // Minimum target range for approach behavior
	crawlerWeight           = 0.8  // Weight for gravity/knockback
)

type crawlerState interface {
	QueueRemoval(entities.EntityId)
}

func UpdateCrawler(world *ecs.World, state crawlerState, dt float64) {
	if world == nil || state == nil || dt == 0 {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		return
	}

	for _, eid := range world.EntitiesWith((*components.Crawler)(nil)) {
		crawler := ecs.GetComponent[components.Crawler](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		facing := ecs.GetComponent[components.Facing](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		ai := ecs.GetComponent[components.AI](world, eid)
		hitbox := ecs.GetComponent[components.Hitbox](world, eid)
		visualEffects := ecs.GetComponent[components.VisualEffects](world, eid)
		deathState := ecs.GetComponent[components.DeathState](world, eid)
		if crawler == nil || anim == nil || hitbox == nil || health == nil || facing == nil {
			continue
		}

		// Initialize crawler-specific kinematics (one-time setup)
		crawlerInitKinematics(world, eid)

		// Initialize AI behavior tree when target is first detected
		if ai != nil && ai.TargetID != 0 && ai.BehaviorTree == nil {
			ai.BehaviorTree = buildCrawlerBehaviorTree()
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
		updateCrawler(ctx, crawler, state)
	}
}

func crawlerInitKinematics(world *ecs.World, eid entities.EntityId) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics != nil {
		physics.Weight = crawlerWeight
	}
}

func updateCrawler(ctx *EnemyUpdateContext, crawler *components.Crawler, state crawlerState) {
	if crawler == nil || ctx.Facing == nil {
		return
	}

	// Use generic enemy update (handles death, flash, AI, animations)
	if UpdateGenericEnemy(ctx.World, state, ctx.EID, &crawler.Paused, crawler.RemovalTarget, ctx.DT) {
		return // Enemy is dead
	}

	// Crawler-specific logic: None currently
	// Crawler's unique behavior (wall mode/patrol) is handled in AI action system
}

func crawlerDirectionTowardsTarget(world *ecs.World, eid entities.EntityId, ai *components.AI, facing *components.Facing) float64 {
	if ai == nil {
		return 1.0
	}
	transform := ecs.GetComponent[components.Transform](world, eid)
	if transform == nil {
		return 0
	}
	// Get target's Transform via ECS
	targetTransform := ecs.GetComponent[components.Transform](world, ai.TargetID)
	if targetTransform == nil {
		return 0
	}
	x, _, w, _ := transform.Rect()
	mid := x + w/2
	targetMid := targetTransform.X + targetTransform.W/2
	dir := -1.0
	if targetMid > mid {
		dir = 1.0
	}
	if facing != nil {
		facing.FlipX = targetMid > mid
	}
	return dir
}
