package ai

import (
	"game/ecs"
	"game/entities"
)

// Status represents the result of a behavior tree node tick.
type Status int

const (
	// Success indicates the node completed successfully.
	Success Status = iota
	// Failure indicates the node failed to complete its task.
	Failure
	// Running indicates the node is still executing and needs more ticks.
	Running
)

// String returns a human-readable representation of the Status.
func (s Status) String() string {
	switch s {
	case Success:
		return "Success"
	case Failure:
		return "Failure"
	case Running:
		return "Running"
	default:
		return "Unknown"
	}
}

// Node is the core interface for all behavior tree nodes.
// Each node implements Tick(), which is called every frame to evaluate the node's behavior.
// Nodes must be stateless regarding ECS data - all state comes from components.
// Nodes may have internal state for tracking progress (e.g., current child index, elapsed time).
type Node interface {
	// Tick evaluates the node's behavior and returns its current status.
	// world: The ECS world for querying components
	// eid: The entity this behavior tree belongs to
	// dt: Delta time since last tick (in seconds)
	Tick(world *ecs.World, eid entities.EntityId, dt float64) Status
}

// Reset is an optional interface that nodes can implement to reset their internal state.
// This is useful for nodes that track progress (e.g., Sequence tracking current child).
type Resettable interface {
	Reset()
}

// Composite is an interface for nodes with multiple children.
// This allows polymorphic access to children without type switching.
type Composite interface {
	Node
	GetChildren() []Node
	SetChildren([]Node)
}

// Decorator is an interface for nodes with a single child.
// This allows polymorphic access to the child without type switching.
type Decorator interface {
	Node
	GetChild() Node
	SetChild(Node)
}
