package resources

import (
	"time"
)

// HitboxEvent represents an active attack hitbox for debug visualization.
// Used by the hitbox debug system to show attack rectangles with fade effect.
type HitboxEvent struct {
	X, Y, W, H float64 // Hitbox rectangle in world coordinates
	Attacker   string  // Name/ID of attacker (for display)
	Timestamp  time.Time
}

// HitboxEventQueue stores recent hitbox events for visualization.
// This replaces the global slice in pkg/debug with Pure ECS architecture.
//
// Usage:
//   - Get resource: events := ecs.Resource[resources.HitboxEventQueue](world)
//   - Record event: events.Record(x, y, w, h, "PlayerAttack")
//   - Get recent: for _, event := range events.Recent() { ... }
//
// Events are automatically expired after the configured lifetime (e.g., 1 second).
type HitboxEventQueue struct {
	events   []HitboxEvent
	lifetime time.Duration
}

// NewHitboxEventQueue creates a new hitbox event queue with the given lifetime.
// Lifetime determines how long events are kept for visualization (e.g., 1 second for attack trail).
//
// Example:
//
//	queue := NewHitboxEventQueue(1000 * time.Millisecond)
func NewHitboxEventQueue(lifetime time.Duration) *HitboxEventQueue {
	return &HitboxEventQueue{
		events:   make([]HitboxEvent, 0, 50),
		lifetime: lifetime,
	}
}

// Record adds a new hitbox event to the queue.
// Called by combat systems when attack hitboxes become active.
func (q *HitboxEventQueue) Record(x, y, w, h float64, attacker string) {
	q.events = append(q.events, HitboxEvent{
		X:         x,
		Y:         y,
		W:         w,
		H:         h,
		Attacker:  attacker,
		Timestamp: time.Now(),
	})
}

// Recent returns all events still within the lifetime window.
// Also removes expired events from the queue.
func (q *HitboxEventQueue) Recent() []HitboxEvent {
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
func (q *HitboxEventQueue) Clear() {
	q.events = q.events[:0]
}

// Len returns the number of events currently in the queue.
func (q *HitboxEventQueue) Len() int {
	return len(q.events)
}
