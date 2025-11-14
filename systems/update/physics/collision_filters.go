package physics

import (
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
)

// shouldIgnoreEntity checks if an entity should be filtered out from collisions.
func shouldIgnoreEntity(entityID entities.EntityId, filterOut []entities.EntityId) bool {
	for _, ignored := range filterOut {
		if entityID == ignored {
			return true
		}
	}
	return false
}

// getEntityCollisionType determines collision type for entity-entity collisions.
func getEntityCollisionType(entityID entities.EntityId, space ecs.CollisionSpace) (bump.ColType, bool) {
	// Solid entities (walls, doors) block movement
	if space.Has(entityID, bump.Tag("solid")) {
		return bump.Slide, true
	}

	// Bodies (player, enemies) block each other
	if space.Has(entityID, bump.Tag("body")) {
		return bump.Slide, true
	}

	// Other entities are detected but don't block
	return bump.Cross, true
}

// getPlatformCollisionType determines collision type for platform interactions.
func getPlatformCollisionType(item, other bump.Item, space ecs.CollisionSpace,
	dropping bool) (bump.ColType, bool) {
	// Handle ladder tiles: non-top ladder tiles are fully passthrough
	if space.Has(other, bump.Tag("ladder")) && !space.Has(other, bump.Tag("ladder_top")) {
		return bump.Cross, true
	}

	// Non-platforms are solid
	if !space.Has(other, bump.Tag("passthrough")) {
		return bump.Slide, true
	}

	// Check if entity is above platform
	itemRect, otherRect := space.Rect(item), space.Rect(other)
	isAbovePlatform := itemRect.Y+itemRect.H <= otherRect.Y

	// Stand on platform if above and not dropping
	if !dropping && isAbovePlatform {
		return bump.Slide, true
	}

	// Pass through platform otherwise
	return bump.Cross, true
}
