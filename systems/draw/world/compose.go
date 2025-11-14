package world

import (
	"game/components"
	"game/resources"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// ComposeRenderQueue drains the ECS render queue and draws all commands
// to the screen and normal map targets in layer order.
// This is the Pure ECS composition phase for queue-based rendering.
// Exported so parent draw.go can use it for HUD composition.
func ComposeRenderQueue(screen, normalMap *ebiten.Image, queue *resources.RenderQueue) {
	if queue == nil {
		return
	}

	// Sort commands by layer (lowest to highest)
	queue.SortByLayer()

	// Draw all queued commands
	for _, cmd := range queue.Commands() {
		if cmd.Image == nil {
			continue
		}

		// Draw to screen if requested
		if (cmd.TargetType & resources.TargetScreen) != 0 {
			drawToScreen(screen, cmd)
		}

		// Draw to normal map if requested
		if (cmd.TargetType & resources.TargetNormal) != 0 {
			drawToNormal(normalMap, cmd)
		}
	}

	// Clear the queue for the next frame
	queue.Clear()
}

// drawToScreen renders a command to the screen buffer with color scaling.
func drawToScreen(screen *ebiten.Image, cmd resources.RenderCommand) {
	if screen == nil {
		return
	}

	op := &ebiten.DrawImageOptions{GeoM: cmd.GeoM}
	if cmd.ColorScale != nil {
		op.ColorScale.ScaleWithColor(cmd.ColorScale)
	}
	screen.DrawImage(cmd.Image, op)
}

// drawToNormal renders a command to the normal map buffer using the fill mask.
func drawToNormal(normalMap *ebiten.Image, cmd resources.RenderCommand) {
	if normalMap == nil {
		return
	}

	op := &colorm.DrawImageOptions{GeoM: cmd.GeoM}
	colorm.DrawImage(normalMap, cmd.Image, components.FillNormalMaskColorM, op)
}
