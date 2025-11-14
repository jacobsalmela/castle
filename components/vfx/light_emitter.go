package vfx

import "game/entities"

// LightEmitter marks an entity as a light source for the dynamic lighting system.
// The lighting system queries all entities with this component each frame to build
// the light list for the shader.
type LightEmitter struct {
	// EntityID is the ECS entity reference
	EntityID entities.EntityId

	// Radius is the light emission radius in pixels
	Radius float64

	// Active determines if this light is currently emitting
	// When false, the lighting system skips this emitter
	Active bool

	// Intensity is the brightness multiplier (0.0 = off, 1.0 = full)
	// The shader uses this to modulate the light strength
	Intensity float64

	// PulseSpeed controls how fast the light pulsates (if > 0)
	// Set to 0 for static lights, > 0 for flickering torches
	PulseSpeed float64

	// PulsePhase is the time offset for pulsation (set at spawn for variety)
	PulsePhase float64
}
