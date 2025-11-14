package animation

import "game/ecs"

// Package animation provides Phase 7.1-7.2 systems: animation updates and rendering.
//
// Systems in this package:
//   - UpdateAnimations: Advance animation frames, state transitions, callbacks
//   - UpdateRenderFromAnimation: Sync Animation → Render for sprite rendering
//
// Subpackages:
//   - state: Animation state management (SetAnimationState, PlayState, SetStateEffect)
//   - callbacks: Frame and slice callback registration (RegisterFrameCallback, RegisterSliceCallback, ExtractSlice)
//
// Order: This runs at start of Phase 7 (Entity Updates), before interactive objects.
//
// ECS Pattern Compliance:
//   - Pure functions with world parameter
//   - Resource-based config access (no globals)
//   - Components are data-only
//   - All behavior in systems
//
// Cognitive Complexity:
//   - All functions ≤10 complexity (ExtractSlice refactored from 11 → 7)
//   - Helper functions extracted for clarity
//   - Clear separation of concerns

// Update orchestrates all animation systems.
// Called by systems/update/entities/package.go
func Update(world *ecs.World, dt float64) {
	UpdateAnimations(world, dt)
	UpdateRenderFromAnimation(world)
}
