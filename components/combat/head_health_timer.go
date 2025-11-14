package combat

// HeadHealthTimer tracks when to show an entity's health bar above their head.
// When an entity takes damage, the timer is set to a configured duration.
// The health bar is visible while Timer > 0, then fades out.
type HeadHealthTimer struct {
	Timer float64 // Remaining seconds to show the headbar
}
