package markers

// Stagger represents a knockback/stagger state for actors.
// When this component is present on an entity, the stagger system will apply
// the stagger animation, knockback force, and animation timing.
type Stagger struct {
	// Force is the knockback force to apply to velocityX.
	Force float64
	// MoveBack reverses force direction if true and entity is facing left.
	MoveBack bool
	// TimeMult is the animation speed multiplier (higher = slower animation).
	// A value of 1.0 means normal speed, 2.0 means half speed.
	TimeMult float64
	// Applied tracks whether the stagger has been processed.
	// The system sets this to true after applying the stagger.
	Applied bool
}
