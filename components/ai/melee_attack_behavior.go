package ai

// MeleeAttackBehavior configures basic melee attack parameters.
// Used by simple melee enemies (Crawler, Ent, basic attacks).
// Complex attacks (Rat jump, Ghoul throw, Knight specials) remain enemy-specific.
type MeleeAttackBehavior struct {
	// Attack damage dealt to hit targets
	Damage float64

	// Knockback force applied to hit targets
	PushForce float64

	// Recoil force applied to attacker (when hitting shield/block)
	ReactForce float64

	// Animation state tag to trigger ("Attack", "Slash", etc.)
	AnimationTag string

	// HitboxSliceName is the slice name in animation data for hitbox
	// Defaults to HitboxSliceName constant if empty
	HitboxSliceName string

	// CooldownTimer tracks time until next attack allowed (runtime state)
	CooldownTimer float64
}
