//go:build !release

package ai

import (
	"fmt"
	"game/ecs"
	"game/entities"
)

// DebugNodeInfo stores debug information about a behavior tree node's execution.
type DebugNodeInfo struct {
	Name       string // Human-readable name for the node
	LastStatus Status // Last status returned by this node
	TickCount  int    // Number of times this node has been ticked
	Depth      int    // Depth in the tree (0 = root)
}

// DebugNode wraps a Node with debug tracking capabilities.
// This is only compiled in non-release builds to avoid performance overhead.
type DebugNode struct {
	Node Node           // The actual node being wrapped
	Info *DebugNodeInfo // Debug information
}

// Tick executes the wrapped node and tracks its status.
func (d *DebugNode) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	d.Info.TickCount++
	status := d.Node.Tick(world, eid, dt)
	d.Info.LastStatus = status
	return status
}

// Reset resets the debug info and the wrapped node if it's Resettable.
func (d *DebugNode) Reset() {
	d.Info.TickCount = 0
	d.Info.LastStatus = Success // Default to success
	if resettable, ok := d.Node.(Resettable); ok {
		resettable.Reset()
	}
}

// WrapForDebug wraps a node with debug tracking.
// This recursively wraps all child nodes in composites and decorators.
// If a node is already wrapped, it returns the existing wrapper to preserve custom names.
func WrapForDebug(node Node, name string, depth int) *DebugNode {
	// Check if already wrapped - preserve custom names
	if existingDebug, ok := node.(*DebugNode); ok {
		return existingDebug
	}

	debugNode := &DebugNode{
		Node: node,
		Info: &DebugNodeInfo{
			Name:       name,
			LastStatus: Success,
			TickCount:  0,
			Depth:      depth,
		},
	}

	// Recursively wrap child nodes using polymorphism
	if composite, ok := node.(Composite); ok {
		// Multi-child node (Sequence, Selector, Parallel)
		children := composite.GetChildren()
		for i, child := range children {
			children[i] = WrapForDebug(child, fmt.Sprintf("%s.Child[%d]", name, i), depth+1)
		}
		composite.SetChildren(children)
	} else if decorator, ok := node.(Decorator); ok {
		// Single-child node (Repeat, Inverter, Timeout, etc.)
		if child := decorator.GetChild(); child != nil {
			decorator.SetChild(WrapForDebug(child, fmt.Sprintf("%s.Child", name), depth+1))
		}
	}

	return debugNode
}

// CollectDebugInfo recursively collects debug information from all nodes in the tree.
func CollectDebugInfo(node Node) []*DebugNodeInfo {
	var infos []*DebugNodeInfo

	// If this is a debug node, add its info
	if debugNode, ok := node.(*DebugNode); ok {
		infos = append(infos, debugNode.Info)
		node = debugNode.Node // Continue with the wrapped node
	}

	// Recursively collect from children using polymorphism
	if composite, ok := node.(Composite); ok {
		// Multi-child node (Sequence, Selector, Parallel)
		for _, child := range composite.GetChildren() {
			infos = append(infos, CollectDebugInfo(child)...)
		}
	} else if decorator, ok := node.(Decorator); ok {
		// Single-child node (Repeat, Inverter, Timeout, etc.)
		if child := decorator.GetChild(); child != nil {
			infos = append(infos, CollectDebugInfo(child)...)
		}
	}

	return infos
}
