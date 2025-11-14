package combat

import "game/pkg/tween"

// Stamina tracks an entity's stamina points with recovery and smooth transitions.
// Stamina is typically used for special attacks, dodging, or blocking.
// It recovers over time at the specified RecoveryRate.
type Stamina struct {
	Current      float64      // Current stamina points
	Max          float64      // Maximum stamina points
	RecoveryRate float64      // Stamina recovered per second
	Lag          float64      // Previous stamina value for smooth transitions
	Tween        *tween.Tween // Animation tween for smooth stamina bar changes
}
