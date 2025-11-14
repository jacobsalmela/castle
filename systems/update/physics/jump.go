package physics

import (
	"game/components"
)

// CanJump validates whether an entity can perform a jump action.
// This migrates the logic from Control.CanJump() to Pure ECS.
//
// Jump is allowed when ALL of the following conditions are met:
// 1. Entity has sufficient stamina (> 0)
// 2. Entity is grounded OR climbing
// 3. Entity is NOT blocking or parrying
// 4. Entity is NOT consuming an item
//
// Parameters:
//   - stamina: Stamina component for stamina check
//   - anim: Animation component for state checks
//   - physics: Physics component for grounded check
//
// Returns: true if entity can jump, false otherwise
func CanJump(stamina *components.Stamina, anim *components.Animation, physics *components.Physics) bool {
	// Nil safety checks
	if stamina == nil || anim == nil || physics == nil {
		return false
	}

	// Must have stamina
	if stamina.Current <= 0 {
		return false
	}

	// Must be grounded or climbing
	isGroundedOrClimbing := false
	if anim.State == components.ClimbTag {
		isGroundedOrClimbing = true
	} else if physics.Grounded {
		isGroundedOrClimbing = true
	}
	if !isGroundedOrClimbing {
		return false
	}

	// Cannot jump while blocking/parrying
	isBlocking := anim.State == components.BlockTag || anim.State == components.ParryBlockTag
	if isBlocking {
		return false
	}

	// Cannot jump while consuming items
	if anim.State == components.ConsumeTag {
		return false
	}

	return true
}
