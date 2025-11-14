//go:build !release

package world

import (
	"game/components"
	"game/ecs"
	"game/resources"
	"game/systems/draw/debug/primitives"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawCollisionDebug renders collision boxes and collision events.
// Enable with Cmd+1. Shows:
// - Cyan boxes: Entity collision rectangles
// - Yellow circles: Collision impact points (with fade trail)
// - Red lines: Collision normal vectors
// - Green lines: Connection between colliding entities
func DrawCollisionDebug(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryCollision) {
		return
	}

	coords := primitives.NewWorldToScreen(cam)

	drawCollisionBoxes(world, screen, coords)
	drawCollisionEvents(world, screen, coords)
}

func drawCollisionBoxes(world *ecs.World, screen *ebiten.Image, coords primitives.WorldToScreen) {
	entities := world.EntitiesWith((*components.Transform)(nil), (*components.Physics)(nil))
	for _, eid := range entities {
		transform := ecs.GetComponent[components.Transform](world, eid)
		x, y, w, h := coords.TransformRect(transform.X, transform.Y, transform.W, transform.H)
		primitives.StrokeRect(screen, x, y, w, h, 1, primitives.CollisionBox)
	}
}

func drawCollisionEvents(world *ecs.World, screen *ebiten.Image, coords primitives.WorldToScreen) {
	events := ecs.Resource[resources.CollisionEventQueue](world)
	if events == nil {
		return
	}

	now := time.Now()
	const lifetime = 2000 * time.Millisecond

	for _, event := range events.Recent() {
		age := now.Sub(event.Timestamp)
		ageRatio := float64(age) / float64(lifetime)
		alpha := 1.0 - ageRatio

		if alpha <= 0 {
			continue
		}

		screenX, screenY := coords.Transform(event.ItemX, event.ItemY)

		size := 8.0 - (5.0 * ageRatio)
		if ageRatio < 0.25 {
			pulseAmount := 2.0 * (1.0 - ageRatio*4)
			size += pulseAmount * math.Sin(float64(age.Milliseconds())/80.0)
		}

		brightness := uint8(255 * alpha)
		highlightColor := color.RGBA{brightness, brightness, 0, brightness}
		primitives.FillCircle(screen, screenX, screenY, float32(size), highlightColor)

		if alpha > 0.6 {
			normalLen := float32(15.0)
			normalAlpha := uint8(255 * ((alpha - 0.6) / 0.4))
			normalColor := color.RGBA{255, 100, 100, normalAlpha}
			vector.StrokeLine(screen, screenX, screenY,
				screenX+float32(event.NormalX)*normalLen,
				screenY+float32(event.NormalY)*normalLen,
				2, normalColor, false)
		}

		if alpha > 0.8 {
			connectionAlpha := uint8(150 * ((alpha - 0.8) / 0.2))
			connectionColor := color.RGBA{100, 255, 100, connectionAlpha}
			otherScreenX, otherScreenY := coords.Transform(event.OtherX, event.OtherY)
			vector.StrokeLine(screen, screenX, screenY, otherScreenX, otherScreenY, 1, connectionColor, false)
		}
	}
}
