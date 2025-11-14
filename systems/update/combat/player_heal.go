package combat

import (
	"game/components"
	"game/ecs"
	"game/systems/update/entities/animation"
)

// UpdatePlayerHeal processes player healing based on input intents.
func UpdatePlayerHeal(world *ecs.World) {
	if world == nil {
		return
	}

	for _, eid := range world.EntitiesWith((*components.Player)(nil), (*components.ActionIntents)(nil)) {
		player := ecs.GetComponent[components.Player](world, eid)
		intents := ecs.GetComponent[components.ActionIntents](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		healing := ecs.GetComponent[components.Healing](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)

		if player == nil || intents == nil || health == nil || healing == nil || anim == nil {
			continue
		}

		// Check if heal intent is set
		if !intents.Heal {
			continue
		}

		// Clear intent flag immediately (consumed)
		intents.Heal = false
		// Check if heal charges available
		if healing.Count <= 0 {
			continue
		}

		// Don't process if already in consume animation
		if anim.State == components.ConsumeTag {
			continue
		}

		// Consume one heal charge
		healing.Count--

		// Get heal amount from healing component
		healAmount := healing.HealAmount
		if healAmount == 0 {
			healAmount = 20.0 // Fallback to default
		}

		// Apply healing immediately (knight.json has no healbox slice)
		health.Current += healAmount
		if health.Current > health.Max {
			health.Current = health.Max
		}

		// Update attack multiplier based on heals used
		if attackMult := ecs.GetComponent[components.AttackMultiplier](world, eid); attackMult != nil {
			healsUsed := healing.MaxCount - healing.Count
			attackMult.Current = attackMult.PerHeal * float64(healsUsed)
		}

		// Trigger consume animation (visual only)
		animation.SetAnimationState(anim, components.ConsumeTag)
	}
}
