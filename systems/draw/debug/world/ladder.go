//go:build !release

package world

import (
	"game/components"
	"game/ecs"
	"game/resources"
	"game/systems/draw/debug/primitives"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawLadderDebugOverlay draws ladder tiles when Shift+D debug toggle is enabled.
func DrawLadderDebugOverlay(world *ecs.World, cam primitives.CameraProvider, screen *ebiten.Image, vp *components.ViewPort) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryLadder) {
		return
	}

	ladderRegistry := ecs.Resource[resources.LadderRegistry](world)
	if ladderRegistry == nil || cam == nil || screen == nil || vp == nil {
		return
	}

	// Get camera position
	cx, cy := cam.Position()

	// Get debug color for ladders
	debugCategories := ecs.Resource[resources.DebugCategories](world)
	ladderColor := color.RGBA{0, 255, 0, 100} // Default green semi-transparent
	if debugCategories != nil {
		if c, ok := debugCategories.GetColor("Ladder"); ok {
			ladderColor = c
		}
	}

	// Outline color (slightly more opaque)
	outlineColor := color.RGBA{ladderColor.R, ladderColor.G, ladderColor.B, 255}

	// Draw all ladders
	for _, ladder := range ladderRegistry.Ladders {
		// Convert world coordinates to screen coordinates (apply DPR)
		screenX := float32((ladder.X - cx) * vp.DPR)
		screenY := float32((ladder.Y - cy) * vp.DPR)
		width := float32(ladder.W * vp.DPR)
		height := float32(ladder.H * vp.DPR)

		// Draw filled rectangle
		vector.DrawFilledRect(screen, screenX, screenY, width, height, ladderColor, false)

		// Draw outline
		vector.StrokeRect(screen, screenX, screenY, width, height, 1, outlineColor, false)
	}
}
