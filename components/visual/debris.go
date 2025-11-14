package visual

// Debris is a VFX particle behavior (destruction particles).
// Debris spawns with random velocity and rotation, then despawns after a timer.
type Debris struct {
	Timer         float64 // Lifetime remaining (seconds)
	RotationSpeed float64 // Rotation speed (radians per second)
	ImageIndex    int     // Current sprite frame index (for multi-frame debris)
}
