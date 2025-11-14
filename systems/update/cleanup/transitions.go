package cleanup

import (
	"game/components"
	"game/ecs"
	"game/pkg/config"
	"game/resources"

	"github.com/hajimehoshi/ebiten/v2"
)

// RunTransitions manages game transitions (restart and death) via the TransitionManager resource.
// This Pure ECS system handles all transition logic during the cleanup phase.
func RunTransitions(world *ecs.World, dt float64, resetFunc func()) {
	tm := ecs.Resource[resources.TransitionManager](world)
	if tm == nil {
		return
	}

	handleRestartTransition(world, tm, dt, resetFunc)
	handleDeathTransition(world, tm, dt, resetFunc)
}

// handleRestartTransition checks for restart requests and updates the restart transition.
func handleRestartTransition(world *ecs.World, tm *resources.TransitionManager, dt float64, resetFunc func()) {
	// Start restart transition if requested
	signals := ecs.Resource[resources.GameSignals](world)
	if tm.GetActiveType() == resources.TransitionNone && signals != nil && signals.ConsumeReset() {
		tm.StartRestart(func() {
			// onComplete callback - defer reset until after frame completes
			// This prevents race conditions between Update and Draw
			if signals != nil && resetFunc != nil {
				signals.SetPendingReset(resetFunc)
			}
		})
	}

	// Update active restart transition
	if tm.GetActiveType() == resources.TransitionRestart {
		tm.UpdateRestart(dt)
	}
}

// handleDeathTransition checks for player death and updates the death transition.
func handleDeathTransition(world *ecs.World, tm *resources.TransitionManager, dt float64, resetFunc func()) {
	// Get player entity
	players := world.EntitiesWith((*components.Player)(nil))
	if world == nil || len(players) == 0 {
		return
	}
	playerID := players[0]

	// Get player health from Pure ECS
	health := ecs.GetComponent[components.Health](world, playerID)
	if health == nil {
		return
	}

	// Start death transition if player died
	if tm.GetActiveType() == resources.TransitionNone && health.Current <= 0 {
		// Disable gravity so player doesn't fall through the ground
		physics := ecs.GetComponent[components.Physics](world, playerID)
		if physics != nil {
			physics.GravityEnabled = false
			physics.Velocity.X = 0
			physics.Velocity.Y = 0
		}

		// Freeze world and slow down time via TimeControl resource
		if tc := ecs.Resource[resources.TimeControl](world); tc != nil {
			tc.RequestFreeze(0.5)
			tc.SetSpeed(0.5)
		}

		// Get action keys from player's Input component
		var actionKeys []ebiten.Key
		if input := ecs.GetComponent[components.Input](world, playerID); input != nil {
			actionKeys = input.KeyBindings[components.InputKeyAction]
		}
		if actionKeys == nil {
			actionKeys = []ebiten.Key{ebiten.KeyX, ebiten.KeyM} // Fallback
		}

		// Get config from world resource
		cfg := ecs.Resource[config.Config](world)

		tm.StartDeath(actionKeys, func() {
			// onComplete callback - defer reset until after frame completes
			// This prevents race conditions between Update and Draw
			if signals := ecs.Resource[resources.GameSignals](world); signals != nil && resetFunc != nil {
				signals.SetPendingReset(resetFunc)
			}
		}, cfg)
	}

	// Update active death transition
	if tm.GetActiveType() == resources.TransitionDeath {
		tm.UpdateDeath(dt)
	}
}
