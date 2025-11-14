package prefabs

import "game/components"

// NewHitbox constructs a Hitbox component for an entity.
//
// This creates the defensive hurtbox for an entity by extracting the
// "hurtbox" slice from the entity's animation data. The hurtbox represents
// the vulnerable area where the entity can take damage.
//
// The hitbox is initialized with:
//   - Position/size from the animation's "hurtbox" slice
//   - ContactType set to Hit (defensive collision)
//
// Parameters:
//   - anim: The entity's animation component containing frame slice data
//
// Returns:
//   - *Hitbox: Initialized hitbox component with hurtbox collision box
//
// Usage in prefabs:
//
//	hitbox := prefabs.NewHitbox(anim)
//	world.AddComponent(entityID, hitbox)
//
// Note: This only handles the initial hurtbox. Attack hitboxes and block
// boxes are added/removed dynamically by systems during gameplay.
func NewHitbox(x, y, w, h float64) *components.Hitbox {
	hitbox := &components.Hitbox{}

	// Create defensive hurtbox as the base collision box
	hitbox.Boxes = append(hitbox.Boxes, components.HitboxRect{
		X:       x,
		Y:       y,
		W:       w,
		H:       h,
		Contact: components.Hit,
	})
	return hitbox
}
