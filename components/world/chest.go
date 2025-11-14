package world

// Chest is an interactive loot container component.
type Chest struct {
	Opened         bool    // Whether chest has been opened (one-time interaction)
	Reward         int     // Number of flake particles to spawn (default 100)
	AnimationStage int     // Current animation stage: 0=closed, 1=semi-open, 2=fully open
	AnimationTimer float64 // Timer for animation transitions (seconds)
}
