package visual

import (
	"time"

	"game/entities"
	"game/pkg/tween"
)

// Flake is a SFX particle behavior (death particles, loot).
// Flakes spawn with random velocity, then home to a target after a delay.
type Flake struct {
	From            entities.EntityId // Source entity that spawned the flake
	Target          entities.EntityId // Target entity the flake homes to
	CaptureTween    *tween.Tween      // Tween for homing animation
	StartX          float64           // Starting X position for tween
	StartY          float64           // Starting Y position for tween
	RandTargetW     float64           // Random offset within target width (0-1)
	RandTargetH     float64           // Random offset within target height (0-1)
	Timer           float64           // Animation timer for sprite flipping
	ImageIndex      int               // Current sprite frame index (0 or 1)
	HomingStartTime time.Time         // When to start homing behavior (Pure ECS)
}
