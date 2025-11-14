//go:build !release

package world

import (
	"game/components"
	"game/ecs"
	"game/systems/draw/debug/primitives"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawVisualPrimitives renders all DebugVisual components (rects, lines, circles, text).
// These are temporary entities created by debug systems each frame.
func DrawVisualPrimitives(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	coords := primitives.NewWorldToScreen(cam)

	entities := world.EntitiesWith((*components.DebugVisual)(nil))
	for _, eid := range entities {
		visual := ecs.GetComponent[components.DebugVisual](world, eid)

		switch visual.Type {
		case components.DebugVisualRect:
			drawVisualRect(screen, visual, coords)
		case components.DebugVisualLine:
			drawVisualLine(screen, visual, coords)
		case components.DebugVisualCircle:
			drawVisualCircle(screen, visual, coords)
		case components.DebugVisualText:
			drawVisualText(screen, visual, coords)
		}
	}
}

func drawVisualRect(screen *ebiten.Image, visual *components.DebugVisual, coords primitives.WorldToScreen) {
	screenX, screenY, width, height := coords.TransformRect(visual.X, visual.Y, visual.W, visual.H)

	if visual.Fill {
		primitives.FillRect(screen, float64(screenX), float64(screenY), float64(width), float64(height), visual.Color)
	} else {
		primitives.StrokeRect(screen, screenX, screenY, width, height, 1, visual.Color)
	}
}

func drawVisualLine(screen *ebiten.Image, visual *components.DebugVisual, coords primitives.WorldToScreen) {
	x1, y1 := coords.Transform(visual.X, visual.Y)
	x2, y2 := coords.Transform(visual.W, visual.H)

	vector.StrokeLine(screen, x1, y1, x2, y2, 2, visual.Color, false)
}

func drawVisualCircle(screen *ebiten.Image, visual *components.DebugVisual, coords primitives.WorldToScreen) {
	centerX, centerY := coords.Transform(visual.X, visual.Y)
	radius := float32(visual.W)

	if visual.Fill {
		primitives.FillCircle(screen, centerX, centerY, radius, visual.Color)
	} else {
		primitives.StrokeCircle(screen, centerX, centerY, radius, visual.Color)
	}
}

func drawVisualText(screen *ebiten.Image, visual *components.DebugVisual, coords primitives.WorldToScreen) {
	// Text rendering via DebugTextLabel component instead
	_ = screen
	_ = visual
	_ = coords
}
