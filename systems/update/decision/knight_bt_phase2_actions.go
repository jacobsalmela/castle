package decision

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/systems/update/entities/animation"
)

// newKnightDashReady creates a condition that checks if dash is ready (cooldown expired).
//
// Returns Success if dash cooldown has expired (<=0).
// Returns Failure if dash is still on cooldown.
func newKnightDashReady() *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			knight := ecs.GetComponent[components.Knight](world, eid)
			if knight == nil || knight.DashCooldown > 0 {
				return ai.Failure
			}
			return ai.Success
		},
	}
}

// newKnightDash creates a dash attack action for the knight.
//
// The knight dashes toward the target at high speed for a short duration.
// Uses closure-captured state to track:
//   - Dash duration timer
//   - Dash direction (toward target)
//   - Previous max velocity (for restoration after dash)
//
// OnStart: Sets dash cooldown, calculates direction, applies dash velocity
// OnTick: Counts down duration timer
// OnEnd: Restores previous max velocity, stops movement
func newKnightDash(cfg *config.Config) *ai.Action {
	duration := 0.0
	direction := 0.0
	prevMaxVelocityX := 0.0
	prevMaxVelocityY := 0.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			knight := ecs.GetComponent[components.Knight](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			if knight == nil || facing == nil || physics == nil || cfg == nil {
				return
			}

			// Save previous max velocity
			prevMaxVelocityX = physics.MaxVelocity.X
			prevMaxVelocityY = physics.MaxVelocity.Y
			if prevMaxVelocityX == 0 {
				prevMaxVelocityX = cfg.Body.MaxX
			}
			if prevMaxVelocityY == 0 {
				prevMaxVelocityY = cfg.Body.MaxX
			}

			// Set dash cooldown
			knight.DashCooldown = knightDashCooldown
			duration = knightDashDuration

			// Calculate direction toward target
			direction = KnightDirectionTowardsTarget(world, eid, knight, facing)
			if direction == 0 {
				direction = 1
			}

			// Apply dash velocity
			maxX := prevMaxVelocityX
			if maxX <= 0 {
				maxX = cfg.Body.MaxX
			}
			dashSpeed := maxX * knightDashSpeedFactor
			physics.MaxVelocity.X = dashSpeed
			physics.MaxVelocity.Y = prevMaxVelocityY
			KnightSetVelocity(world, eid, dashSpeed*direction, 0)
			knight.Paused = false
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			duration -= dt
			if duration <= 0 {
				return ai.Success
			}
			return ai.Running
		},
		OnEnd: func(world *ecs.World, eid entities.EntityId) {
			// Restore previous max velocity
			physics := ecs.GetComponent[components.Physics](world, eid)
			if physics != nil {
				physics.MaxVelocity.X = prevMaxVelocityX
				physics.MaxVelocity.Y = prevMaxVelocityY
			}
			KnightSetVelocity(world, eid, 0, 0)
		},
	}
}

// newKnightShield creates a shield defense action for the knight.
//
// The knight raises a shield that can block or parry attacks.
// Uses closure-captured state to track:
//   - Shield duration timer
//   - Previous max velocity (for restoration after shield)
//
// OnStart:
//   - Extracts shield hitbox slice from animation
//   - Reduces movement speed while shielding
//   - Adds shield hitbox with block/parry contact detection
//   - Sets shield active flag
//
// OnTick: Counts down duration timer
// OnEnd: Restores previous max velocity, removes shield hitbox
func newKnightShield(cfg *config.Config) *ai.Action {
	duration := 0.0
	prevMaxVelocityX := 0.0
	prevMaxVelocityY := 0.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			knight := ecs.GetComponent[components.Knight](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			hitbox := ecs.GetComponent[components.Hitbox](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			if knight == nil || facing == nil || anim == nil || hitbox == nil || physics == nil || cfg == nil {
				return
			}

			// Extract block slice
			blockSlice, err := animation.ExtractSlice(anim, components.BlockSliceName, facing.FlipX, false)
			if err != nil {
				return
			}

			// Save previous max velocity
			prevMaxVelocityX = physics.MaxVelocity.X
			prevMaxVelocityY = physics.MaxVelocity.Y
			if prevMaxVelocityX == 0 {
				prevMaxVelocityX = cfg.Body.MaxX
			}
			if prevMaxVelocityY == 0 {
				prevMaxVelocityY = cfg.Body.MaxX
			}

			duration = knightShieldDuration

			// Face target and activate shield
			KnightEnsureFacingTarget(world, eid, knight, facing)
			animation.SetAnimationState(anim, components.ParryBlockTag)
			animation.SetStateEffect(anim, func() func() {
				knight.Paused = true
				return func() { knight.Paused = false }
			}, components.ParryBlockTag, components.BlockTag)

			// Reduce movement speed while shielding
			physics.MaxVelocity.X = prevMaxVelocityX / 2
			physics.MaxVelocity.Y = prevMaxVelocityY

			// Add shield hitbox
			hitbox.Boxes = append(hitbox.Boxes, components.HitboxRect{
				X:       blockSlice.X,
				Y:       blockSlice.Y,
				W:       blockSlice.W,
				H:       blockSlice.H,
				Contact: components.Block,
				ContactFunc: func() components.ContactType {
					if anim != nil && anim.State == components.ParryBlockTag {
						return components.ParryBlock
					}
					return components.Block
				},
			})
			knight.ShieldActive = true
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			duration -= dt
			if duration <= 0 {
				return ai.Success
			}
			return ai.Running
		},
		OnEnd: func(world *ecs.World, eid entities.EntityId) {
			knight := ecs.GetComponent[components.Knight](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			hitbox := ecs.GetComponent[components.Hitbox](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			// Restore previous max velocity
			if physics != nil {
				physics.MaxVelocity.X = prevMaxVelocityX
				physics.MaxVelocity.Y = prevMaxVelocityY
			}

			// Stop shield
			KnightStopShield(knight, anim, hitbox)
		},
	}
}
