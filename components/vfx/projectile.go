package vfx

import "game/entities"

// Projectile is a projectile component.
type Projectile struct {
	Owner       entities.EntityId // Entity that spawned this projectile
	Damage      float64           // Damage dealt on contact (0 when bouncing)
	SpawnGrace  int               // Frames to skip collision checks after spawn
	Bouncing    bool              // True when in bounce/roll state
	BounceTimer float64           // Time remaining before cleanup (seconds)
}
