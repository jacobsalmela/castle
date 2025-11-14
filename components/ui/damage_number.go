package ui

import (
	"game/entities"
	"game/pkg/tween"
)

// DamageNumber is a VFX text particle that displays damage dealt to entities.
// Damage numbers spawn centered on the damaged entity, then drift upward/outward
// with random trajectories while fading out.
type DamageNumber struct {
	Tween    *tween.Tween      // Animation interpolator (0→1 over duration)
	StartX   float64           // Starting X position for tween
	StartY   float64           // Starting Y position for tween
	TargetX  float64           // X offset to drift (not absolute position)
	TargetY  float64           // Y offset to drift (not absolute position)
	Damage   float64           // Damage amount to display
	Critical bool              // True if damage >= 50% of target's max health
	EntityID entities.EntityId // Entity reference for cleanup
}
