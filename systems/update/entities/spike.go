package entities

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/resources"
	"game/systems/update/physics"
)

const (
	// Damage properties
	spikeDamageAmount = 20 // Damage dealt per contact
)

// UpdateSpike processes spike hazard entities for damage application.
// Uses deterministic cooldown tracking via SpikeCooldown resource.
func UpdateSpike(world *ecs.World, _ interface{}, dt float64) {
	if world == nil || dt == 0 {
		return
	}

	// Get cooldown resource
	cooldown := ecs.Resource[resources.SpikeCooldown](world)
	if cooldown == nil {
		return
	}

	// Get deterministic game time from world resource
	timeCtrl := ecs.Resource[resources.TimeControl](world)
	if timeCtrl == nil {
		return
	}
	currentTime := timeCtrl.ElapsedTime

	// Clean up expired cooldowns
	cooldown.CleanupExpired(currentTime)

	// Process all spike entities
	spikes := world.EntitiesWith((*components.Spike)(nil), (*components.Transform)(nil))
	for _, eid := range spikes {
		processSpike(world, eid, cooldown, currentTime)
	}
}

// processSpike handles damage application for a single spike entity.
func processSpike(world *ecs.World, eid entities.EntityId, cooldown *resources.SpikeCooldown, currentTime float64) {
	// Get spike transform for area query
	transform := ecs.GetComponent[components.Transform](world, eid)
	if transform == nil {
		return
	}

	// Early exit if no bodies nearby (performance optimization)
	if !hasBodiesNearby(world, transform) {
		return
	}

	// Query entities in spike area and apply damage
	spikeArea := bump.NewRect(transform.X, transform.Y, transform.W, transform.H)
	contacted := querySpikeContacts(world, eid, spikeArea)

	// Apply damage to entities not on cooldown
	for _, targetID := range contacted {
		if cooldown.CanDamage(targetID, currentTime) {
			// Apply damage to target
			applySpikeDamage(world, targetID)

			// Record damage time to start cooldown
			cooldown.RecordDamage(targetID, currentTime)
		}
	}
}

// hasBodiesNearby checks if any bodies are in the spike's area.
// Returns true if at least one body entity is within the spike bounds.
func hasBodiesNearby(world *ecs.World, transform *components.Transform) bool {
	area := bump.NewRect(transform.X, transform.Y, transform.W, transform.H)
	space := physics.GetCollisionSpace(world)
	bodies := physics.QueryItems(space, entities.EntityId(0), area, "body")
	return len(bodies) > 0
}

// querySpikeContacts finds all entities touching the spike area.
// Returns a list of entity IDs that are in contact with the spike.
func querySpikeContacts(world *ecs.World, spikeID entities.EntityId, spikeArea bump.Rect) []entities.EntityId {
	var contacted []entities.EntityId

	// Iterate through all entities with hitboxes
	for _, targetID := range world.EntitiesWith((*components.Hitbox)(nil), (*components.Transform)(nil)) {
		// Skip self
		if targetID == spikeID {
			continue
		}

		// Get target components
		hitbox := ecs.GetComponent[components.Hitbox](world, targetID)
		transform := ecs.GetComponent[components.Transform](world, targetID)
		if hitbox == nil || transform == nil || len(hitbox.Boxes) == 0 {
			continue
		}

		// Check if any hurtbox overlaps with spike area
		for _, box := range hitbox.Boxes {
			// Only check Hit contact type (hurtboxes)
			if box.ResolveContact() != components.Hit {
				continue
			}

			// Convert box to world coordinates
			boxRect := box.ToBumpRect(transform.X, transform.Y)

			// Check if box overlaps with spike area
			if bump.Overlaps(spikeArea, boxRect) {
				contacted = append(contacted, targetID)
				break // Only count each entity once
			}
		}
	}

	return contacted
}

// applySpikeDamage applies spike damage to a target entity.
func applySpikeDamage(world *ecs.World, targetID entities.EntityId) {
	// Get health component
	health := ecs.GetComponent[components.Health](world, targetID)
	if health == nil {
		return
	}

	// Apply damage
	health.Current -= spikeDamageAmount
	if health.Current < 0 {
		health.Current = 0
	}

	// Enqueue hit event for other systems (visual effects, etc.)
	eventQueue := ecs.Resource[resources.EventQueue](world)
	if eventQueue != nil {
		eventQueue.Hits = append(eventQueue.Hits, resources.HitEvent{
			Attacker: 0, // No attacker for environmental hazards
			Target:   targetID,
			Contact:  int(components.Hit),
			Damage:   spikeDamageAmount,
			Handled:  false,
		})
	}
}
