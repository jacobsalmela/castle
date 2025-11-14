package combat

import "game/pkg/tween"

// Health tracks an entity's health points with smooth transitions.
// The Lag and Tween fields enable smooth visual transitions when health changes,
// typically rendered as animated bars in the HUD or enemy headbars.
type Health struct {
	Current float64      // Current health points
	Max     float64      // Maximum health points
	Lag     float64      // Previous health value for smooth transitions
	Tween   *tween.Tween // Animation tween for smooth health bar changes
}
