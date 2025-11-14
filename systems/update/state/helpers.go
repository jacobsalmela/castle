package state

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

// AddHealth modifies the Health component of an entity by the given amount.
//
// Parameters:
//   - world: ECS world instance
//   - entityID: Entity to modify
//   - amount: Health change (positive = heal, negative = damage)
func AddHealth(world *ecs.World, entityID entities.EntityId, amount float64) {
	if world == nil || entityID == 0 {
		return
	}

	health := ecs.GetComponent[components.Health](world, entityID)
	if health == nil {
		return
	}

	health.Current += amount
	if health.Current > health.Max {
		health.Current = health.Max
	}
	if health.Current < 0 {
		health.Current = 0
	}

	// Immediately disable gravity when entity dies to prevent falling through ground
	if health.Current <= 0 {
		if physics := ecs.GetComponent[components.Physics](world, entityID); physics != nil {
			physics.GravityEnabled = false
			physics.Velocity.X = 0
			physics.Velocity.Y = 0
		}
	}
}

// AddStamina modifies the Stamina component of an entity by the given amount.
//
// Parameters:
//   - world: ECS world instance
//   - entityID: Entity to modify
//   - amount: Stamina change (positive = restore, negative = drain)
func AddStamina(world *ecs.World, entityID entities.EntityId, amount float64) {
	if world == nil || entityID == 0 {
		return
	}

	stamina := ecs.GetComponent[components.Stamina](world, entityID)
	if stamina == nil {
		return
	}

	stamina.Current += amount
	if stamina.Current > stamina.Max {
		stamina.Current = stamina.Max
	}
	if stamina.Current < 0 {
		stamina.Current = 0
	}
}

// AddPoise modifies the Poise component of an entity by the given amount.
//
// Parameters:
//   - world: ECS world instance
//   - entityID: Entity to modify
//   - amount: Poise change (positive = restore, negative = damage)
func AddPoise(world *ecs.World, entityID entities.EntityId, amount float64) {
	if world == nil || entityID == 0 {
		return
	}

	poise := ecs.GetComponent[components.Poise](world, entityID)
	if poise == nil {
		return
	}

	poise.Current += amount
	if poise.Current > poise.Max {
		poise.Current = poise.Max
	}
	if poise.Current < 0 {
		poise.Current = 0
	}
}

// AddExperience modifies the Experience component of an entity by the given amount.
//
// Parameters:
//   - world: ECS world instance
//   - entityID: Entity to modify
//   - amount: Experience points to add
func AddExperience(world *ecs.World, entityID entities.EntityId, amount int) {
	if world == nil || entityID == 0 {
		return
	}

	exp := ecs.GetComponent[components.Experience](world, entityID)
	if exp == nil {
		return
	}

	exp.Points += amount
}
