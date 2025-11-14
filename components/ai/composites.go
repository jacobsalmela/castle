package ai

import (
	"game/ecs"
	"game/entities"
)

// Sequence executes its children in order until one fails or all succeed.
// Returns Success if all children succeed.
// Returns Failure if any child fails.
// Returns Running if the current child is still running.
//
// Use case: "Do A, then B, then C" - all must succeed in order.
// Example: Sequence{CheckInRange, FaceTarget, Attack}
type Sequence struct {
	Children []Node
	current  int
}

func (s *Sequence) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	for s.current < len(s.Children) {
		status := s.Children[s.current].Tick(world, eid, dt)

		switch status {
		case Failure:
			s.current = 0 // Reset for next run
			return Failure
		case Running:
			return Running // Keep current index for next tick
		case Success:
			s.current++ // Move to next child
		}
	}

	// All children succeeded
	s.current = 0 // Reset for next run
	return Success
}

func (s *Sequence) Reset() {
	s.current = 0
	for _, child := range s.Children {
		if r, ok := child.(Resettable); ok {
			r.Reset()
		}
	}
}

// GetChildren returns the sequence's children (implements Composite interface)
func (s *Sequence) GetChildren() []Node {
	return s.Children
}

// SetChildren sets the sequence's children (implements Composite interface)
func (s *Sequence) SetChildren(children []Node) {
	s.Children = children
}

// Selector executes its children in order until one succeeds or all fail.
// Returns Success if any child succeeds.
// Returns Failure if all children fail.
// Returns Running if the current child is still running.
//
// Use case: "Try A, if that fails try B, if that fails try C" - priority-based selection.
// Example: Selector{AttackIfClose, ChaseIfDetected, Patrol}
type Selector struct {
	Children []Node
	current  int
}

func (s *Selector) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	for s.current < len(s.Children) {
		status := s.Children[s.current].Tick(world, eid, dt)

		switch status {
		case Success:
			s.current = 0 // Reset for next run
			return Success
		case Running:
			return Running // Keep current index for next tick
		case Failure:
			s.current++ // Try next child
		}
	}

	// All children failed
	s.current = 0 // Reset for next run
	return Failure
}

func (s *Selector) Reset() {
	s.current = 0
	for _, child := range s.Children {
		if r, ok := child.(Resettable); ok {
			r.Reset()
		}
	}
}

// GetChildren returns the selector's children (implements Composite interface)
func (s *Selector) GetChildren() []Node {
	return s.Children
}

// SetChildren sets the selector's children (implements Composite interface)
func (s *Selector) SetChildren(children []Node) {
	s.Children = children
}

// Parallel executes all children simultaneously and returns based on policy.
// Note: This is a simplified parallel that ticks all children each frame.
// It doesn't actually run in parallel threads.
type Parallel struct {
	Children []Node
	// SuccessPolicy determines how many children must succeed for the parallel to succeed.
	// If SuccessPolicy == 0, uses RequireAll (all must succeed).
	// If SuccessPolicy == 1, uses RequireOne (one must succeed).
	// Otherwise, uses the specific count (e.g., 2 means at least 2 must succeed).
	SuccessPolicy int
	// FailurePolicy determines how many children must fail for the parallel to fail.
	// If FailurePolicy == 0, uses RequireAll (all must fail).
	// If FailurePolicy == 1, uses RequireOne (one must fail).
	// Otherwise, uses the specific count.
	FailurePolicy int
}

const (
	// RequireOne means at least one child must succeed/fail.
	RequireOne = 1
	// RequireAll means all children must succeed/fail.
	RequireAll = -1
)

// calculateThreshold determines the required count based on policy and total children.
// If policy == RequireOne (1), returns 1.
// If policy > 0, returns the policy value.
// Otherwise (RequireAll or 0), returns totalChildren.
func (p *Parallel) calculateThreshold(policy int, totalChildren int) int {
	if policy == RequireOne {
		return 1
	}
	if policy > 0 {
		return policy
	}
	return totalChildren // RequireAll or default
}

func (p *Parallel) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	successCount := 0
	failureCount := 0

	for _, child := range p.Children {
		status := child.Tick(world, eid, dt)
		switch status {
		case Success:
			successCount++
		case Failure:
			failureCount++
		}
	}

	successThreshold := p.calculateThreshold(p.SuccessPolicy, len(p.Children))
	failureThreshold := p.calculateThreshold(p.FailurePolicy, len(p.Children))

	// Check failure first (early exit)
	if failureCount >= failureThreshold {
		return Failure
	}

	// Check success
	if successCount >= successThreshold {
		return Success
	}

	// Still running
	return Running
}

func (p *Parallel) Reset() {
	for _, child := range p.Children {
		if r, ok := child.(Resettable); ok {
			r.Reset()
		}
	}
}

// GetChildren returns the parallel's children (implements Composite interface)
func (p *Parallel) GetChildren() []Node {
	return p.Children
}

// SetChildren sets the parallel's children (implements Composite interface)
func (p *Parallel) SetChildren(children []Node) {
	p.Children = children
}
