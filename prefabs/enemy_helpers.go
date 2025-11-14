package prefabs

import (
	"game/components"
	"game/systems/update/entities/animation"
)

// addHurtboxToHitbox extracts the hurtbox slice from an animation and adds it
// to the given Hitbox component. This is called during enemy prefab construction
// to set up the defensive hitbox that allows the enemy to receive damage.
//
// Parameters:
//   - anim: Animation component with hurtbox slice data
//   - hitbox: Hitbox component to add the hurtbox to
//
// Returns: true if hurtbox was successfully added, false otherwise
func addHurtboxToHitbox(anim *components.Animation, hitbox *components.Hitbox) bool {
	if anim == nil || hitbox == nil {
		return false
	}

	// Extract hurtbox slice from animation
	// Use flipX=false because hurtbox collision area should be consistent
	// regardless of sprite facing direction
	hurtboxSlice, err := animation.ExtractSlice(anim, components.HurtboxSliceName, false, false)
	if err != nil {
		return false
	}

	// Add hurtbox to hitbox component
	hitbox.Boxes = append(hitbox.Boxes, components.HitboxRect{
		X:       hurtboxSlice.X,
		Y:       hurtboxSlice.Y,
		W:       hurtboxSlice.W,
		H:       hurtboxSlice.H,
		Contact: components.Hit,
	})

	return true
}
