package combat

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/systems/update/entities/animation"
	"game/systems/update/state"
)

// AttackIntentConsume converts one-shot AttackIntent components into AttackActive
// components for frame-by-frame processing. This system provides a simple entry
// point for entities to trigger attacks through the Pure ECS combat pipeline.
func AttackIntentConsume(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	intents := world.EntitiesWith((*components.AttackIntent)(nil))

	for _, eid := range intents {
		intent := ecs.GetComponent[components.AttackIntent](world, eid)
		if intent == nil || intent.Damage == 0 {
			continue
		}

		// Resolve area, apply the attack, deduct stamina, and clear the intent.
		area := resolveAttackArea(world, eid)
		applyAttackIntent(world, entities.EntityId(eid), area, intent)
		deductStamina(world, eid, intent)

		// Reset intent so it behaves as a one-shot (Pure ECS - inline reset logic)
		intent.AttackTag = ""
		intent.Damage = 0
		intent.StaminaDamage = 0
		intent.ReactForce = 0
		intent.PushForce = 0
		intent.Mult = 0
		intent.FilterOut = nil
	}
}

// resolveAttackArea returns the attack area for an entity by preferring an
// Animation frame slice named components.HitboxSliceName, falling back to a conservative
// Transform-anchored 8x8 rect when no slice is available.
func resolveAttackArea(world *ecs.World, eid entities.EntityId) bump.Rect {
	var area bump.Rect
	if anim := ecs.GetComponent[components.Animation](world, eid); anim != nil {
		// Get flip state for hitbox positioning
		facing := ecs.GetComponent[components.Facing](world, eid)
		flipX := false
		if facing != nil {
			flipX = facing.FlipX
		}

		if rect, err := animation.ExtractSlice(anim, components.HitboxSliceName, flipX, false); err == nil {
			return rect
		}
	}

	// Fallback: anchored to Transform
	area.W = 8
	area.H = 8
	if transform := ecs.GetComponent[components.Transform](world, eid); transform != nil {
		flip := false
		if facing := ecs.GetComponent[components.Facing](world, eid); facing != nil {
			flip = facing.FlipX
		}
		if flip {
			area.X = transform.X - area.W
		} else {
			area.X = transform.X + transform.W
		}
		area.Y = transform.Y
	}
	return area
}

// applyAttackIntent converts the one-shot intent into an AttackActive component
// for frame-by-frame processing by the AttackInProgress system.
func applyAttackIntent(world *ecs.World, attacker entities.EntityId, area bump.Rect, intent *components.AttackIntent) {
	// Convert the one-shot intent into an AttackActive component so the
	// AttackInProgress system can process the attack across animation frames
	// and coalesce contacts (tracking which targets have already been hit).
	if intent == nil || intent.Damage == 0 {
		return
	}
	// If an AttackActive already exists, overwrite it with the new intent.
	aa := &components.AttackActive{
		AttackTag:     intent.AttackTag,
		Damage:        intent.Damage,
		StaminaDamage: intent.StaminaDamage,
		ReactForce:    intent.ReactForce,
		PushForce:     intent.PushForce,
		Mult:          intent.Mult,
		FilterOut:     intent.FilterOut,
	}
	// Attach to the entity so the AttackInProgress system will pick it up.
	world.AddComponent(attacker, aa)
}

// deductStamina deducts stamina from an entity when the intent specifies a stamina cost.
// deductStamina reduces the entity's stamina by the attack's stamina cost.
func deductStamina(world *ecs.World, eid entities.EntityId, intent *components.AttackIntent) {
	if intent == nil || intent.StaminaDamage == 0 {
		return
	}
	// Pure ECS Stamina component
	state.AddStamina(world, eid, -intent.StaminaDamage)
}
