package physics

import (
	"game/ecs"
	"game/pkg/bump"
)

// GetCollisionSpace returns the collision space stored on the ECS world.
// Returns the bump.Space resource from the ECS world.
func GetCollisionSpace(world *ecs.World) *bump.Space {
	if world == nil {
		return nil
	}
	return ecs.Resource[bump.Space](world)
}

// NotSelfFilter returns a bump.SelectFilter that accepts items matching the type
// while excluding the item itself. Prevents entities from colliding with themselves.
func NotSelfFilter[T comparable](item T) bump.SelectFilter {
	return func(other bump.Item) bool {
		if candidate, ok := other.(T); ok {
			return candidate != item
		}
		return false
	}
}
