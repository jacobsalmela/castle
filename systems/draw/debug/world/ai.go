//go:build !release

package world

import (
	"game/components"
	"game/ecs"
	"game/resources"
	"game/systems/draw/debug/primitives"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawAIDebug renders AI debug information.
// Enable with Cmd+4. Shows:
// - Yellow boxes: AI target detection ranges (via DebugOverlay component)
// - Yellow lines: AI target connections
func DrawAIDebug(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryAI) {
		return
	}

	coords := primitives.NewWorldToScreen(cam)

	entities := world.EntitiesWith((*components.AI)(nil), (*components.Transform)(nil))
	for _, eid := range entities {
		ai := ecs.GetComponent[components.AI](world, eid)
		transform := ecs.GetComponent[components.Transform](world, eid)

		entityCenterX, entityCenterY := coords.Transform(
			transform.X+transform.W/2,
			transform.Y+transform.H/2,
		)

		if ai.TargetID != 0 {
			targetTransform := ecs.GetComponent[components.Transform](world, ai.TargetID)
			if targetTransform != nil {
				targetScreenX, targetScreenY := coords.Transform(
					targetTransform.X+targetTransform.W/2,
					targetTransform.Y+targetTransform.H/2,
				)

				// Solid yellow line for targeting
				vector.StrokeLine(screen, entityCenterX, entityCenterY, targetScreenX, targetScreenY,
					1, primitives.TargetLine, false)
			}
		}
	}

	overlayEntities := world.EntitiesWith((*components.DebugOverlay)(nil))
	for _, eid := range overlayEntities {
		overlay := ecs.GetComponent[components.DebugOverlay](world, eid)
		if overlay.Rect != nil {
			x, y, w, h := coords.TransformRect(overlay.Rect.X, overlay.Rect.Y, overlay.Rect.W, overlay.Rect.H)
			primitives.StrokeRect(screen, x, y, w, h, 1, overlay.Color)
		}
	}
}
