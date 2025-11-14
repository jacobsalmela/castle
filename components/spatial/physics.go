package spatial

import "game/pkg/config"

// Vec2 represents a 2D vector for physics calculations.
type Vec2 struct {
	X float64
	Y float64
}

// Physics stores all movement and force data for an entity.
type Physics struct {
	// === MOVEMENT ===
	Velocity     Vec2 // Current velocity (pixels/sec)
	PrevVelocity Vec2 // Last frame velocity (for friction calculation)
	Acceleration Vec2 // Forces applied this frame (cleared each frame by systems)
	MaxVelocity  Vec2 // Speed caps (X: horizontal, Y: vertical)

	// === GROUNDING STATE ===
	Grounded bool    // On solid ground
	CanJump  bool    // Jump available (includes coyote time)
	AirTime  float64 // Seconds since leaving ground (for coyote time)

	// === PHYSICS PROPERTIES ===
	Weight     float64 // Gravity multiplier (1.0 = normal, 0 = no gravity)
	Friction   float64 // Drag coefficient (0 = ice, 1 = instant stop)
	Bounciness float64 // Restitution coefficient (0 = no bounce, 1 = perfect bounce)

	// === PLATFORM MECHANICS ===
	OnPlatform       bool    // Currently on pass-through platform
	DroppingThrough  bool    // Actively falling through platform
	PlatformDropTime float64 // Drop-through duration remaining (seconds)

	// === FLAGS ===
	FrictionEnabled bool // Apply friction this frame
	GravityEnabled  bool // Apply gravity this frame
}

// NewPhysics creates a Physics component with default values from config.
func NewPhysics() *Physics {
	cfg := config.Cfg
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	return &Physics{
		MaxVelocity:     Vec2{X: cfg.Body.MaxX, Y: cfg.Body.MaxY},
		Weight:          1.0,
		Friction:        cfg.Body.GroundFriction,
		Bounciness:      0.0,
		FrictionEnabled: true,
		GravityEnabled:  true,
	}
}

// NewPhysicsStatic creates a Physics component for static/immovable entities.
// These entities don't move but may need grounding state for logic.
func NewPhysicsStatic() *Physics {
	return &Physics{
		Weight:          0,
		GravityEnabled:  false,
		FrictionEnabled: false,
	}
}

// ApplyForce adds a force to acceleration (cleared each frame by physics systems).
// Use for continuous forces like wind, conveyor belts, or acceleration.
func (p *Physics) ApplyForce(fx, fy float64) {
	if p == nil {
		return
	}
	p.Acceleration.X += fx
	p.Acceleration.Y += fy
}

// ApplyImpulse adds directly to velocity (instant).
// Use for instant forces like jumping, knockback, explosions.
func (p *Physics) ApplyImpulse(vx, vy float64) {
	if p == nil {
		return
	}
	p.Velocity.X += vx
	p.Velocity.Y += vy
}

// ClampVelocity enforces max velocity limits.
// Called by physics systems after applying forces.
func (p *Physics) ClampVelocity() {
	if p == nil {
		return
	}

	// Clamp X
	if p.MaxVelocity.X > 0 {
		if p.Velocity.X > p.MaxVelocity.X {
			p.Velocity.X = p.MaxVelocity.X
		} else if p.Velocity.X < -p.MaxVelocity.X {
			p.Velocity.X = -p.MaxVelocity.X
		}
	}

	// Clamp Y
	if p.MaxVelocity.Y > 0 {
		if p.Velocity.Y > p.MaxVelocity.Y {
			p.Velocity.Y = p.MaxVelocity.Y
		} else if p.Velocity.Y < -p.MaxVelocity.Y {
			p.Velocity.Y = -p.MaxVelocity.Y
		}
	}
}

// ResetAcceleration clears forces (called each frame by physics systems).
func (p *Physics) ResetAcceleration() {
	if p == nil {
		return
	}
	p.Acceleration.X = 0
	p.Acceleration.Y = 0
}

// SetVelocity directly sets velocity (use sparingly - prefer ApplyForce/ApplyImpulse).
func (p *Physics) SetVelocity(vx, vy float64) {
	if p == nil {
		return
	}
	p.Velocity.X = vx
	p.Velocity.Y = vy
}

// StopHorizontal stops horizontal movement immediately.
func (p *Physics) StopHorizontal() {
	if p == nil {
		return
	}
	p.Velocity.X = 0
	p.Acceleration.X = 0
}

// StopVertical stops vertical movement immediately.
func (p *Physics) StopVertical() {
	if p == nil {
		return
	}
	p.Velocity.Y = 0
	p.Acceleration.Y = 0
}

// Stop stops all movement immediately.
func (p *Physics) Stop() {
	if p == nil {
		return
	}
	p.Velocity.X = 0
	p.Velocity.Y = 0
	p.Acceleration.X = 0
	p.Acceleration.Y = 0
}
