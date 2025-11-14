package resources

// TimeControl tracks pending freeze requests and desired speed multipliers for the ECS world.
// It also maintains deterministic elapsed time for game logic.
type TimeControl struct {
	targetSpeed   float64
	speedDirty    bool
	pendingFreeze float64
	ElapsedTime   float64 // Total elapsed game time in seconds (deterministic)
}

// NewTimeControl returns a controller initialized to normal speed.
func NewTimeControl() *TimeControl {
	return &TimeControl{targetSpeed: 1, speedDirty: true}
}

// SetSpeed queues a new speed multiplier for the world.
func (t *TimeControl) SetSpeed(mult float64) {
	if t == nil {
		return
	}
	t.targetSpeed = mult
	t.speedDirty = true
}

// NextSpeed returns the most recently requested speed multiplier, reporting whether
// it represents a pending change.
func (t *TimeControl) NextSpeed() (float64, bool) {
	if t == nil {
		return 1, false
	}
	changed := t.speedDirty
	t.speedDirty = false
	return t.targetSpeed, changed
}

// RequestFreeze records a pending freeze window; the longest pending request wins.
func (t *TimeControl) RequestFreeze(duration float64) {
	if t == nil || duration <= 0 {
		return
	}
	if duration > t.pendingFreeze {
		t.pendingFreeze = duration
	}
}

// ConsumeFreeze returns the largest queued freeze duration, clearing the request.
func (t *TimeControl) ConsumeFreeze() float64 {
	if t == nil {
		return 0
	}
	duration := t.pendingFreeze
	t.pendingFreeze = 0
	return duration
}

// Reset restores the controller to its default state.
func (t *TimeControl) Reset() {
	if t == nil {
		return
	}
	t.targetSpeed = 1
	t.speedDirty = true
	t.pendingFreeze = 0
}
