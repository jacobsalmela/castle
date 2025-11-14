package ui

// HeadbarData contains all data needed to render enemy health bars.
// These appear above enemies when they take damage and fade after a timer.
type HeadbarData struct {
	// Health state
	Health    float64
	MaxHealth float64
	HealthLag float64 // For lag animation effect (yellow bar)

	// Visibility timer
	ShowTimer float64 // Seconds remaining to show the bar

	// Positioning
	EntityWidth float64 // Width of the entity for centering the bar
}
