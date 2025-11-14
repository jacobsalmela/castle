package player

// Player is a component for player-controlled character behavior.
// The player has complex input handling, combat, movement, stamina, and healing systems.
type Player struct {
	// Movement and combat tuning
	Speed           float64 // Horizontal movement speed (pixels/second)
	JumpSpeed       float64 // Vertical jump velocity
	ReactForce      float64 // Knockback reaction force
	AttackPushForce float64 // Attack push force
	AttackLevel     float64 // Heavy attack charge level (0.0-1.0)
	JumpCost        float64 // Stamina cost per jump
	AttackDamage    float64 // Base attack damage
	ClimbSpeed      float64 // Climb movement speed multiplier
}
