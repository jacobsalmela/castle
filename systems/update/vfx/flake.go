package vfx

import (
	"math/rand"
	"time"

	"game/assets"
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/tween"
	"game/systems/update/state"
)

const (
	flakeAnimDuration   = 0.2
	flakeHomingDuration = 0.8 // seconds for tween from current position to target
)

// UpdateFlake advances flake VFX particles (death loot, etc).
// Handles sprite animation, velocity updates, homing behavior, and cleanup.
func UpdateFlake(world *ecs.World, _ interface{}, dt float64) {
	if world == nil || dt == 0 {
		return
	}

	for _, eid := range world.EntitiesWith((*components.Flake)(nil)) {
		flake := ecs.GetComponent[components.Flake](world, eid)
		if flake == nil {
			continue
		}

		updateFlake(world, eid, flake, dt)
	}
}

func updateFlake(world *ecs.World, eid entities.EntityId, flake *components.Flake, dt float64) {
	// Get components
	transform := ecs.GetComponent[components.Transform](world, eid)
	render := ecs.GetComponent[components.Render](world, eid)
	if transform == nil || render == nil {
		return
	}

	// Retarget player if changed
	retargetPlayer(world, flake)

	// Animate sprite flipping (cycles between 2 frames)
	flake.Timer += dt
	if flake.Timer >= flakeAnimDuration {
		flake.Timer = 0
		flake.ImageIndex = (flake.ImageIndex + 1) % 2
		// Update render image based on flake size
		// Note: This should use Aseprite JSON frames in the future
		// For now, we just use the base image (manual frame slicing removed)
		render := ecs.GetComponent[components.Render](world, eid)
		if render != nil {
			render.Image = assets.GetSpriteImage("flake")
		}
	}

	// Start homing after delay (Pure ECS - no goroutines)
	if !flake.HomingStartTime.IsZero() && time.Now().After(flake.HomingStartTime) && flake.CaptureTween == nil {
		// Create the tween to pull flake toward target
		flake.CaptureTween = tween.New(0, 1, flakeHomingDuration, tween.EaseInQuad)
		// Record current position as tween start (flake may have drifted)
		flake.StartX = transform.X
		flake.StartY = transform.Y
	}

	// If no capture tween, just drift with physics
	if flake.CaptureTween == nil || flake.Target == 0 {
		return
	}

	// Homing behavior: interpolate toward target
	targetTransform := ecs.GetComponent[components.Transform](world, flake.Target)
	if targetTransform == nil {
		return
	}
	tx, ty, tw, th := targetTransform.X, targetTransform.Y, targetTransform.W, targetTransform.H
	distX := tx + tw*flake.RandTargetW - flake.StartX
	distY := ty + th*flake.RandTargetH - flake.StartY

	flake.CaptureTween.Update(dt)
	path := flake.CaptureTween.Value()
	transform.X = flake.StartX + path*distX
	transform.Y = flake.StartY + path*distY

	// Cleanup when captured
	if flake.CaptureTween.IsDone() {
		// Award experience to target
		if flake.Target != 0 {
			// Experience component
			state.AddExperience(world, flake.Target, 1)
		}
		// Destroy entity
		world.DestroyEntity(eid)
	}
}

func retargetPlayer(world *ecs.World, flake *components.Flake) {
	// Get player entity
	players := world.EntitiesWith((*components.Player)(nil))
	if len(players) == 0 {
		return
	}
	playerID := players[0]

	// Skip if already targeting this player
	if playerID == flake.Target {
		return
	}

	flake.Target = playerID
	flake.RandTargetW = rand.Float64()
	flake.RandTargetH = rand.Float64()
}
