//go:build !release

package primitives

import (
	"game/ecs"
	"game/resources"
)

// IsDebugEnabled checks if a debug category is currently enabled.
// Returns false if DebugState resource is not available.
func IsDebugEnabled(world *ecs.World, category string) bool {
	debugState := ecs.Resource[resources.DebugState](world)
	return debugState != nil && debugState.IsEnabled(category)
}

// IsAnyDebugEnabled checks if any of the given debug categories are enabled.
func IsAnyDebugEnabled(world *ecs.World, categories ...string) bool {
	debugState := ecs.Resource[resources.DebugState](world)
	if debugState == nil {
		return false
	}

	for _, category := range categories {
		if debugState.IsEnabled(category) {
			return true
		}
	}
	return false
}

// GetDebugState returns the DebugState resource or nil if not available.
func GetDebugState(world *ecs.World) *resources.DebugState {
	return ecs.Resource[resources.DebugState](world)
}
