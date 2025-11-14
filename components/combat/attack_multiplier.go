package combat

// AttackMultiplier tracks bonus damage from consumed heal potions.
// Each heal used increases the attack multiplier, rewarding aggressive play
// that uses healing strategically rather than hoarding potions.
type AttackMultiplier struct {
	// Current multiplier value (e.g., 0.4 = +40% damage)
	Current float64

	// Multiplier gained per heal consumed
	PerHeal float64
}
