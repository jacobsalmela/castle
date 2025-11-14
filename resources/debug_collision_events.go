package resources

import (
	"time"
)

// CollisionEvent represents a collision that occurred for debug visualization.
// Used by the collision debug system to show impact points, normals, and connections.
type CollisionEvent struct {
	ItemX, ItemY     float64 // Position of the item that moved
	OtherX, OtherY   float64 // Position of the other item
	NormalX, NormalY float64 // Collision normal vector
	Overlap          float64 // How much the items overlap
	Timestamp        time.Time
}

// CollisionEventQueue stores recent collision events for visualization.
// This replaces the global slice in pkg/debug with Pure ECS architecture.
//
// Usage:
//   - Get resource: events := ecs.Resource[resources.CollisionEventQueue](world)
//   - Record event: events.Record(itemX, itemY, otherX, otherY, normalX, normalY, overlap)
//   - Get recent: for _, event := range events.Recent() { ... }
//
// Events are automatically expired after the configured lifetime (e.g., 2 seconds).
type CollisionEventQueue struct {
	events   []CollisionEvent
	lifetime time.Duration
}

// NewCollisionEventQueue creates a new collision event queue with the given lifetime.
// Lifetime determines how long events are kept for visualization (e.g., 2 seconds for trail effect).
//
// Example:
//
//	queue := NewCollisionEventQueue(2000 * time.Millisecond)
func NewCollisionEventQueue(lifetime time.Duration) *CollisionEventQueue {
	return &CollisionEventQueue{
		events:   make([]CollisionEvent, 0, 100),
		lifetime: lifetime,
	}
}

// Record adds a new collision event to the queue.
// Called by physics systems when collisions occur.
func (q *CollisionEventQueue) Record(itemX, itemY, otherX, otherY, normalX, normalY, overlap float64) {
	q.events = append(q.events, CollisionEvent{
		ItemX:     itemX,
		ItemY:     itemY,
		OtherX:    otherX,
		OtherY:    otherY,
		NormalX:   normalX,
		NormalY:   normalY,
		Overlap:   overlap,
		Timestamp: time.Now(),
	})
}

// Recent returns all events still within the lifetime window.
// Also removes expired events from the queue.
func (q *CollisionEventQueue) Recent() []CollisionEvent {
	now := time.Now()
	valid := q.events[:0]

	for _, event := range q.events {
		if now.Sub(event.Timestamp) < q.lifetime {
			valid = append(valid, event)
		}
	}

	q.events = valid
	return q.events
}

// Clear removes all events from the queue.
func (q *CollisionEventQueue) Clear() {
	q.events = q.events[:0]
}

// Len returns the number of events currently in the queue.
func (q *CollisionEventQueue) Len() int {
	return len(q.events)
}
