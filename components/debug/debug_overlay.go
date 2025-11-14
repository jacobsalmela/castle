// TODO: make pure data or move logic to systems, condense with other debug toggles.
package debug

import (
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"image/color"
)

// DebugOverlay provides visual debug information for an entity.
// Can be attached to any entity for debugging spatial bounds, ranges, or zones.
// All rendering logic is handled by systems/draw/debug_overlay.go
type DebugOverlay struct {
	// Rect is the rectangle to visualize in world space
	Rect *bump.Rect
	// Color is the overlay color (RGBA with alpha for transparency)
	Color color.RGBA
	// Label is an optional text label to display with the overlay
	Label string
}

// DefaultDebugColor is the standard yellow overlay color used for AI debug rects.
var DefaultDebugColor = color.RGBA{255, 255, 0, 75}

// SetDebugRect is a convenience function to create or update a DebugOverlay component
// on an entity. If the entity already has a DebugOverlay, it updates the Rect.
// If not, it creates a new DebugOverlay with the default yellow color.
func SetDebugRect(world *ecs.World, entityID entities.EntityId, rect *bump.Rect) {
	if world == nil || entityID == 0 {
		return
	}

	overlay := ecs.GetComponent[DebugOverlay](world, entityID)
	if overlay == nil {
		// Create new overlay with default color
		newOverlay := &DebugOverlay{
			Rect:  rect,
			Color: DefaultDebugColor,
		}
		world.AddComponent(entityID, newOverlay)
	} else {
		// Update existing overlay
		overlay.Rect = rect
	}
}

// SetDebugRectWithLabel is like SetDebugRect but also sets a text label.
func SetDebugRectWithLabel(world *ecs.World, entityID entities.EntityId, rect *bump.Rect, label string) {
	if world == nil || entityID == 0 {
		return
	}

	overlay := ecs.GetComponent[DebugOverlay](world, entityID)
	if overlay == nil {
		newOverlay := &DebugOverlay{
			Rect:  rect,
			Color: DefaultDebugColor,
			Label: label,
		}
		world.AddComponent(entityID, newOverlay)
	} else {
		overlay.Rect = rect
		overlay.Label = label
	}
}

// SetDebugRectWithColor is like SetDebugRect but allows custom color specification.
func SetDebugRectWithColor(world *ecs.World, entityID entities.EntityId, rect *bump.Rect, col color.RGBA) {
	if world == nil || entityID == 0 {
		return
	}

	overlay := ecs.GetComponent[DebugOverlay](world, entityID)
	if overlay == nil {
		newOverlay := &DebugOverlay{
			Rect:  rect,
			Color: col,
		}
		world.AddComponent(entityID, newOverlay)
	} else {
		overlay.Rect = rect
		overlay.Color = col
	}
}
