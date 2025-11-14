package decision

import (
	"image/color"
	"math"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/prefabs"
	"game/systems/update/entities/animation"
)

const (
	knightDieDuration      = 1.0
	knightFlashDuration    = 0.05
	knightApproachSpeed    = 100.0
	knightBackSpeed        = 80.0
	knightApproachMinRange = 20.0 // Minimum target range for approach behavior
)

type knightState interface {
	QueueRemoval(entities.EntityId)
}

func UpdateKnight(world *ecs.World, state knightState, dt float64) {
	if world == nil || state == nil || dt == 0 {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		return
	}

	for _, eid := range world.EntitiesWith((*components.Knight)(nil)) {
		knight := ecs.GetComponent[components.Knight](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		physics := ecs.GetComponent[components.Physics](world, eid)
		facing := ecs.GetComponent[components.Facing](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		ai := ecs.GetComponent[components.AI](world, eid)
		visualEffects := ecs.GetComponent[components.VisualEffects](world, eid)
		deathState := ecs.GetComponent[components.DeathState](world, eid)

		if knight == nil || physics == nil || health == nil || facing == nil || anim == nil || ai == nil {
			continue
		}

		// Initialize knight-specific state (one-time setup)
		knightInitTimers(knight)
		knightEnsureTint(anim)
		knightEnsureKinematics(world, eid, cfg)

		// Initialize AI behavior tree when target is first detected
		if ai != nil && ai.TargetID != 0 && ai.BehaviorTree == nil {
			ai.BehaviorTree = buildKnightBehaviorTree(world, eid, cfg)
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
			Physics:       physics,
		}
		updateKnight(ctx, knight, state)
	}
}

func knightInitTimers(knight *components.Knight) {
	// Flash and death timers now managed by shared VisualEffects and DeathState components
	// This function kept as no-op for compatibility during migration
}

func knightEnsureTint(anim *components.Animation) {
	if anim == nil {
		return
	}
	knightTintRed(anim)
}

func knightEnsureKinematics(world *ecs.World, eid entities.EntityId, cfg *config.Config) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics.MaxVelocity.X == 0 {
		physics.MaxVelocity.X = cfg.Body.MaxX
	}
}

func updateKnight(ctx *EnemyUpdateContext, knight *components.Knight, state knightState) {
	if knight == nil || ctx.Physics == nil || ctx.Health == nil {
		return
	}

	// Knight has special death handling (shield stop), so use custom function instead of generic
	if knightHandleDeath(ctx.World, state, ctx.EID, knight, ctx.Health, ctx.Animation, ctx.DeathState, ctx.DT) {
		return // Enemy is dead
	}

	// Update flash effect (Pure ECS: inline Update)
	if ctx.VisualEffects != nil {
		if ctx.VisualEffects.FlashTimer > 0 {
			ctx.VisualEffects.FlashTimer -= ctx.DT
			if ctx.VisualEffects.FlashTimer < 0 {
				ctx.VisualEffects.FlashTimer = 0
			}
		}
	}
	ApplyFlashColorToAnimation(ctx.Animation, ctx.VisualEffects, ctx.DeathState, ctx.Health)

	// Update AI
	knightUpdateAI(ctx.World, ctx.EID, knight, ctx.Facing, ctx.AI, ctx.DT)

	// Knight-specific logic
	knightUpdateSecondPhase(ctx.World, ctx.EID, knight, ctx.Health, ctx.AI, ctx.Animation)
	knightUpdateDashCooldown(knight, ctx.DT)
	knightUpdateAnimation(knight, ctx.Physics, ctx.Animation)
}

func knightHandleDeath(world *ecs.World, state knightState, eid entities.EntityId, knight *components.Knight, health *components.Health, anim *components.Animation, deathState *components.DeathState, dt float64) bool {
	if knight == nil || health == nil || health.Current > 0 {
		return false
	}
	knight.Paused = true

	// Disable gravity so dead knight doesn't fall through the ground
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics != nil && physics.GravityEnabled {
		physics.GravityEnabled = false
		physics.Velocity.X = 0
		physics.Velocity.Y = 0
	}

	// Get hitbox for shield cleanup
	hitbox := ecs.GetComponent[components.Hitbox](world, eid)
	KnightStopShield(knight, anim, hitbox)
	if anim != nil && anim.Data != nil {
		anim.Data.PlaySpeed = 0
	}
	if deathState != nil && deathState.DieTimer > 0 {
		deathState.DieTimer -= dt
		if deathState.DieTimer < 0 {
			deathState.DieTimer = 0
		}
		if anim != nil {
			alpha := uint8(float64(math.MaxUint8) * deathState.DieTimer / knightDieDuration)
			anim.ColorScale = color.RGBA{alpha, alpha, alpha, alpha}
		}
		return true
	}
	if deathState != nil && deathState.DieTimer == 0 {
		knightRemove(world, state, eid, knight)
		deathState.DieTimer = -9999
	}
	return true
}

func knightRemove(world *ecs.World, state knightState, eid entities.EntityId, knight *components.Knight) {
	// Wave 5.2: Get transform and experience BEFORE destroying entity
	var w, h float64
	var particleX, particleY float64
	if transform := ecs.GetComponent[components.Transform](world, eid); transform != nil {
		particleX, particleY, w, h = transform.X, transform.Y, transform.W, transform.H
	}

	// Get Experience component for XP drop amount (before entity destruction)
	expComp := ecs.GetComponent[components.Experience](world, eid)

	if world != nil {
		world.DestroyEntity(eid)
	}

	// Queue removal using EntityId
	if state != nil {
		state.QueueRemoval(eid)
	}

	// Spawn exp particles using transform data - early returns to reduce nesting
	if w <= 0 || h <= 0 {
		return
	}
	if expComp == nil {
		return
	}

	exp := expComp.Points
	if exp < 0 {
		exp = 0
	}

	// Get player entity for flake targeting
	players := world.EntitiesWith((*components.Player)(nil))
	if len(players) == 0 {
		return
	}
	playerID := players[0]

	for i := 0; i < exp; i++ {
		// Center the flake on the knight
		centerX := particleX + w/2
		centerY := particleY + h/2

		// Create flake targeting player (Pure ECS: pass world explicitly)
		flakeID := prefabs.NewFlakePrefab(world, centerX, centerY, 0, playerID)
		if flakeID != 0 && world != nil {
			world.QueueInit(flakeID)
		}
	}
}

func knightUpdateDashCooldown(knight *components.Knight, dt float64) {
	if knight == nil || knight.DashCooldown <= 0 {
		return
	}
	knight.DashCooldown -= dt
	if knight.DashCooldown < 0 {
		knight.DashCooldown = 0
	}
}

func knightUpdateAI(world *ecs.World, eid entities.EntityId, knight *components.Knight, facing *components.Facing, ai *components.AI, dt float64) {
	if knight == nil || ai == nil {
		return
	}
	// Process AI logic using system functions
	PruneDeadTarget(world, ai)
	if ai.BehaviorTree != nil {
		ai.BehaviorTree.Tick(world, eid, dt)
	}
	KnightEnsureFacingTarget(world, eid, knight, facing)
}

func knightUpdateAnimation(knight *components.Knight, physics *components.Physics, anim *components.Animation) {
	if knight == nil || anim == nil {
		return
	}
	if knight.ShieldActive {
		return
	}
	velocity := components.Vec2{}
	if physics != nil {
		velocity = physics.Velocity
	}
	moving := math.Abs(velocity.X) > 0.1
	if moving && anim.State == components.IdleTag {
		animation.SetAnimationState(anim, components.WalkTag)
	} else if !moving && anim.State == components.WalkTag {
		animation.SetAnimationState(anim, components.IdleTag)
	}
}

func knightUpdateSecondPhase(world *ecs.World, eid entities.EntityId, knight *components.Knight, health *components.Health, ai *components.AI, anim *components.Animation) {
	if knight == nil || health == nil {
		return
	}
	threshold := health.Max * 0.8
	if knight.SecondPhase || health.Current > threshold {
		return
	}
	knight.SecondPhase = true
	if anim != nil && anim.Data != nil {
		anim.Data.PlaySpeed = 1.5
	}
	knight.DashCooldown = 0
}
