package markers

// Facing tracks which direction an entity is facing for sprite orientation.
// Used by AI entities to face their target, and controls sprite FlipX.
type Facing struct {
	FlipX bool // true = facing right, false = facing left
}
