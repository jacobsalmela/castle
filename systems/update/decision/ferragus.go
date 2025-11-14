package decision

import (
	"game/ecs"
)

// UpdateFerragus handles Ferragus boss NPC behavior in Pure ECS style.
//
// Ferragus is a passive boss NPC with dialogue. Unlike Oscar, Ferragus is
// completely invulnerable (no Health component) but can show hit reactions
// via Poise stagger.
//
// Pure ECS - hitbox initialized in prefab:
// - Hurtbox added during entity construction
// - Hurt flash handled by shared combat reaction systems
//
// Required Components:
//   - Ferragus: Entity-specific state
//   - Poise: Stagger resistance (Ferragus has no Health - invulnerable)
//   - Animation: Animation state
//   - Textbox: Dialogue display and interaction
//   - Collider: Physics state (unmovable/solid)
//   - Hitbox: Collision detection
//   - Facing: Sprite orientation
func UpdateFerragus(world *ecs.World, _ interface{}, dt float64) {
	// Ferragus is purely reactive - no per-frame behavior needed
	// All combat reactions handled by shared systems
	// Textbox interactions handled by dedicated textbox system
}
