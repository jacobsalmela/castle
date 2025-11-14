package combat

import "game/pkg/bump"

// ContactType defines how a hitbox interacts with other entities.
type ContactType int

const (
	// Hit represents a hurtbox that receives damage (defensive)
	Hit ContactType = iota
	// Block prevents damage while still registering contact (defensive enhancement)
	Block
	// ParryBlock represents a parry window that can reflect damage (defensive counter)
	ParryBlock
)

// HitboxRect represents a single collision box relative to an entity's position.
// All coordinates are offsets from the entity's Transform position.
// Hitbox queries CollisionSpace to find targets to damage
type HitboxRect struct {
	X       float64     // X offset from entity position
	Y       float64     // Y offset from entity position
	W       float64     // Width of hitbox
	H       float64     // Height of hitbox
	Contact ContactType // Interaction behavior (static)

	// ContactFunc provides dynamic contact type resolution.
	// When set, this function is called instead of using the static Contact field.
	// Used for mechanics like knight shield parry→block transitions.
	// Optional - most hitboxes use static Contact field.
	ContactFunc func() ContactType
}

// ToBumpRect converts HitboxRect to bump.Rect with world position offset.
func (r HitboxRect) ToBumpRect(entityX, entityY float64) bump.Rect {
	return bump.Rect{
		X: entityX + r.X,
		Y: entityY + r.Y,
		W: r.W,
		H: r.H,
	}
}

// ResolveContact returns the active ContactType for this hitbox.
// If ContactFunc is set, it calls the function for dynamic resolution.
// Otherwise, returns the static Contact field.
func (r HitboxRect) ResolveContact() ContactType {
	if r.ContactFunc != nil {
		return r.ContactFunc()
	}
	return r.Contact
}

// Hitbox is a component containing collision boxes for an entity.
// Boxes are stored as offsets from the entity's Transform position, making
// them independent of entity location. Systems handle collision detection
// by combining Hitbox + Transform components.
type Hitbox struct {
	// Boxes contains all active collision boxes for this entity.
	// Typically:
	//   - Index 0: Hurtbox (defensive, ContactType=Hit)
	//   - Index 1+: Block boxes, attack boxes, etc.
	Boxes []HitboxRect

	// DebugRect stores the last attack hitbox for debug rendering.
	// Optional - only set during active attacks for visualization.
	DebugRect *HitboxRect
}
