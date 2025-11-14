package resources

import (
	"game/entities"
	"game/pkg/bump"
)

// ImpactType categorizes what a projectile collided with.
type ImpactType int

const (
	ImpactWall ImpactType = iota
	ImpactEntity
	ImpactCeiling
	ImpactFloor
)

// TriggerType categorizes the kind of trigger zone activated.
type TriggerType int

const (
	TriggerDoor TriggerType = iota
	TriggerChest
	TriggerDialogue
	TriggerZone
)

// ImpactEvent records when a projectile hits something (wall, entity, etc).
type ImpactEvent struct {
	Projectile entities.EntityId // The projectile that impacted
	Target     entities.EntityId // The entity hit (0 if wall/map)
	Position   bump.Vec2         // Impact location
	ImpactType ImpactType        // What was hit
	Normal     bump.Vec2         // Surface normal (for bounces)
}

// TriggerEvent records when an entity enters a trigger zone.
type TriggerEvent struct {
	Trigger     entities.EntityId // The trigger entity (door, chest, NPC)
	Activator   entities.EntityId // Who activated it (usually player)
	TriggerType TriggerType       // Type of trigger
	Position    bump.Vec2         // Trigger location
}

// HitEvent represents the outcome of resolving a single hitbox overlap.
type HitEvent struct {
	Attacker   entities.EntityId
	Target     entities.EntityId
	Damage     float64
	Contact    int
	AttackRect bump.Rect
	TargetRect bump.Rect
	// Handled marks when an entity-specific system consumes the hit,
	// preventing generic combat systems from applying default reactions.
	// Example: A door opens on hit and sets Handled=true to avoid taking damage.
	Handled bool
}

// EventQueue stores all game events for the current frame.
// This unified queue consolidates collision, combat, and trigger events
// into a single resource for simpler event handling.
//
// Replaces the old pattern of separate CombatEvents and CollisionEvents resources.
type EventQueue struct {
	Hits     []HitEvent     // Combat hit events
	Impacts  []ImpactEvent  // Projectile impact events
	Triggers []TriggerEvent // Trigger zone activation events
}

// NewEventQueue constructs an empty event queue ready for reuse.
func NewEventQueue() *EventQueue {
	return &EventQueue{
		Hits:     make([]HitEvent, 0, 16),
		Impacts:  make([]ImpactEvent, 0, 8),
		Triggers: make([]TriggerEvent, 0, 4),
	}
}

// === HIT EVENTS ===

// PushHit enqueues a combat hit event.
func (e *EventQueue) PushHit(event HitEvent) {
	if e == nil {
		return
	}
	e.Hits = append(e.Hits, event)
}

// DrainHits returns pending hit events and clears the queue.
func (e *EventQueue) DrainHits() []HitEvent {
	if e == nil || len(e.Hits) == 0 {
		return nil
	}
	hits := e.Hits
	e.Hits = e.Hits[:0]
	return hits
}

// HasHits reports whether any hit events are pending.
func (e *EventQueue) HasHits() bool {
	return e != nil && len(e.Hits) > 0
}

// === IMPACT EVENTS ===

// PushImpact enqueues a projectile impact event.
func (e *EventQueue) PushImpact(event ImpactEvent) {
	if e == nil {
		return
	}
	e.Impacts = append(e.Impacts, event)
}

// DrainImpacts returns pending impact events and clears the queue.
func (e *EventQueue) DrainImpacts() []ImpactEvent {
	if e == nil || len(e.Impacts) == 0 {
		return nil
	}
	impacts := e.Impacts
	e.Impacts = e.Impacts[:0]
	return impacts
}

// HasImpacts reports whether any impact events are pending.
func (e *EventQueue) HasImpacts() bool {
	return e != nil && len(e.Impacts) > 0
}

// === TRIGGER EVENTS ===

// PushTrigger enqueues a trigger zone activation event.
func (e *EventQueue) PushTrigger(event TriggerEvent) {
	if e == nil {
		return
	}
	e.Triggers = append(e.Triggers, event)
}

// DrainTriggers returns pending trigger events and clears the queue.
func (e *EventQueue) DrainTriggers() []TriggerEvent {
	if e == nil || len(e.Triggers) == 0 {
		return nil
	}
	triggers := e.Triggers
	e.Triggers = e.Triggers[:0]
	return triggers
}

// HasTriggers reports whether any trigger events are pending.
func (e *EventQueue) HasTriggers() bool {
	return e != nil && len(e.Triggers) > 0
}

// === UTILITY METHODS ===

// Clear resets all event queues for the next frame.
// This should be called at the end of each update cycle after all
// reaction systems have processed their events.
func (e *EventQueue) Clear() {
	if e == nil {
		return
	}
	e.Hits = e.Hits[:0]
	e.Impacts = e.Impacts[:0]
	e.Triggers = e.Triggers[:0]
}

// EventCount returns the total number of pending events across all types.
func (e *EventQueue) EventCount() int {
	if e == nil {
		return 0
	}
	return len(e.Hits) + len(e.Impacts) + len(e.Triggers)
}
