package spatial

import (
	"game/entities"
	"game/pkg/bump"
)

// Collider defines collision behavior and filtering for an entity.
type Collider struct {
	// === SHAPE (relative to Transform) ===
	OffsetX float64 // X offset from Transform.X
	OffsetY float64 // Y offset from Transform.Y
	Width   float64 // Override Transform.W (0 = use Transform.W)
	Height  float64 // Override Transform.H (0 = use Transform.H)

	// === COLLISION TAGS ===
	Tags      []string // What this entity IS (for others to filter)
	QueryTags []string // What to check collisions against

	// === FILTERING ===
	FilterOut []entities.EntityId // Ignore these specific entities

	// === PROPERTIES ===
	Solid     bool // Blocks movement (false = ghost/trigger)
	Immovable bool // Cannot be pushed by collisions
}

// GetBounds returns the collision rectangle in world space.
// If Width/Height are zero, uses Transform bounds.
func (c *Collider) GetBounds(transform *Transform) (x, y, w, h float64) {
	if c == nil || transform == nil {
		return 0, 0, 0, 0
	}

	x = transform.X + c.OffsetX
	y = transform.Y + c.OffsetY

	// Default to Transform size if not overridden
	w = c.Width
	if w == 0 {
		w = transform.W
	}

	h = c.Height
	if h == 0 {
		h = transform.H
	}

	return x, y, w, h
}

// ToBumpRect converts collider to collision library format.
func (c *Collider) ToBumpRect(transform *Transform) bump.Rect {
	x, y, w, h := c.GetBounds(transform)
	return bump.Rect{X: x, Y: y, W: w, H: h}
}

// ToBumpTags converts string tags to bump.Tag format.
func (c *Collider) ToBumpTags() []bump.Tag {
	if c == nil || len(c.Tags) == 0 {
		return []bump.Tag{"body"} // Default fallback
	}

	tags := make([]bump.Tag, len(c.Tags))
	for i, tag := range c.Tags {
		tags[i] = bump.Tag(tag)
	}
	return tags
}

// ToQueryTags converts query tags to bump.Tag format.
func (c *Collider) ToQueryTags() []bump.Tag {
	if c == nil || len(c.QueryTags) == 0 {
		return []bump.Tag{"body", "map", "solid"} // Default fallback
	}

	tags := make([]bump.Tag, len(c.QueryTags))
	for i, tag := range c.QueryTags {
		tags[i] = bump.Tag(tag)
	}
	return tags
}

// ShouldIgnore checks if an entity should be filtered out.
func (c *Collider) ShouldIgnore(other entities.EntityId) bool {
	if c == nil || len(c.FilterOut) == 0 {
		return false
	}

	for _, id := range c.FilterOut {
		if id == other {
			return true
		}
	}
	return false
}

// AddFilter adds an entity to the ignore list.
func (c *Collider) AddFilter(entityID entities.EntityId) {
	if c == nil {
		return
	}

	// Check if already in list
	for _, id := range c.FilterOut {
		if id == entityID {
			return
		}
	}

	c.FilterOut = append(c.FilterOut, entityID)
}

// RemoveFilter removes an entity from the ignore list.
func (c *Collider) RemoveFilter(entityID entities.EntityId) {
	if c == nil || len(c.FilterOut) == 0 {
		return
	}

	for i, id := range c.FilterOut {
		if id == entityID {
			// Remove by swapping with last element
			c.FilterOut[i] = c.FilterOut[len(c.FilterOut)-1]
			c.FilterOut = c.FilterOut[:len(c.FilterOut)-1]
			return
		}
	}
}

// ClearFilters removes all entities from the ignore list.
func (c *Collider) ClearFilters() {
	if c == nil {
		return
	}
	c.FilterOut = c.FilterOut[:0]
}

// SetShape sets a custom collision shape (different from visual bounds).
func (c *Collider) SetShape(offsetX, offsetY, width, height float64) {
	if c == nil {
		return
	}
	c.OffsetX = offsetX
	c.OffsetY = offsetY
	c.Width = width
	c.Height = height
}

// ResetShape clears custom shape, reverting to Transform bounds.
func (c *Collider) ResetShape() {
	if c == nil {
		return
	}
	c.OffsetX = 0
	c.OffsetY = 0
	c.Width = 0
	c.Height = 0
}
