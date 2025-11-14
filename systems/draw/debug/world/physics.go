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

// DrawBodyDebug renders body physics state for entities with Physics component.
// Enable with Cmd+2. Shows:
// - Yellow boxes: Entity collision boundaries
// - Green arrows: Velocity vector
// - Green/Red line above entity: Grounded status
func DrawBodyDebug(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryPhysics) {
		return
	}

	coords := primitives.NewWorldToScreen(cam)

	entities := world.EntitiesWith((*components.Transform)(nil), (*components.Physics)(nil))
	for _, eid := range entities {
		transform := ecs.GetComponent[components.Transform](world, eid)
		physics := ecs.GetComponent[components.Physics](world, eid)

		if transform == nil || physics == nil {
			continue
		}

		screenX, screenY, width, height := coords.TransformRect(transform.X, transform.Y, transform.W, transform.H)

		// Draw yellow collision box
		primitives.StrokeRect(screen, screenX, screenY, width, height, 1, primitives.PhysicsBox)

		// Draw velocity vector as green arrow
		if physics.Velocity.X != 0 || physics.Velocity.Y != 0 {
			centerX := screenX + width/2
			centerY := screenY + height/2
			endX := centerX + float32(physics.Velocity.X)*0.1
			endY := centerY + float32(physics.Velocity.Y)*0.1

			vector.StrokeLine(screen, centerX, centerY, endX, endY, 2, primitives.VelocityVector, false)
		}

		// Draw grounded status as colored line above entity
		statusY := screenY - 5
		var statusColor = primitives.AirborneLine
		if physics.Grounded {
			statusColor = primitives.GroundedLine
		}

		vector.StrokeLine(screen, screenX, statusY, screenX+width, statusY, 3, statusColor, false)
	}
}

// DrawPhysicsDebug renders detailed physics information.
// Enable with Cmd+9. Shows:
// - White boxes: Collision boundaries
// - Cyan arrows: Velocity vectors
func DrawPhysicsDebug(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryPhysics) {
		return
	}

	coords := primitives.NewWorldToScreen(cam)

	entities := world.EntitiesWith((*components.Transform)(nil), (*components.Physics)(nil))
	for _, eid := range entities {
		transform := ecs.GetComponent[components.Transform](world, eid)
		physics := ecs.GetComponent[components.Physics](world, eid)

		if transform == nil || physics == nil {
			continue
		}

		screenX, screenY, width, height := coords.TransformRect(transform.X, transform.Y, transform.W, transform.H)

		// Draw white collision box
		primitives.StrokeRect(screen, screenX, screenY, width, height, 1, primitives.PhysicsWhite)

		// Draw velocity vector
		centerX := screenX + width/2
		centerY := screenY + height/2
		velEndX := centerX + float32(physics.Velocity.X)*0.2
		velEndY := centerY + float32(physics.Velocity.Y)*0.2

		vector.StrokeLine(screen, centerX, centerY, velEndX, velEndY, 2, primitives.PhysicsCyan, false)
	}
}
