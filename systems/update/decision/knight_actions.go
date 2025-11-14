package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/systems/update/entities/animation"
)

const (
	knightAttackDamage       = 25.0
	knightAttackReact        = 12.0
	knightAttackPush         = 12.0
	knightShieldDuration     = 1.2
	knightShieldStaminaDrain = 20.0
	knightDashDuration       = 0.35
	knightDashCooldown       = 2.5
	knightDashSpeedFactor    = 4.0
	knightThrowFrame         = 2
	knightThrowRecover       = 0.4
)

// KnightEnterAttack sets up a melee attack for the knight.
// Used by the knight behavior tree.
func KnightEnterAttack(world *ecs.World, eid entities.EntityId, knight *components.Knight, facing *components.Facing, tag string, anim *components.Animation) {
	if knight == nil || anim == nil {
		return
	}
	animation.SetAnimationState(anim, tag)
	animation.SetStateEffect(anim, func() func() {
		knight.Paused = true
		return func() { knight.Paused = false }
	}, tag)
	knightRegisterAttackSlice(world, eid, knight, facing, anim)
	KnightSetVelocity(world, eid, 0, 0)
	KnightEnsureFacingTarget(world, eid, knight, facing)
}
