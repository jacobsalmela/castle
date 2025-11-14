package combat

// AttackActive represents an ongoing attack action from an entity.
type AttackActive struct {
	AttackTag     string
	Damage        float64
	StaminaDamage float64
	ReactForce    float64
	PushForce     float64
	Mult          float64
	FilterOut     []*Hitbox

	// Transient state below (initialized by attack system)
	Contacted   map[*Hitbox]struct{}
	OnceApplied bool
	// LastSlicePresent tracks whether the hitbox slice was present on the previous frame
	LastSlicePresent bool
	// ShakeNum is the number of screen shakes to apply on hit
	ShakeNum int
}
