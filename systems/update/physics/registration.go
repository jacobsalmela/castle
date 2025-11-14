package physics

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
)

// RegisterInSpace updates the collision space with all entities that have a Collider component.
// This is called each frame before movement integration to ensure the spatial hash is up-to-date.
func RegisterInSpace(world *ecs.World, space ecs.CollisionSpace) {
	if world == nil || space == nil {
		return
	}

	// Query all entities with Collider component
	entityList := world.EntitiesWith((*components.Collider)(nil))

	for _, eid := range entityList {
		collider := ecs.GetComponent[components.Collider](world, eid)
		if collider == nil {
			continue
		}

		// Skip very fresh projectiles (SpawnGrace > 0) to allow Init to complete
		// This prevents physics from running before projectile initialization is done
		if IsProjectileSpawning(world, eid) {
			continue
		}

		transform := ecs.GetComponent[components.Transform](world, eid)
		if transform == nil {
			continue
		}

		// **FIX FOR ISSUE #2 (REVISED)**: Only remove dead entities if they're still in the space
		// Don't repeatedly try to remove them every frame (causes fall-through)
		if shouldRemoveFromSpace(world, eid) {
			// Check if entity is actually in the space before removing
			if space.Has(eid) {
				space.Remove(eid)
			}
			continue
		}

		registerEntityWithCollider(eid, transform, collider, space)
	}
}

// registerEntityWithCollider registers a single entity in the collision space using the Collider component.
// Phase 3: Updated to use Collider instead of BodyKinematics.
func registerEntityWithCollider(entityID entities.EntityId, transform *components.Transform, collider *components.Collider, space ecs.CollisionSpace) {
	if space == nil || collider == nil || transform == nil {
		return
	}

	// Calculate collision rect from transform + collider offset
	rect := collider.ToBumpRect(transform)

	// Convert Collider.Tags to bump.Tag
	tags := collider.ToBumpTags()

	// If entity is not solid and has no tags, remove it from space
	if !collider.Solid && len(tags) == 0 {
		if entityID != 0 {
			space.Remove(entityID)
		}
		return
	}

	// Default to "body" tag if no tags specified
	if len(tags) == 0 {
		tags = []bump.Tag{"body"}
	}

	// Register entity in collision space
	if entityID != 0 {
		space.Set(entityID, rect, tags...)
	}
}
