// TODO: make pure data or move logic to systems, condense with other debug toggles.
package debug

import (
	"image/color"
)

// DebugVisual represents a visual debug primitive (line, rect, circle, text).
// These are temporary entities created each frame by debug systems and cleared at frame start.
// All rendering logic is handled by systems/draw/debug_visual_primitives.go
//
// Usage:
//   - Create entity: entityID := world.NewEntity()
//   - Add Transform: world.AddComponent(entityID, &Transform{X: x, Y: y})
//   - Add DebugVisual: world.AddComponent(entityID, &DebugVisual{Type: DebugVisualRect, ...})
//   - System will render it and entity will be cleared next frame
type DebugVisual struct {
	Type  DebugVisualType
	Color color.RGBA

	// For rect: X, Y = position, W, H = size
	// For line: X, Y = start, W, H = end (reusing W/H for x2, y2)
	// For circle: X, Y = center, W = radius (H unused)
	// For text: X, Y = position
	X, Y, W, H float64
	Fill       bool // Whether to fill (rect/circle only)

	// For text only
	Text  string
	Scale float64 // Text scale multiplier (default 1.0)
}

// DebugVisualType represents the type of debug primitive to draw.
type DebugVisualType string

const (
	DebugVisualRect   DebugVisualType = "rect"
	DebugVisualLine   DebugVisualType = "line"
	DebugVisualCircle DebugVisualType = "circle"
	DebugVisualText   DebugVisualType = "text"
)
