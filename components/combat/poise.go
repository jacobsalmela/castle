package combat

import "game/pkg/tween"

// Poise tracks an entity's poise (stagger resistance) with timed recovery.
// When poise is broken (depleted), the entity is staggered.
// After taking poise damage, recovery is delayed by RecoveryTimer seconds.
type Poise struct {
	Current        float64      // Current poise points
	Max            float64      // Maximum poise points
	RecoveryTimer  float64      // Remaining time (seconds) before poise recovers
	RecoverSeconds float64      // Delay (seconds) before poise starts recovering
	Lag            float64      // Previous poise value for smooth transitions
	Tween          *tween.Tween // Animation tween for smooth poise bar changes
}
