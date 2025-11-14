package ai

import (
	"game/ecs"
	"game/entities"
)

// Condition is a leaf node that evaluates a condition and returns Success or Failure.
// Conditions should never return Running - they're instantaneous checks.
//
// Use case: Checking game state without modifying it.
// Example: "Is player in range?", "Do I have ammo?", "Am I grounded?"
type Condition struct {
	// Check is called to evaluate the condition.
	// Should return true for Success, false for Failure.
	Check func(world *ecs.World, eid entities.EntityId) bool
}

func (c *Condition) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	if c.Check == nil {
		return Failure
	}

	if c.Check(world, eid) {
		return Success
	}
	return Failure
}

// Action is a leaf node that performs an action and may take multiple frames to complete.
// Actions can return Success, Failure, or Running.
//
// Use case: Doing something in the game world.
// Example: "Move toward target", "Play attack animation", "Fire weapon"
type Action struct {
	// OnStart is called once when the action begins (optional).
	OnStart func(world *ecs.World, eid entities.EntityId)

	// OnTick is called every frame while the action is running.
	// Should return the current status of the action.
	OnTick func(world *ecs.World, eid entities.EntityId, dt float64) Status

	// OnEnd is called once when the action completes (optional).
	// Called regardless of whether action succeeded or failed.
	OnEnd func(world *ecs.World, eid entities.EntityId)

	started bool
}

func (a *Action) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	// Call OnStart on first tick
	if !a.started {
		a.started = true
		if a.OnStart != nil {
			a.OnStart(world, eid)
		}
	}

	// Execute the action
	var status Status
	if a.OnTick != nil {
		status = a.OnTick(world, eid, dt)
	} else {
		status = Success // No tick function means instant success
	}

	// Call OnEnd when action completes
	if status != Running {
		a.started = false
		if a.OnEnd != nil {
			a.OnEnd(world, eid)
		}
	}

	return status
}

func (a *Action) Reset() {
	a.started = false
}

// AlwaysSuccess is a leaf node that always returns Success.
// Useful as a fallback or placeholder.
type AlwaysSuccess struct{}

func (a *AlwaysSuccess) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	return Success
}

// AlwaysFailure is a leaf node that always returns Failure.
// Useful for testing or placeholders.
type AlwaysFailure struct{}

func (a *AlwaysFailure) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	return Failure
}

// AlwaysRunning is a leaf node that always returns Running.
// Useful for blocking a branch or testing.
type AlwaysRunning struct{}

func (a *AlwaysRunning) Tick(world *ecs.World, eid entities.EntityId, dt float64) Status {
	return Running
}
