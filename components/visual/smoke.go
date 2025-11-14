package visual

import (
	"game/pkg/tween"
)

// Smoke is a VFX particle with drift and rotation behavior.
// Smoke particles spawn with an initial velocity (TargetX/Y offset),
// drift from their starting position, rotate, and fade out.
type Smoke struct {
	Tween        *tween.Tween // Animation interpolator (0→1 over duration)
	StartX       float64      // Starting X position for tween
	StartY       float64      // Starting Y position for tween
	TargetX      float64      // X offset to drift (not absolute position)
	TargetY      float64      // Y offset to drift (not absolute position)
	RotationRate float64      // Rotation speed (radians per second)
}
