package vfx

import (
	"game/components"
	"game/ecs"
)

// UpdateDamageNumber advances tween, updates position, and removes when done.
func UpdateDamageNumber(world *ecs.World, dt float64) {
	if world == nil || dt == 0 {
		return
	}

	entities := world.EntitiesWith((*components.DamageNumber)(nil))
	for _, eid := range entities {
		dmg := ecs.GetComponent[components.DamageNumber](world, eid)
		if dmg == nil {
			continue
		}

		// Update animation tween
		dmg.Tween.Update(dt)
		prog := dmg.Tween.Value()

		// Remove entity when animation completes
		if dmg.Tween.IsDone() {
			world.DestroyEntity(eid)
			continue
		}

		// Update position along drift trajectory
		newX := dmg.StartX + prog*dmg.TargetX
		newY := dmg.StartY + prog*dmg.TargetY

		transform := ecs.GetComponent[components.Transform](world, eid)
		if transform != nil {
			transform.X = newX
			transform.Y = newY
		}
	}
}
