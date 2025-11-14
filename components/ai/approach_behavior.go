package ai

// ApproachBehavior defines how an entity moves closer to its target.
// Example: Rat approaches slowly, Ghoul approaches aggressively
type ApproachBehavior struct {
	Speed           float64 // Movement acceleration toward target (pixels/sec²)
	MaxSpeed        float64 // Maximum velocity during approach (pixels/sec)
	MinRange        float64 // Stop when within this distance of target (pixels)
	RangeAdjustment float64 // Additional range padding (0 = exact MinRange)
}
