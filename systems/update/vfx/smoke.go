package vfx

import (
	"image/color"

	"game/components"
	"game/ecs"
)

// UpdateSmoke advances tween, updates transform/color, and removes when done.
func UpdateSmoke(world *ecs.World, _ interface{}, dt float64) {
	if world == nil || dt == 0 {
		return
	}
	for _, eid := range world.EntitiesWith((*components.Smoke)(nil)) {
		sm := ecs.GetComponent[components.Smoke](world, eid)
		if sm == nil {
			continue
		}
		render := ecs.GetComponent[components.Render](world, eid)
		if render == nil {
			continue
		}
		sm.Tween.Update(dt)
		prog := sm.Tween.Value()
		if sm.Tween.IsDone() {
			// Cleanup: destroy entity when animation completes
			world.DestroyEntity(eid)
			continue
		}
		// Update Transform position
		newX := sm.StartX + prog*sm.TargetX
		newY := sm.StartY + prog*sm.TargetY
		if t := ecs.GetComponent[components.Transform](world, eid); t != nil {
			t.X, t.Y = newX, newY
		}
		render.R += sm.RotationRate * dt
		alpha := uint8(100 + (255-100)*(1-prog))
		render.ColorScale = color.RGBA{alpha, alpha, alpha, alpha}
	}
}
