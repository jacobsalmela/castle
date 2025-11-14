package decision

import (
	"time"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/systems/update/entities/animation"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const playerBufferWindow = 500 * time.Millisecond

// ══════════════════════════════════════════════════════════════════════════════
// PLAYER GUARD FUNCTIONS
// Consolidated from player_guard.go during Phase 8 refactor (10/30/2025)
// ══════════════════════════════════════════════════════════════════════════════

// processPlayerGuard handles blocking state transitions for the player entity.
// Pure ECS implementation - replaces Control.ShieldUp/ShieldDown/BlockingState.
//
// Blocking mechanics:
//   - ShieldUp adds a ParryBlock hitbox and reduces movement speed
//   - ShieldDown removes the block hitbox and restores normal movement
//   - BlockingState prevents facing direction changes while guarding
//
// Note: Sprite overlap with enemies during blocking is a visual artifact of the block hitbox
// extending beyond the player sprite. This is intentional for gameplay feel. Modifying
// Transform.W causes collision system desync and breaks ground collision.
//
// Implementation note: playerEnterBlock is called every frame while ShieldHeld is true.
// It checks if already blocking with hitbox present before applying changes to avoid
// flickering caused by repeatedly calling SetStateEffect (which restores then re-applies
// the stat modifications each call). The hitbox is verified by checking BoxCount and
// ContactType to ensure it's present throughout the blocking duration.
func processPlayerGuard(anim *components.Animation, hitbox *components.Hitbox, stamina *components.Stamina, physics *components.Physics, intents *components.ActionIntents) {
	currentlyBlocking := playerIsBlocking(anim)

	// Remove shield hitbox during consume animation (but don't change animation state)
	if anim.State == components.ConsumeTag {
		if currentlyBlocking && len(hitbox.Boxes) > 0 {
			// Just remove hitbox, let heal animation handle state
			hitbox.Boxes = hitbox.Boxes[:len(hitbox.Boxes)-1]
		}
		return
	}

	// Process shield held state
	if intents.ShieldHeld {
		// Enter/maintain blocking state
		// Always call playerEnterBlock to ensure animation and effects are properly set
		// even if animation FSM caused a transition away from blocking states
		playerEnterBlock(anim, hitbox, stamina, physics)
	} else {
		// Exit blocking state if currently blocking
		if currentlyBlocking {
			playerExitBlockWithStamina(anim, hitbox, stamina, physics)
		}
	}

	// Process explicit shield release (takes priority)
	if intents.ShieldRelease {
		if currentlyBlocking {
			playerExitBlockWithStamina(anim, hitbox, stamina, physics)
		}
		intents.ShieldRelease = false // Clear flag
	}
}

// playerIsBlocking reports whether the player is currently blocking (Pure ECS - replaces Control.BlockingState).
func playerIsBlocking(anim *components.Animation) bool {
	return anim.State == components.BlockTag || anim.State == components.ParryBlockTag
}

// playerEnterBlock enters blocking/parry state (Pure ECS - replaces Control.ShieldUp).
// This function is idempotent and can be called every frame while blocking to maintain the state.
//
// KNOWN ISSUES:
//   - Quick shield raise after lowering may fail to re-add hitbox in some timing scenarios
//     (Timing between hitbox removal and re-detection can cause shield to not go up immediately)
//   - Animation flickering may occur during rapid state transitions
//
// BLOCKING MECHANICS (WORKING AS DESIGNED):
// - ParryBlock reduces damage to 10% (chip damage) - see player_block.go ApplyPlayerBlock()
// - This is a 90% damage reduction, not a full block
// - Example: 20 damage attack → 2 chip damage when blocking
// - Also drains stamina and prevents knockback/stagger (unless stamina depleted)
//
// TODO: Add telemetry/logging to track hitbox add/remove cycles for debugging re-raise issue
func playerEnterBlock(anim *components.Animation, hitbox *components.Hitbox, stamina *components.Stamina, physics *components.Physics) {
	wasBlocking := playerIsBlocking(anim)

	// Check if block hitbox already exists
	// Search through all hitboxes to find a ParryBlock (not just the last one)
	hasBlockHitbox := false
	for i := 0; i < len(hitbox.Boxes); i++ {
		if hitbox.Boxes[i].Contact == components.ParryBlock {
			hasBlockHitbox = true
			break
		}
	}

	// If already blocking with hitbox present, nothing to do
	if wasBlocking && hasBlockHitbox {
		return
	}

	// Entering blocking for the first time OR re-entering after FSM transition or quick release
	if !wasBlocking {
		animation.SetAnimationState(anim, components.ParryBlockTag)
	}

	// Apply state effect (only when first entering or re-entering)
	// This reduces movement speed and stamina recovery
	animation.SetStateEffect(anim, func() func() {
		prevMaxX := physics.MaxVelocity.X
		prevStaminaRecoverRate := stamina.RecoveryRate

		physics.MaxVelocity.X = prevMaxX / 2
		stamina.RecoveryRate /= 3

		return func() {
			physics.MaxVelocity.X = prevMaxX
			stamina.RecoveryRate = prevStaminaRecoverRate
		}
	}, components.ParryBlockTag, components.BlockTag)

	// Always ensure block hitbox is present when blocking
	// This handles quick re-raise scenarios where hitbox might be missing
	if !hasBlockHitbox {
		// Get block hitbox from animation frame slice map
		keys := anim.SliceMap[components.BlockSliceName]
		if keys == nil {
			// BUG: If we can't get the block slice, blocking won't work
			// This can happen if animation doesn't have BlockSliceName defined
			return
		}

		blockSlice, ok := keys[anim.Data.CurrentFrame]
		if !ok {
			// BUG: Slice not present on current frame
			return
		}

		// Add parry block hitbox
		hitbox.Boxes = append(hitbox.Boxes, components.HitboxRect{
			X:       blockSlice.X,
			Y:       blockSlice.Y,
			W:       blockSlice.W,
			H:       blockSlice.H,
			Contact: components.ParryBlock,
		})
	}
}

// playerExitBlockWithStamina exits blocking state (Pure ECS - replaces Control.ShieldDown).
func playerExitBlockWithStamina(anim *components.Animation, hitbox *components.Hitbox, stamina *components.Stamina, physics *components.Physics) {
	if !playerIsBlocking(anim) {
		return
	}

	// Exit blocking animation
	animation.SetAnimationState(anim, components.IdleTag)

	// Remove block hitbox (last added box)
	if len(hitbox.Boxes) > 0 {
		hitbox.Boxes = hitbox.Boxes[:len(hitbox.Boxes)-1]
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// PLAYER UPDATE SYSTEM
// ══════════════════════════════════════════════════════════════════════════════

func UpdatePlayer(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		return
	}

	for _, eid := range world.EntitiesWith((*components.Player)(nil), (*components.ActionIntents)(nil)) {
		player := ecs.GetComponent[components.Player](world, eid)
		intents := ecs.GetComponent[components.ActionIntents](world, eid)
		facing := ecs.GetComponent[components.Facing](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		stamina := ecs.GetComponent[components.Stamina](world, eid)
		healing := ecs.GetComponent[components.Healing](world, eid)
		input := ecs.GetComponent[components.Input](world, eid)
		physics := ecs.GetComponent[components.Physics](world, eid)
		if player == nil || intents == nil || facing == nil || anim == nil || health == nil || stamina == nil || healing == nil || input == nil || physics == nil {
			continue
		}
		if !playerReady(anim, health, stamina, intents, physics) {
			continue
		}

		// Initialize player hurtbox from animation data if needed
		initPlayerHurtbox(world, eid)

		// Process player guard/blocking state (consolidated from player_guard.go)
		hitbox := ecs.GetComponent[components.Hitbox](world, eid)
		if hitbox != nil {
			processPlayerGuard(anim, hitbox, stamina, physics, intents)
		}

		// Get transform for collision operations
		transform := ecs.GetComponent[components.Transform](world, eid)

		if health.Current > 0 {
			applyPlayerInput(world, eid, player, intents, facing, anim, health, stamina, healing, transform, dt, cfg)
			heavyAttackTick(player, anim, input) // Moved from Player method
		}
	}
}

func playerReady(anim *components.Animation, health *components.Health, stamina *components.Stamina, intents *components.ActionIntents, physics *components.Physics) bool {
	return anim != nil && health != nil && stamina != nil && intents != nil && physics != nil
}

func applyPlayerInput(world *ecs.World, eid entities.EntityId, player *components.Player, intents *components.ActionIntents, facing *components.Facing, anim *components.Animation, health *components.Health, stamina *components.Stamina, healing *components.Healing, transform *components.Transform, dt float64, cfg *config.Config) {
	if intents == nil {
		return
	}
	if !canProcessInput(anim) {
		// Clear intents during consume animation
		intents.ShieldHeld = false
		intents.ClimbHeld = false
		intents.ClimbDrop = false
		intents.Heal = false
		return
	}

	// Get hitbox component for attack system
	hitbox := ecs.GetComponent[components.Hitbox](world, eid)

	// Get input component
	input := ecs.GetComponent[components.Input](world, eid)
	if input == nil {
		return // Cannot process input without Input component
	}

	// Get physics for climb detection
	physics := ecs.GetComponent[components.Physics](world, eid)

	bufferedActions(world, eid, player, intents, facing, anim, hitbox, input)
	applyGuard(intents, input)
	applyClimb(world, eid, player, intents, anim, dt, input, cfg)
	processClimbIntents(world, eid, intents, physics, anim, transform) // Pure ECS climb
	applyHorizontalMovement(world, eid, player, anim, dt, input, cfg)  // Pure ECS
	applyJump(world, eid, player, intents, anim, stamina, input)
	applyDebugShortcuts(healing, cfg)
}

func canProcessInput(anim *components.Animation) bool {
	return anim.State != components.ConsumeTag
}

func bufferedActions(world *ecs.World, eid entities.EntityId, player *components.Player, intents *components.ActionIntents, facing *components.Facing, anim *components.Animation, hitbox *components.Hitbox, input *components.Input) {
	handleAttack(world, eid, player, anim, hitbox, facing, input)
	if intents != nil {
		// Try to consume buffered heal press (Pure ECS: check and clear buffer)
		if input.Buffer[components.InputKeyHeal] {
			input.Buffer[components.InputKeyHeal] = false
			intents.Heal = true
		}
	}
	handleDash(world, eid, player, facing, input)
}

func handleDash(world *ecs.World, eid entities.EntityId, player *components.Player, facing *components.Facing, input *components.Input) {
	// Try to consume buffered dash press (Pure ECS: check and clear buffer)
	if !input.Buffer[components.InputKeyDash] {
		return
	}
	input.Buffer[components.InputKeyDash] = false

	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}

	maxX := physics.MaxVelocity.X
	if maxX == 0 {
		maxX = player.Speed
	}
	speed := maxX * 4
	if speed == 0 {
		speed = player.Speed
	}
	if (!facing.FlipX && !input.KeyDown[components.InputKeyRight]) || input.KeyDown[components.InputKeyLeft] {
		speed *= -1
	}
	physics.Velocity.X = speed
}

func applyGuard(intents *components.ActionIntents, input *components.Input) {
	if intents == nil {
		return
	}
	intents.ShieldHeld = input.KeyDown[components.InputKeyGuard]
	if input.KeyReleased[components.InputKeyGuard] {
		intents.ShieldRelease = true
	}
}

func applyDebugShortcuts(healing *components.Healing, cfg *config.Config) {
	if !cfg.Debug {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		healing.Count = healing.MaxCount
	}
}
