package decision

import (
	"game/ecs"
)

// UpdateGram processes Gram entities.
//
// Gram is an invulnerable passive NPC with dialogue:
// - No Health component (cannot be killed)
// - Poise-only for hit reactions
// - Shows poise bar when hit
// - Dialogue interaction via Textbox
//
// Pure ECS - hitbox initialized in prefab:
// - Hurtbox added during entity construction
// - Hurt flash handled by shared combat reaction systems
//
// Components used:
// - Gram: Entity-specific behavior state
// - Animation: Sprite animation
// - Hitbox: Collision detection
// - Facing: Sprite orientation
// - Poise: Hit reaction (no Health - invulnerable)
func UpdateAcedian(world *ecs.World, _ interface{}, dt float64) {
	// Acedian is purely reactive - no per-frame behavior needed
	// All combat reactions handled by shared systems
	// Textbox interactions handled by dedicated textbox system
}
