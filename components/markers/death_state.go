package markers

// DeathState tracks enemy death animation and removal timing.
type DeathState struct {
	// DieTimer counts down from DieDuration to 0 during death animation.
	// At 0, entity is removed from the world.
	DieTimer float64

	// DieDuration is the total length of death animation in seconds.
	// Set during entity creation based on enemy type.
	DieDuration float64
}
