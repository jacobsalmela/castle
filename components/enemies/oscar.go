package enemies

// Oscar is a passive NPC with dialogue and health.
type Oscar struct {
	DeadText     string // Dialogue text to display when Oscar dies
	HitboxInited bool   // Tracks if hitbox has been initialized from animation
	DeathHandled bool   // Prevents death animation from retriggering every frame
}
