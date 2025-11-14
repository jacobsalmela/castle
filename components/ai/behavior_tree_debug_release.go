//go:build release

package ai

import (
	"game/ecs"
	"game/entities"
)

// WrapForDebug is a no-op in release builds.
// Returns the node unchanged.
func WrapForDebug(node Node, name string, depth int) Node {
	return node
}

// CollectDebugInfo is a no-op in release builds.
// Returns an empty slice.
func CollectDebugInfo(node Node) []*DebugNodeInfo {
	return nil
}

// DebugNodeInfo is a stub type for release builds.
type DebugNodeInfo struct {
	Name       string
	LastStatus Status
	TickCount  int
	Depth      int
}
