package combat

import "game/pkg/tween"

// AttackModifier provides a temporary damage multiplier for an entity.
// The multiplier can be increased by overhealing (healing beyond max health).
// It decays over time via a tween animation back to 0.
type AttackModifier struct {
	Multiplier  float64      // Current attack damage multiplier (additive)
	MultPerHeal float64      // Multiplier gained per heal amount of overheal
	Tween       *tween.Tween // Animation tween for multiplier decay
}
