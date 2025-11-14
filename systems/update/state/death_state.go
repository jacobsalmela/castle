package state

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"image/color"
	"math"
)

// UpdateDeathState processes death animation timers for all entities with DeathState.
// This system handles:
//  1. Countdown of DieTimer
//  2. Fade-out alpha animation
//  3. Marking entities for removal (timer = -9999)
//
// This system runs in the State Updates phase, after combat but before rendering.
//
// The actual entity removal and flake spawning is handled by HandleDeath() in
// enemy_common.go, which checks for the -9999 marker and calls RemoveEnemy().
//
// Division of responsibility:
// - UpdateDeathState: Timer countdown, fade animation (generic, runs once per frame)
// - HandleDeath: Pausing, animation stopping, entity removal (per-enemy, has removalTarget)
func UpdateDeathState(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	entities := world.EntitiesWith(
		(*components.DeathState)(nil),
		(*components.Health)(nil),
	)

	for _, eid := range entities {
		deathState := ecs.GetComponent[components.DeathState](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)

		if deathState == nil || health == nil {
			continue
		}

		// Process death animation for dead entities
		if health.Current <= 0 {
			updateDeathAnimation(world, eid, deathState, dt)
		} else {
			resetDeathTimer(deathState)
		}
	}
}

// updateDeathAnimation handles the death fade animation and timer countdown.
func updateDeathAnimation(world *ecs.World, eid entities.EntityId, deathState *components.DeathState, dt float64) {
	if deathState.DieTimer > 0 {
		// Countdown timer
		deathState.DieTimer -= dt
		if deathState.DieTimer < 0 {
			deathState.DieTimer = 0
		}

		// Apply fade-out alpha
		applyDeathFade(world, eid, deathState)
	}

	// Mark for removal when timer reaches 0
	if deathState.DieTimer == 0 {
		deathState.DieTimer = -9999
	}
}

// applyDeathFade updates the animation alpha based on death timer progress.
func applyDeathFade(world *ecs.World, eid entities.EntityId, deathState *components.DeathState) {
	anim := ecs.GetComponent[components.Animation](world, eid)
	if anim == nil {
		return
	}

	fadeProgress := deathState.DieTimer / deathState.DieDuration
	alpha := uint8(float64(math.MaxUint8) * fadeProgress)
	anim.ColorScale = color.RGBA{alpha, alpha, alpha, alpha}
}

// resetDeathTimer resets the death timer for alive entities (e.g., revived).
func resetDeathTimer(deathState *components.DeathState) {
	if deathState.DieTimer < deathState.DieDuration {
		deathState.DieTimer = deathState.DieDuration
	}
}
