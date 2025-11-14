package ui

// HUDData contains all data needed to render the player HUD.
type HUDData struct {
	// Health state
	Health    float64
	MaxHealth float64
	HealthLag float64 // For lag animation effect

	// Stamina state
	Stamina    float64
	MaxStamina float64
	StaminaLag float64 // For lag animation effect

	// Poise state (not currently displayed but kept for future use)
	Poise    float64
	MaxPoise float64
	PoiseLag float64

	// Heal count
	Heal    int
	MaxHeal int

	// Experience
	Exp int

	// Attack multiplier (displayed when > threshold)
	AttackMult     float64
	ShowAttackMult bool // True when AttackMult should be displayed
}
