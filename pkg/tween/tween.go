package tween

// Package tween provides simple tweening/easing functions for animations.
// This replaces the third-party gween library with lightweight, self-contained implementations.

import "math"

// Lerp performs linear interpolation between two float64 values.
// Returns: a + (b-a)*t
//
// Parameters:
//   - a: Starting value
//   - b: Ending value
//   - t: Interpolation factor (0.0 to 1.0)
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// LerpUint8 performs linear interpolation between two uint8 values.
// Useful for color channel interpolation.
//
// Parameters:
//   - a: Starting value
//   - b: Ending value
//   - t: Interpolation factor (0.0 to 1.0)
func LerpUint8(a, b uint8, t float64) uint8 {
	return uint8(float64(a) + (float64(b)-float64(a))*t)
}

// EaseLinear applies linear easing (no easing, just returns t).
// Useful as a baseline or when you want constant-speed animation.
func EaseLinear(t float64) float64 {
	return t
}

// EaseOutCubic applies cubic ease-out easing to a progress value.
// Fast at the start, slowing down at the end.
// Formula: 1 - (1-t)³
//
// Parameters:
//   - t: Progress (0.0 to 1.0)
//
// Returns: Eased progress (0.0 to 1.0)
func EaseOutCubic(t float64) float64 {
	t = 1 - t
	return 1 - t*t*t
}

// EaseInCubic applies cubic ease-in easing to a progress value.
// Slow at the start, speeding up at the end.
// Formula: t³
func EaseInCubic(t float64) float64 {
	return t * t * t
}

// EaseInOutCubic applies cubic ease-in-out easing to a progress value.
// Slow at start and end, fast in the middle.
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	t = 2*t - 2
	return 1 + t*t*t/2
}

// EaseOutQuad applies quadratic ease-out easing.
// Formula: 1 - (1-t)²
func EaseOutQuad(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

// EaseInQuad applies quadratic ease-in easing.
// Formula: t²
func EaseInQuad(t float64) float64 {
	return t * t
}

// EaseOutElastic applies elastic ease-out easing (bounce effect).
func EaseOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	p := 0.3
	s := p / 4
	return math.Pow(2, -10*t)*math.Sin((t-s)*(2*math.Pi)/p) + 1
}

// Tween represents a simple time-based animation between two values.
type Tween struct {
	start    float64
	end      float64
	duration float64
	elapsed  float64
	easing   func(float64) float64
}

// New creates a new Tween that animates from start to end over duration seconds.
//
// Parameters:
//   - start: Starting value
//   - end: Ending value
//   - duration: Duration in seconds
//   - easing: Easing function (use EaseLinear, EaseOutCubic, etc.)
func New(start, end, duration float64, easing func(float64) float64) *Tween {
	if easing == nil {
		easing = EaseLinear
	}
	return &Tween{
		start:    start,
		end:      end,
		duration: duration,
		elapsed:  0,
		easing:   easing,
	}
}

// Update advances the tween by dt seconds and returns the current value and whether it's done.
//
// Returns:
//   - value: Current interpolated value
//   - done: True if the tween has completed
func (t *Tween) Update(dt float64) (value float64, done bool) {
	if t.elapsed >= t.duration {
		return t.end, true
	}

	t.elapsed += dt
	if t.elapsed >= t.duration {
		t.elapsed = t.duration
		return t.end, true
	}

	// Calculate progress (0 to 1)
	progress := t.elapsed / t.duration

	// Apply easing
	easedProgress := t.easing(progress)

	// Interpolate
	value = Lerp(t.start, t.end, easedProgress)
	return value, false
}

// Reset resets the tween to the beginning.
func (t *Tween) Reset() {
	t.elapsed = 0
}

// Progress returns the current progress (0.0 to 1.0).
func (t *Tween) Progress() float64 {
	if t.duration == 0 {
		return 1
	}
	progress := t.elapsed / t.duration
	if progress > 1 {
		return 1
	}
	return progress
}

// Value returns the current interpolated value without advancing time.
func (t *Tween) Value() float64 {
	if t.elapsed >= t.duration {
		return t.end
	}
	progress := t.elapsed / t.duration
	easedProgress := t.easing(progress)
	return Lerp(t.start, t.end, easedProgress)
}

// IsDone returns true if the tween has completed.
func (t *Tween) IsDone() bool {
	return t.elapsed >= t.duration
}
