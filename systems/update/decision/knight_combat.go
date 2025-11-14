package decision

import (
	"game/components"
	"game/systems/update/entities/animation"
)

// KnightStopShield deactivates the knight's shield and removes the shield hitbox.
// Used by the knight behavior tree and death handling.
func KnightStopShield(knight *components.Knight, anim *components.Animation, hitbox *components.Hitbox) {
	if knight == nil || !knight.ShieldActive {
		return
	}
	if hitbox != nil && len(hitbox.Boxes) > 0 {
		// Pure ECS: Remove shield box (last box added)
		hitbox.Boxes = hitbox.Boxes[:len(hitbox.Boxes)-1]
	}
	knight.ShieldActive = false
	if anim != nil && anim.State == components.ParryBlockTag {
		animation.SetAnimationState(anim, components.IdleTag)
	}
}
