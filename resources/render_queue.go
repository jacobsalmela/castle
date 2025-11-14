package resources

import (
	"image/color"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// RenderTarget specifies which render buffer(s) to draw to.
type RenderTarget int

const (
	// TargetScreen renders to the main screen buffer.
	TargetScreen RenderTarget = 1 << iota
	// TargetNormal renders to the normal map buffer.
	TargetNormal
	// TargetBoth renders to both screen and normal map buffers.
	TargetBoth = TargetScreen | TargetNormal
)

// RenderCommand represents a single draw operation queued for rendering.
// Commands are collected during the update phase and executed in layer order
// during the compose phase.
type RenderCommand struct {
	// Image is the source image to draw.
	Image *ebiten.Image
	// TargetType specifies which buffer(s) to render to.
	TargetType RenderTarget
	// Layer determines draw order (higher layers drawn on top).
	Layer int
	// GeoM is the geometry transformation matrix for positioning/rotation/scale.
	GeoM ebiten.GeoM
	// ColorScale tints the image when rendering to screen. Ignored for normal map.
	ColorScale color.Color
}

// RenderQueue collects draw commands from ECS systems during the frame,
// sorts them by layer, and provides an iterator for the compose phase.
// This is the Pure ECS rendering system for queue-based rendering.
type RenderQueue struct {
	commands []RenderCommand
}

// NewRenderQueue creates an empty render queue.
func NewRenderQueue() *RenderQueue {
	return &RenderQueue{
		commands: make([]RenderCommand, 0, 256), // Pre-allocate for typical frame
	}
}

// Push adds a render command to the queue.
// Commands are not immediately drawn; call SortByLayer and then iterate Commands.
func (q *RenderQueue) Push(cmd RenderCommand) {
	if q == nil || cmd.Image == nil {
		return
	}
	q.commands = append(q.commands, cmd)
}

// PushMapTile adds a map tile render command to the queue.
// This is a convenience method for map rendering that implements tilemap.MapRenderQueue interface.
// It creates a RenderCommand with the provided parameters.
func (q *RenderQueue) PushMapTile(image *ebiten.Image, targetType int, layer int, geom ebiten.GeoM) {
	if q == nil || image == nil {
		return
	}
	q.commands = append(q.commands, RenderCommand{
		Image:      image,
		TargetType: RenderTarget(targetType),
		Layer:      layer,
		GeoM:       geom,
	})
}

// Clear removes all queued commands. Call this after composing a frame.
func (q *RenderQueue) Clear() {
	if q == nil {
		return
	}
	q.commands = q.commands[:0] // Preserve capacity
}

// SortByLayer orders commands from lowest to highest layer.
// Must be called before iterating Commands to ensure correct draw order.
func (q *RenderQueue) SortByLayer() {
	if q == nil {
		return
	}
	slices.SortStableFunc(q.commands, func(a, b RenderCommand) int {
		return a.Layer - b.Layer
	})
}

// Commands returns the queued commands in their current order.
// Call SortByLayer first to ensure proper layering.
func (q *RenderQueue) Commands() []RenderCommand {
	if q == nil {
		return nil
	}
	return q.commands
}

// Len returns the number of queued commands.
func (q *RenderQueue) Len() int {
	if q == nil {
		return 0
	}
	return len(q.commands)
}
