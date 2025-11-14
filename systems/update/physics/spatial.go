package physics

import (
	"game/pkg/bump"
)

// QueryItems finds items intersecting rect, filtered by type and tags.
//
// Example:
//
//	space := physics.GetCollisionSpace(world)
//	bodies := physics.QueryItems(space, entities.EntityId(0), rect, "body")
func QueryItems[T comparable](space *bump.Space, item T, rect bump.Rect, tags ...bump.Tag) []T {
	if space == nil {
		return nil
	}

	filter := NotSelfFilter(item)
	cols := space.Query(rect, filter, tags...)

	var items []T
	for _, c := range cols {
		if e, ok := c.Other.(T); ok {
			items = append(items, e)
		}
	}
	return items
}
