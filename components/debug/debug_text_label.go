// TODO: verify if these are being used, pair with other logic or condense with other debug toggles.
package debug

import (
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// DebugTextLabel represents a text label to render with device-pixel fonts.
// This replaces pkg/overlay functionality with Pure ECS architecture.
// All rendering logic is handled by systems/draw/debug_text_labels.go
//
// Device-pixel rendering ensures crisp text regardless of viewport scale/DPR.
// Labels can be anchored to logical coordinates (world space) or device coordinates (screen space).
//
// Usage:
//   - Create entity: entityID := world.NewEntity()
//   - Add DebugTextLabel: world.AddComponent(entityID, &DebugTextLabel{Text: "HP: 100", ...})
//   - System will render it at the specified coordinates with device-pixel font
//
// Stacking:
//   - Multiple labels at the same logical position are automatically stacked
//   - Uses fixed device-pixel gap (e.g., 12px) for consistent spacing
//   - Stack is centered on the anchor point with small upward bias
type DebugTextLabel struct {
	Text     string
	LogicalX float64 // X position in logical (world) coordinates
	LogicalY float64 // Y position in logical (world) coordinates
	Align    text.Align
	DevicePx int // Desired device pixel height (e.g., 16)
}
