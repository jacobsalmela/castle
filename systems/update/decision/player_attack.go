package decision

import (
	"strings"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/resources"
	"game/systems/update/entities/animation"
	"game/systems/update/combat"
)

// playerAttackState tracks attack hitbox contacts and application state for a single attack.
// Used by playerExecuteAttack to prevent multi-hitting the same target.
type playerAttackState struct {
	contacted      []*components.Hitbox
	lastShakeCount int
	applied        bool
}

// contains reports whether the hitbox was already contacted during this attack.
func (s *playerAttackState) contains(hb *components.Hitbox) bool {
	for _, e := range s.contacted {
		if e == hb {
			return true
		}
	}
	return false
}

// initPlayerHurtbox extracts hurtbox from player animation data and initializes hitbox component.
// This allows the player to be hit by enemies. Called once per frame to handle animation changes.
func initPlayerHurtbox(world *ecs.World, eid entities.EntityId) {
	hitbox := ecs.GetComponent[components.Hitbox](world, eid)
	anim := ecs.GetComponent[components.Animation](world, eid)
	facing := ecs.GetComponent[components.Facing](world, eid)
	transform := ecs.GetComponent[components.Transform](world, eid)

	if hitbox == nil || anim == nil || facing == nil || transform == nil {
		return
	}

	// Clear existing boxes (will be repopulated from current animation frame)
	hitbox.Boxes = nil

	// Get hurtbox slice directly from SliceMap WITHOUT sprite offsets
	// The slice data is sprite-relative, but hitboxes should be transform-relative (collision box)
	// NOT sprite-relative (visual rendering position), so we use the raw slice data
	if sliceData, ok := anim.SliceMap[components.HurtboxSliceName]; ok {
		if rect, ok := sliceData[anim.Data.CurrentFrame]; ok {
			hitbox.Boxes = append(hitbox.Boxes, components.HitboxRect{
				X:       rect.X,
				Y:       rect.Y,
				W:       rect.W,
				H:       rect.H,
				Contact: components.Hit,
			})
		}
	}
}

// playerExecuteAttack executes a player attack: coordinates animation, stamina, hit detection, and movement response.
// Pure ECS implementation - replaces Control.MultAttack.
//
//nolint:gocognit,cyclop,funlen // Gameplay branching intentionally complex; matches legacy MultAttack behavior
func playerExecuteAttack(world *ecs.World, entityID entities.EntityId, player *components.Player, anim *components.Animation, stamina *components.Stamina, poise *components.Poise, health *components.Health, hitbox *components.Hitbox, facing *components.Facing, attackTag string, damage, staminaDamage, reactForce, pushForce float64, attackMult *float64) {
	// Check stamina
	if stamina.Current <= 0 {
		return
	}

	// TODO: Apply attack multiplier when that system is migrated to Pure ECS
	// For now, assume no multiplier (damage *= 1 + 0)

	// Handle combo chaining (if AttackC animation exists, continue combo)
	if strings.HasPrefix(anim.State, attackTag) && anim.Data.Animation(anim.State+"C") != nil {
		attackTag = anim.State + "C"
	}

	// Exit blocking before attacking
	if physics := ecs.GetComponent[components.Physics](world, entityID); physics != nil {
		playerExitBlockWithStamina(anim, hitbox, stamina, physics)
	}

	// Set attack animation
	animation.SetAnimationState(anim, attackTag)

	// Pause player during attack
	animation.SetStateEffect(anim, func() func() {
		// Pause on entry
		return func() {
			// Unpause on exit (restored by animation)
		}
	})

	// Handle combo continuation - unpause on last frame if combo available
	if anim.Data.Animation(attackTag+"C") != nil {
		lastFrame := anim.Data.CurrentAnimation.To - anim.Data.CurrentAnimation.From
		animation.RegisterFrameCallback(anim, lastFrame, func() {
			// Unpause to allow combo input (callback clears paused state)
		})
	}

	// Register hitbox slice handler for attack frames
	playerRegisterAttackHitboxes(world, entityID, anim, stamina, poise, health, facing, damage, staminaDamage, reactForce, pushForce, *attackMult)
}

// playerRegisterAttackHitboxes registers a slice handler for multi-slice attacks.
// Pure ECS implementation - replaces Control.handleAttackHitboxes.
func playerRegisterAttackHitboxes(world *ecs.World, entityID entities.EntityId, anim *components.Animation, stamina *components.Stamina, poise *components.Poise, health *components.Health, facing *components.Facing, damage, staminaDamage, reactForce, pushForce, mult float64) {
	state := &playerAttackState{}

	// Get flip state from Facing component
	flipX := facing != nil && facing.FlipX
	flipY := false

	// Register slice handler for multi-slice attacks
	animation.RegisterSliceCallback(anim, components.HitboxSliceName, flipX, flipY,
		func(x, y, w, h float64, firstFrame bool) {
			callback := playerAttackSliceCallback(world, entityID, state, stamina, poise, health, facing, damage, staminaDamage, reactForce, pushForce, mult)
			callback(x, y, w, h, firstFrame)
		})
}

// playerAttackSliceCallback creates the callback for attack hitbox collision checks.
// Pure ECS implementation - replaces Control.attackSliceMultAttackCallback.
//
//nolint:gocognit,cyclop,funlen // Gameplay branching intentionally complex; matches legacy MultAttack behavior
func playerAttackSliceCallback(world *ecs.World, entityID entities.EntityId, state *playerAttackState, stamina *components.Stamina, poise *components.Poise, health *components.Health, facing *components.Facing, damage, staminaDamage, reactForce, pushForce, mult float64) func(float64, float64, float64, float64, bool) {
	return func(x, y, w, h float64, firstFrame bool) {
		// Convert to Rect for collision system
		slice := bump.Rect{X: x, Y: y, W: w, H: h}

		// Clear contacted list on first frame (segmented)
		if firstFrame {
			state.contacted = nil
			state.lastShakeCount = 0
		}

		attackMult := mult + 1
		totalDamage := damage * attackMult

		// Call combat system for hitbox resolution
		var contactType components.ContactType = components.Hit
		var hits []*components.Hitbox

		if world != nil && entityID != 0 {
			// Perform hitbox overlap detection and enqueue combat events
			contactType, hits = combat.ResolveHitboxArea(world, entityID, slice, totalDamage, state.contacted)

			// Track contacted hitboxes to prevent multi-hit
			for _, hb := range hits {
				if hb != nil && !state.contains(hb) {
					state.contacted = append(state.contacted, hb)
				}
			}

			// Player camera shake on hit
			if len(state.contacted) != state.lastShakeCount {
				state.lastShakeCount = len(state.contacted)

				if camera := ecs.Resource[resources.Camera](world); camera != nil {
					camera.Shake(0.1*float32(attackMult), 0.5*attackMult)
				}
			}

			// Handle parry-block contact (attacker gets staggered)
			if contactType == components.ParryBlock {
				poise.Current -= totalDamage
				if poise.Current < 0 {
					poise.Current = 0
				}
				if poise.Current <= 0 {
					// Get anim component for stagger
					if anim := ecs.GetComponent[components.Animation](world, entityID); anim != nil {
						playerStagger(world, entityID, anim, poise, health, facing, reactForce*(totalDamage/health.Max), true, 1)
					}
				}
			}
		}

		// Apply stamina/force once per attack (not per slice)
		if !state.applied {
			state.applied = true
			stamina.Current -= staminaDamage * attackMult
			if stamina.Current < 0 {
				stamina.Current = 0
			}

			force := pushForce
			if contactType >= components.Block {
				force = reactForce
			}

			// Apply knockback based on contact type and facing direction
			flipX := facing != nil && facing.FlipX
			if (contactType >= components.Block && flipX) || (contactType < components.Block && !flipX) {
				force *= -1
			}
			if physics := ecs.GetComponent[components.Physics](world, entityID); physics != nil {
				physics.Velocity.X += force
			}
		}
	}
}

// playerStagger applies a stagger animation and knockback (Pure ECS - replaces Control.Stagger for player).
func playerStagger(world *ecs.World, eid entities.EntityId, anim *components.Animation, poise *components.Poise, health *components.Health, facing *components.Facing, force float64, moveBack bool, timeMult float64) {
	// Exit blocking first
	if hitbox := ecs.GetComponent[components.Hitbox](world, eid); hitbox != nil {
		if stamina := ecs.GetComponent[components.Stamina](world, eid); stamina != nil {
			if physics := ecs.GetComponent[components.Physics](world, eid); physics != nil {
				playerExitBlockWithStamina(anim, hitbox, stamina, physics)
			}
		}
	}

	// Set stagger animation
	animation.SetAnimationState(anim, components.StaggerTag)

	// Adjust animation speed if needed
	if timeMult != 1 {
		animation.SetStateEffect(anim, func() func() {
			prevPlaySpeed := anim.Data.PlaySpeed
			anim.Data.PlaySpeed = float32(1.0 / timeMult)
			return func() { anim.Data.PlaySpeed = prevPlaySpeed }
		})
	}

	// Apply knockback force
	if moveBack && facing != nil && facing.FlipX {
		force *= -1
	}
	if physics := ecs.GetComponent[components.Physics](world, eid); physics != nil {
		physics.Velocity.X += force
	}
}

// handleAttack triggers player attack when action key is pressed (Pure ECS - no Control).
func handleAttack(world *ecs.World, eid entities.EntityId, player *components.Player, anim *components.Animation, hitbox *components.Hitbox, facing *components.Facing, input *components.Input) {
	// Try to consume buffered action press (Pure ECS: check and clear buffer)
	if !input.Buffer[components.InputKeyAction] {
		return
	}
	input.Buffer[components.InputKeyAction] = false

	// Query Pure ECS components
	stamina := ecs.GetComponent[components.Stamina](world, eid)
	poise := ecs.GetComponent[components.Poise](world, eid)
	health := ecs.GetComponent[components.Health](world, eid)
	if stamina == nil || poise == nil || health == nil {
		return
	}

	damage := player.AttackDamage
	if damage == 0 {
		damage = player.ReactForce
	}
	// Pure ECS attack - no Control component needed
	playerExecuteAttack(world, eid, player, anim, stamina, poise, health, hitbox, facing, components.AttackTag, damage, damage, player.ReactForce, player.AttackPushForce, &player.AttackLevel)
}

// heavyAttackTick handles heavy attack charging logic.
// Moved from Player.HeavyAttackTick() method during Pure ECS migration.
func heavyAttackTick(player *components.Player, anim *components.Animation, input *components.Input) {
	if player == nil || anim == nil || anim.Data == nil {
		return
	}
	if !strings.HasPrefix(anim.State, components.AttackTag) {
		return
	}
	state := anim.State
	holdAnim := anim.Data.Animation(state + "Hold")
	if holdAnim == nil {
		return
	}
	startingFrames := holdAnim.From - anim.Data.CurrentAnimation.From
	frame := anim.Data.CurrentFrame
	if frame >= holdAnim.From && frame <= holdAnim.To {
		den := float64(holdAnim.To - holdAnim.From + startingFrames)
		if den <= 0 {
			return
		}
		player.AttackLevel = float64(frame-holdAnim.From+startingFrames) / den
		if input.KeyReleased[components.InputKeyAction] {
			anim.Data.CurrentFrame = holdAnim.To + 2
			anim.Data.FrameCounter = 0
		}
		return
	}
	if frame == holdAnim.From-startingFrames && input.KeyReleased[components.InputKeyAction] {
		anim.Data.CurrentFrame = holdAnim.To + 1
	}
}
