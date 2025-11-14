package combat

// AttackIntent captures a one-shot attack request from an entity.
type AttackIntent struct {
	AttackTag     string
	Damage        float64
	StaminaDamage float64
	ReactForce    float64
	PushForce     float64
	Mult          float64
	// FilterOut allows callers to provide a slice of HitboxState to exclude
	// from resolution (e.g., self or linked hitboxes). Systems should honor
	// this when calling ResolveHitboxArea.
	FilterOut []*Hitbox
}
