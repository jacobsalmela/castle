package enemies

// Knight is the Pure ECS representation of the knight enemy.
type Knight struct {
	// Behavioral state (pure data)
	Paused       bool    // Movement paused during attacks
	SecondPhase  bool    // Aggressive phase at 50% health
	ShieldActive bool    // Shield intent active
	DashCooldown float64 // Cooldown between dash attacks
}
