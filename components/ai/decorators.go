package ai

import (
	"game/ecs"
	"game/entities"
)

// Inverter inverts the result of its child node.
// Success becomes Failure, Failure becomes Success.
// Running remains Running.
//
// Use case: "Do the opposite" - useful for negating conditions.
// Example: Inverter{IsInRange} becomes "IsNotInRange"
type Inverter struct {
	Child Node
}

func (i *Inverter) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	status := i.Child.Tick(world, eid, dt)
	switch status {
	case Success:
		return Failure
	case Failure:
		return Success
	default:
		return Running
	}
}

func (i *Inverter) Reset() {
	if r, ok := i.Child.(Resettable); ok {
		r.Reset()
	}
}

// GetChild returns the inverter's child (implements Decorator interface)
func (i *Inverter) GetChild() Node {
	return i.Child
}

// SetChild sets the inverter's child (implements Decorator interface)
func (i *Inverter) SetChild(child Node) {
	i.Child = child
}

// Repeat repeats its child node a specified number of times or until it fails.
// If Count <= 0, repeats indefinitely (always returns Running).
// Returns Success after completing all repetitions.
// Returns Failure if child fails before completing all repetitions.
//
// Use case: "Do X multiple times" or "Loop forever".
// Example: Repeat{Count: 3, Child: Attack} - attack 3 times
type Repeat struct {
	Child Node
	Count int // Number of times to repeat (0 = infinite)

	currentCount int
}

func (r *Repeat) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	if r.Count <= 0 {
		return r.tickInfinite(world, eid, dt)
	}
	return r.tickFinite(world, eid, dt)
}

// tickInfinite handles infinite loop strategy (Count <= 0)
func (r *Repeat) tickInfinite(world *ecs.World, eid entities.EntityId, dt float64) Status {
	status := r.Child.Tick(world, eid, dt)
	if status == Failure {
		return Failure
	}
	if status == Success {
		r.resetChild()
	}
	return Running
}

// tickFinite handles finite loop strategy (Count > 0)
func (r *Repeat) tickFinite(world *ecs.World, eid entities.EntityId, dt float64) Status {
	for r.currentCount < r.Count {
		status := r.Child.Tick(world, eid, dt)

		if status == Failure {
			r.currentCount = 0
			return Failure
		}

		if status == Running {
			return Running
		}

		// Success - increment and reset child for next iteration
		r.currentCount++
		r.resetChild()
	}

	// Completed all repetitions
	r.currentCount = 0
	return Success
}

// resetChild resets the child node if it's resettable
func (r *Repeat) resetChild() {
	if res, ok := r.Child.(Resettable); ok {
		res.Reset()
	}
}

func (r *Repeat) Reset() {
	r.currentCount = 0
	if res, ok := r.Child.(Resettable); ok {
		res.Reset()
	}
}

// GetChild returns the repeat's child (implements Decorator interface)
func (r *Repeat) GetChild() Node {
	return r.Child
}

// SetChild sets the repeat's child (implements Decorator interface)
func (r *Repeat) SetChild(child Node) {
	r.Child = child
}

// UntilFail repeats its child until it returns Failure.
// Always returns Running (unless child fails).
// Returns Success when child fails.
//
// Use case: "Keep doing X until it fails".
// Example: UntilFail{Patrol} - patrol forever until something interrupts
type UntilFail struct {
	Child Node
}

func (u *UntilFail) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	status := u.Child.Tick(world, eid, dt)

	if status == Failure {
		return Success
	}

	if status == Success {
		// Reset child for next iteration
		if r, ok := u.Child.(Resettable); ok {
			r.Reset()
		}
	}

	return Running
}

func (u *UntilFail) Reset() {
	if r, ok := u.Child.(Resettable); ok {
		r.Reset()
	}
}

// GetChild returns the until-fail's child (implements Decorator interface)
func (u *UntilFail) GetChild() Node {
	return u.Child
}

// SetChild sets the until-fail's child (implements Decorator interface)
func (u *UntilFail) SetChild(child Node) {
	u.Child = child
}

// UntilSuccess repeats its child until it returns Success.
// Always returns Running (unless child succeeds).
// Returns Success when child succeeds.
//
// Use case: "Keep trying X until it succeeds".
// Example: UntilSuccess{FindTarget} - keep searching until target found
type UntilSuccess struct {
	Child Node
}

func (u *UntilSuccess) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	status := u.Child.Tick(world, eid, dt)

	if status == Success {
		return Success
	}

	if status == Failure {
		// Reset child for next iteration
		if r, ok := u.Child.(Resettable); ok {
			r.Reset()
		}
	}

	return Running
}

func (u *UntilSuccess) Reset() {
	if r, ok := u.Child.(Resettable); ok {
		r.Reset()
	}
}

// GetChild returns the until-success's child (implements Decorator interface)
func (u *UntilSuccess) GetChild() Node {
	return u.Child
}

// SetChild sets the until-success's child (implements Decorator interface)
func (u *UntilSuccess) SetChild(child Node) {
	u.Child = child
}

// Timeout limits the execution time of its child node.
// Returns Failure if the child takes longer than Duration to complete.
// Returns the child's status if it completes within Duration.
//
// Use case: "Give up if this takes too long".
// Example: Timeout{Duration: 5.0, Child: ChasePlayer} - chase for max 5 seconds
type Timeout struct {
	Child    Node
	Duration float64 // Maximum time in seconds

	elapsed float64
}

func (t *Timeout) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	t.elapsed += dt

	if t.elapsed >= t.Duration {
		t.elapsed = 0
		return Failure
	}

	status := t.Child.Tick(world, eid, dt)

	if status != Running {
		t.elapsed = 0 // Reset timer when child completes
	}

	return status
}

func (t *Timeout) Reset() {
	t.elapsed = 0
	if r, ok := t.Child.(Resettable); ok {
		r.Reset()
	}
}

// GetChild returns the timeout's child (implements Decorator interface)
func (t *Timeout) GetChild() Node {
	return t.Child
}

// SetChild sets the timeout's child (implements Decorator interface)
func (t *Timeout) SetChild(child Node) {
	t.Child = child
}

// Succeeder always returns Success regardless of child's result.
// Useful for ensuring a branch always succeeds.
//
// Use case: "Try this, but don't fail if it doesn't work".
// Example: Succeeder{TryOptionalAction} - action is optional
type Succeeder struct {
	Child Node
}

func (s *Succeeder) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	status := s.Child.Tick(world, eid, dt)
	if status == Running {
		return Running
	}
	return Success
}

func (s *Succeeder) Reset() {
	if r, ok := s.Child.(Resettable); ok {
		r.Reset()
	}
}

// GetChild returns the succeeder's child (implements Decorator interface)
func (s *Succeeder) GetChild() Node {
	return s.Child
}

// SetChild sets the succeeder's child (implements Decorator interface)
func (s *Succeeder) SetChild(child Node) {
	s.Child = child
}

// Failer always returns Failure regardless of child's result.
// Useful for testing or forcing a branch to fail.
//
// Use case: Testing or placeholder nodes.
type Failer struct {
	Child Node
}

func (f *Failer) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	status := f.Child.Tick(world, eid, dt)
	if status == Running {
		return Running
	}
	return Failure
}

func (f *Failer) Reset() {
	if r, ok := f.Child.(Resettable); ok {
		r.Reset()
	}
}

// GetChild returns the failer's child (implements Decorator interface)
func (f *Failer) GetChild() Node {
	return f.Child
}

// SetChild sets the failer's child (implements Decorator interface)
func (f *Failer) SetChild(child Node) {
	f.Child = child
}
