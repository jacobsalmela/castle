//go:build !release

package debug

import (
	"fmt"
	"game/assets/fonts"
	"game/components"
	"game/ecs"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// TODO: verify if this is used or not
// drawTextLabels renders all DebugTextLabel components with device-pixel fonts.
// This replaces pkg/overlay functionality with Pure ECS architecture.
// Labels at the same logical position are automatically stacked with consistent spacing.
func drawTextLabels(world *ecs.World, screen *ebiten.Image, vp *components.ViewPort) {
	if vp == nil {
		return
	}

	// Get all text label entities
	entities := world.EntitiesWith((*components.DebugTextLabel)(nil))
	if len(entities) == 0 {
		return
	}

	// Group labels by their device anchor (rounded device coords)
	const stackGapDevicePx = 12
	const upwardBias = -8

	type labelData struct {
		label *components.DebugTextLabel
		face  text.Face
		dx    float64
		dy    float64
	}

	groups := make(map[string][]labelData)

	for _, eid := range entities {
		label := ecs.GetComponent[components.DebugTextLabel](world, eid)
		face := fonts.DeviceFace(label.DevicePx, vp.DPR)
		dx, dy, _ := vp.LogicalToDevice(label.LogicalX, label.LogicalY)

		// Round to device pixel
		dx = math.Round(dx)
		dy = math.Round(dy)

		key := fmt.Sprintf("%d:%d", int(dx), int(dy))
		groups[key] = append(groups[key], labelData{
			label: label,
			face:  face,
			dx:    dx,
			dy:    dy,
		})
	}

	// Draw each group centered on the device anchor
	for _, arr := range groups {
		// Total height in device pixels for the stacked texts
		total := len(arr) * stackGapDevicePx
		// Top offset to center the block on the anchor
		topOffset := float64(-total / 2)
		// Apply an upward bias so the block sits above the sprite
		topOffset += upwardBias

		for i, item := range arr {
			dx := item.dx
			dy := item.dy + topOffset + float64(i*stackGapDevicePx)

			opts := &text.DrawOptions{}
			opts.GeoM.Reset()
			opts.GeoM.Translate(dx, dy)
			opts.PrimaryAlign = item.label.Align
			text.Draw(screen, item.label.Text, item.face, opts)
		}
	}
}
