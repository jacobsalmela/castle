package decision

import (
	"math"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/pkg/config"
	"game/resources"
	"game/systems/update/entities/animation"
)

// ============================================================================
// LADDER CLIMBING SYSTEM
// ============================================================================
//
// Ladder mechanics:
// 1. Ladders are tiles marked with type="ladder" or class="ladder" in Tiled
// 2. Player can walk through ladders horizontally (no collision)
// 3. Player can stand on top of ladders (when positioned above)
// 4. Player can climb ladders with UP/DOWN keys (when overlapping)
// 5. Player can descend through ladder top with DOWN key
//
// State transitions:
// - Press UP while overlapping ladder → Enter climb mode
// - Press DOWN while on top of ladder → Drop through and climb
// - Press DOWN while overlapping ladder → Climb down
// - Jump or move off ladder → Exit climb mode
// ============================================================================

const (
	ladderTopTolerance = 2.0 // pixels of tolerance for "on top of" ladder detection
)

// applyClimb processes climb input and physics while in climb state.
func applyClimb(world *ecs.World, eid entities.EntityId, player *components.Player, intents *components.ActionIntents, anim *components.Animation, dt float64, input *components.Input, cfg *config.Config) {
	phys := ecs.GetComponent[components.Physics](world, eid)
	if intents == nil || phys == nil {
		return
	}

	// Update climb intents from input
	up := input.KeyDown[components.InputKeyUp]
	down := input.KeyDown[components.InputKeyDown]
	intents.ClimbHeld = up || down
	intents.ClimbDrop = down
	intents.ClimbRelease = false // Clear release flag each frame

	// Only apply climb movement if currently climbing
	if anim.State != components.ClimbTag {
		return
	}

	// Check if still on ladder - exit if not
	transform := ecs.GetComponent[components.Transform](world, eid)
	if transform == nil {
		return
	}

	ladderRegistry := ecs.Resource[resources.LadderRegistry](world)
	if ladderRegistry == nil {
		// No ladders registered - can't climb
		intents.ClimbRelease = true
		return
	}

	// Check if still overlapping with any ladder OR on top and descending
	playerRect := bump.Rect{X: transform.X, Y: transform.Y, W: transform.W, H: transform.H}
	overlapping := ladderRegistry.FindOverlapping(playerRect)
	stillOnLadder := len(overlapping) > 0

	// If not overlapping, check if on top and pressing down (starting descent from top)
	// Note: We don't check phys.Grounded here because climb state disables gravity (Weight=0),
	// which causes Grounded to be false even when standing on ladder top
	if !stillOnLadder && input.KeyDown[components.InputKeyDown] {
		const ladderTopTolerance = 2.0
		onTop, _ := ladderRegistry.IsOnTopOf(playerRect, ladderTopTolerance)
		if onTop {
			stillOnLadder = true
			// Enable drop-through so we can descend through the platform
			phys.DroppingThrough = true
		}
	}

	if !stillOnLadder {
		// Not on ladder anymore - fall off
		intents.ClimbRelease = true
		return
	}

	// Apply climb movement
	climbSpeed := player.Speed
	if climbSpeed == 0 {
		climbSpeed = cfg.Body.MaxX
	}
	factor := player.ClimbSpeed
	speed := climbSpeed * factor * dt

	// Zero velocity (climbing overrides physics)
	phys.Velocity.X = 0
	phys.Velocity.Y = 0

	// Apply directional movement
	if input.KeyDown[components.InputKeyLeft] {
		phys.Velocity.X = -speed
	}
	if input.KeyDown[components.InputKeyRight] {
		phys.Velocity.X = speed
	}
	if input.KeyDown[components.InputKeyUp] {
		phys.Velocity.Y = -speed
	}
	if input.KeyDown[components.InputKeyDown] {
		phys.Velocity.Y = speed
		// Enable drop-through every frame while descending to pass through ladder platforms
		phys.DroppingThrough = true
	}
}

// processClimbIntents applies climb state changes based on action intents.
// Handles entering and exiting climb state based on ladder detection.
func processClimbIntents(world *ecs.World, eid entities.EntityId, intents *components.ActionIntents, phys *components.Physics, anim *components.Animation, transform *components.Transform) {
	if intents == nil || phys == nil || anim == nil || transform == nil {
		return
	}

	// Attempt to start climbing when climb input held
	if intents.ClimbHeld {
		playerClimbOn(world, eid, intents.ClimbDrop, anim, phys, transform)
	}

	// Exit climbing when climb is released
	if intents.ClimbRelease {
		playerClimbOff(anim)
	}
}

// playerClimbOn initiates climbing state.
// Player can climb if:
// 1. Overlapping with a ladder (pressing UP or DOWN)
// 2. Standing on top of ladder and pressing DOWN (descend)
func playerClimbOn(world *ecs.World, eid entities.EntityId, pressedDown bool, anim *components.Animation, phys *components.Physics, transform *components.Transform) {
	// Already climbing - do nothing
	if anim.State == components.ClimbTag {
		return
	}

	// Get ladder registry
	ladderRegistry := ecs.Resource[resources.LadderRegistry](world)
	if ladderRegistry == nil {
		return // No ladders in world
	}

	// Create player rectangle
	playerRect := bump.Rect{X: transform.X, Y: transform.Y, W: transform.W, H: transform.H}

	// Check if overlapping with any ladder
	overlapping := ladderRegistry.FindOverlapping(playerRect)
	canClimb := len(overlapping) > 0
	onTopOfLadder := false

	// When pressing DOWN, also check if standing on top of ladder
	// This handles the case where player climbed to top and is standing on it
	if pressedDown {
		onTop, _ := ladderRegistry.IsOnTopOf(playerRect, ladderTopTolerance)
		if onTop && phys.Grounded {
			// Allow descending through ladder top
			canClimb = true
			onTopOfLadder = true
		}
	}

	if !canClimb {
		return
	}

	// Exit blocking if active
	if hitbox := ecs.GetComponent[components.Hitbox](world, eid); hitbox != nil {
		if stamina := ecs.GetComponent[components.Stamina](world, eid); stamina != nil {
			playerExitBlockWithStamina(anim, hitbox, stamina, phys)
		}
	}

	// If descending from top of ladder, enable drop-through so physics allows passing through
	if onTopOfLadder && pressedDown {
		phys.DroppingThrough = true
	}

	// Enter climb state - set animation and remove gravity
	animation.SetAnimationState(anim, components.ClimbTag)
	animation.SetStateEffect(anim, func() func() {
		prevWeight := phys.Weight
		phys.Weight = 0 // Zero gravity while climbing
		return func() {
			phys.Weight = prevWeight // Restore gravity when done
		}
	})
}

// playerClimbOff exits climbing state.
func playerClimbOff(anim *components.Animation) {
	// Only exit if currently climbing
	if anim.State != components.ClimbTag {
		return
	}
	animation.SetAnimationState(anim, components.IdleTag)
}

// playerExitBlock exits blocking state without Control component.
// This is a Pure ECS helper called by climb and other systems.
func playerExitBlock(world *ecs.World, eid entities.EntityId, anim *components.Animation) {
	// Check if currently blocking
	if anim.State != components.BlockTag && anim.State != components.ParryBlockTag {
		return
	}

	// Exit blocking animation
	animation.SetAnimationState(anim, components.IdleTag)

	// Remove block hitbox (last added box)
	if hitbox := ecs.GetComponent[components.Hitbox](world, eid); hitbox != nil && len(hitbox.Boxes) > 0 {
		hitbox.Boxes = hitbox.Boxes[:len(hitbox.Boxes)-1]
	}
}

// ============================================================================
// HORIZONTAL MOVEMENT
// ============================================================================

// applyHorizontalMovement processes horizontal movement input.
func applyHorizontalMovement(world *ecs.World, eid entities.EntityId, player *components.Player, anim *components.Animation, dt float64, input *components.Input, cfg *config.Config) {
	phys := ecs.GetComponent[components.Physics](world, eid)
	if phys == nil {
		return
	}

	facing := ecs.GetComponent[components.Facing](world, eid)
	if facing == nil {
		return
	}

	speed := player.Speed
	if speed == 0 {
		speed = cfg.Body.MaxX
	}
	if !phys.Grounded {
		speed /= 2
	}

	maxX := phys.MaxVelocity.X
	if maxX == 0 {
		maxX = cfg.Body.MaxX
	}

	// Process left/right input and update velocity
	leftPressed := input.KeyDown[components.InputKeyLeft]
	rightPressed := input.KeyDown[components.InputKeyRight]
	isMovingInput := leftPressed || rightPressed

	if leftPressed && math.Abs(phys.Velocity.X) <= maxX {
		phys.Velocity.X -= speed * dt
	}
	if rightPressed && math.Abs(phys.Velocity.X) <= maxX {
		phys.Velocity.X += speed * dt
	}

	// Update facing direction if not blocking
	if !playerIsBlocking(anim) {
		if leftPressed {
			facing.FlipX = false
		} else if rightPressed {
			facing.FlipX = true
		}
	}

	// Update walk animation based on movement input
	updateWalkAnimation(anim, isMovingInput)
}

// updateWalkAnimation transitions between idle and walk based on movement input.
func updateWalkAnimation(anim *components.Animation, isMoving bool) {
	// Only transition if in idle or walk state (don't interrupt attacks, climbs, etc.)
	if anim.State != components.IdleTag && anim.State != components.WalkTag {
		return
	}

	if isMoving && anim.State == components.IdleTag {
		animation.SetAnimationState(anim, components.WalkTag)
	} else if !isMoving && anim.State == components.WalkTag {
		animation.SetAnimationState(anim, components.IdleTag)
	}
}

// ============================================================================
// JUMPING
// ============================================================================

// applyJump handles jump input and execution.
func applyJump(world *ecs.World, eid entities.EntityId, player *components.Player, intents *components.ActionIntents, anim *components.Animation, stamina *components.Stamina, input *components.Input) {
	if !input.KeyPressed[components.InputKeyJump] {
		return
	}

	phys := ecs.GetComponent[components.Physics](world, eid)
	if phys == nil {
		return
	}

	// Pure ECS jump validation - inline instead of using CanJump
	// Must have stamina
	if stamina.Current <= 0 {
		return
	}

	// Must be grounded or climbing
	isGroundedOrClimbing := false
	if anim.State == components.ClimbTag {
		isGroundedOrClimbing = true
	} else if phys.Grounded {
		isGroundedOrClimbing = true
	}
	if !isGroundedOrClimbing {
		return
	}

	// Cannot jump while blocking/parrying
	if anim.State == components.BlockTag || anim.State == components.ParryBlockTag {
		return
	}

	// Cannot jump while consuming items
	if anim.State == components.ConsumeTag {
		return
	}

	// Valid jump - apply it
	if intents != nil {
		intents.ClimbRelease = true
	}
	jumpCost := player.JumpCost
	if jumpCost != 0 {
		stamina.Current -= jumpCost
		if stamina.Current < 0 {
			stamina.Current = 0
		}
	}
	jumpSpeed := player.JumpSpeed
	if jumpSpeed == 0 {
		jumpSpeed = player.Speed
	}
	phys.Velocity.Y = -jumpSpeed
}
