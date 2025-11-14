package ecs

import "game/pkg/bump"

// CollisionSpace exposes the bump operations required by physics and game systems.
// This interface abstracts the spatial hash collision library, allowing systems
// to register entities, query collisions, and perform movement integration without
// direct dependency on bump internals.
//
// Purpose: Architectural boundary between ECS world and collision detection.
// This enables:
//   - Testing: Mock CollisionSpace for unit tests
//   - Flexibility: Swap collision implementations without changing systems
//   - Decoupling: Systems don't depend on concrete bump.Space type
//
// Implementations:
//   - *bump.Space implements this interface natively
//
// Used by:
//   - systems/update/physics/ - Physics simulation systems
//   - Game systems that need spatial queries
type CollisionSpace interface {
	// Set registers or updates an item in the collision space with the given rect and tags.
	Set(item bump.Item, rect bump.Rect, tags ...bump.Tag)

	// Move attempts to move an item to a goal position, applying collision filtering.
	// Returns the final position and any collisions encountered.
	Move(item bump.Item, goal bump.Vec2, filter bump.Filter, tags ...bump.Tag) (bump.Vec2, []*bump.Collision)

	// Remove removes an item from the collision space.
	Remove(item bump.Item)

	// Has checks if an item exists in the space with the given tags.
	Has(item bump.Item, tags ...bump.Tag) bool

	// Rect returns the current rect for an item in the collision space.
	Rect(item bump.Item) bump.Rect

	// Query performs a spatial query for items overlapping the given rect.
	Query(rect bump.Rect, filter bump.SelectFilter, tags ...bump.Tag) []*bump.Collision
}
