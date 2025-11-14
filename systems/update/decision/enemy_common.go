package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/prefabs"
	"game/systems/update/entities/animation"
)

// EnemyUpdateContext groups common parameters for enemy update functions.
// This reduces parameter lists from 11 parameters to 3 parameters,
// improving readability and making calls easier to understand.
//
// Usage:
//
//	ctx := &EnemyUpdateContext{
//	    World:         world,
//	    EID:           eid,
//	    DT:            dt,
//	    Health:        health,
//	    Facing:        facing,
//	    Animation:     anim,
//	    AI:            ai,
//	    VisualEffects: visualEffects,
//	    DeathState:    deathState,
//	}
//	updateCrawler(ctx, crawler, state)
type EnemyUpdateContext struct {
	World *ecs.World
	EID   entities.EntityId
	DT    float64

	// Component pointers
	Health        *components.Health
	Facing        *components.Facing
	Animation     *components.Animation
	AI            *components.AI
	VisualEffects *components.VisualEffects
	DeathState    *components.DeathState
	Physics       *components.Physics // Optional: some enemies need this (e.g., knight)
}

// HandleDeath processes enemy death: pause entity, stop animation, spawn particles on removal.
// Pure ECS: Takes explicit component parameters instead of EnemyCommon wrapper.
// Returns true if entity is dead/dying, false if still alive.
//
// NOTE: Death timer countdown is now handled by UpdateDeathState system.
// This function only handles: pausing, animation stopping, and entity removal.
func HandleDeath(
	world *ecs.World,
	state interface{ QueueRemoval(entities.EntityId) },
	eid entities.EntityId,
	health *components.Health,
	deathState *components.DeathState,
	anim *components.Animation,
	removalTarget entities.EntityId,
	pausedPtr *bool,
	dt float64,
) bool {
	if deathState == nil || health == nil {
		return false
	}

	isDead := health.Current <= 0

	// Not dead yet - ensure timer is reset
	if !isDead {
		if deathState.DieTimer <= 0 {
			deathState.DieTimer = deathState.DieDuration
		}
		return false
	}

	// Pause entity, stop animation, and disable physics (only do this once)
	if pausedPtr != nil && !*pausedPtr {
		*pausedPtr = true

		// Disable gravity so dead entities don't fall through the ground
		physics := ecs.GetComponent[components.Physics](world, eid)
		if physics != nil {
			physics.GravityEnabled = false
			physics.Velocity.X = 0
			physics.Velocity.Y = 0
		}
	}
	if anim != nil && anim.Data != nil && anim.Data.PlaySpeed != 0 {
		anim.Data.PlaySpeed = 0
	}

	// Death timer is now managed by UpdateDeathState system
	// We only check if removal is needed (timer reached -9999)
	if deathState.DieTimer == -9999 {
		RemoveEnemy(world, state, eid, removalTarget)
		return true
	}

	return deathState.DieTimer < deathState.DieDuration
}

// RemoveEnemy handles the full entity removal process.
// Pure ECS: Takes explicit parameters instead of EnemyCommon wrapper.
func RemoveEnemy(world *ecs.World, state interface{ QueueRemoval(entities.EntityId) }, eid entities.EntityId, removalTarget entities.EntityId) {
	// Get transform and experience BEFORE destroying entity
	var x, y, w, h float64
	var exp int
	if world != nil {
		if transform := ecs.GetComponent[components.Transform](world, eid); transform != nil {
			x, y, w, h = transform.X, transform.Y, transform.W, transform.H
		}
		// Query Experience component (Pure ECS pattern)
		if expComp := ecs.GetComponent[components.Experience](world, eid); expComp != nil {
			exp = expComp.Points
		}
	}

	// Destroy ECS entity
	if world != nil {
		world.DestroyEntity(eid)
	}

	// Queue entity removal using EntityId
	if state != nil && removalTarget != 0 {
		state.QueueRemoval(removalTarget)
	}

	// Spawn exp particles using saved transform and experience data
	if w > 0 && h > 0 {
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
			// Center the flake on the entity that died
			centerX := x + w/2
			centerY := y + h/2

			// Create flake targeting player (Pure ECS: pass world explicitly)
			flakeID := prefabs.NewFlakePrefab(world, centerX, centerY, 0, playerID)
			if flakeID != 0 && world != nil {
				world.QueueInit(flakeID)
			}
		}
	}
}

// UpdateAI runs the AI state machine and updates animation facing direction.
// Pure ECS: Takes explicit component parameters.
func UpdateAI(world *ecs.World, eid entities.EntityId, ai *components.AI, anim *components.Animation, transform *components.Transform, dt float64) {
	if ai == nil {
		return
	}

	// Clean up dead targets
	PruneDeadTarget(world, ai)

	// Tick behavior tree
	if ai.BehaviorTree != nil {
		ai.BehaviorTree.Tick(world, eid, dt)
	}

	// Update facing direction based on target
	if ai.TargetID != 0 && anim != nil && transform != nil {
		// Query target transform via ECS
		targetTransform := GetTargetTransform(world, ai)
		if targetTransform != nil {
			tx, tw := targetTransform.X, targetTransform.W
			x, w := transform.X, transform.W
			// Update Facing component instead of Anim.FlipX
			facing := ecs.GetComponent[components.Facing](world, eid)
			if facing != nil {
				facing.FlipX = tx+tw/2 > x+w/2
			}
		}
	}
}

// UpdateAirGroundState transitions animation between Jump and ground states based on grounded status.
// Pure ECS: Takes explicit component parameters.
func UpdateAirGroundState(physics *components.Physics, anim *components.Animation) {
	if physics == nil || anim == nil {
		return
	}

	if anim.Data == nil {
		return
	}

	// Transition to Jump when leaving ground
	if !physics.Grounded && (anim.State == "Idle" || anim.State == "Walk") {
		if animation.HasState(anim, "Jump") {
			animation.SetAnimationState(anim, "Jump")
		}
	}

	// Transition back to ground state when landing
	if physics.Grounded && anim.State == "Jump" {
		if animation.HasState(anim, "Block") {
			animation.SetAnimationState(anim, "Block")
		} else {
			animation.SetAnimationState(anim, "Idle")
		}
	}
}
