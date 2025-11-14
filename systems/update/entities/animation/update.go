package animation

import (
	"log"

	"game/components"
	"game/ecs"
	"game/pkg/config"
	"game/systems/update/entities/animation/callbacks"
	"game/systems/update/entities/animation/state"
)

// UpdateAnimations advances all animation components in the ECS world.
// This system handles:
//   - Frame advancement based on delta time
//   - State transitions via FSM
//   - Callback invocation (frame and slice callbacks)
//   - State effect management
//   - Slice extraction for hitboxes/hurtboxes
//
// Pure ECS pattern: All logic in system, component is data-only.
func UpdateAnimations(world *ecs.World, dt float64) {
	if world == nil || dt <= 0 {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		return
	}

	// Extract debug flag once for all entities
	debug := cfg.Debug

	for _, eid := range world.EntitiesWith((*components.Animation)(nil)) {
		anim := ecs.GetComponent[components.Animation](world, eid)
		if anim == nil || anim.Data == nil {
			continue
		}

		// Ensure animation state is valid
		if anim.State == "" {
			continue
		}

		// Ensure current animation is set
		if !ensureCurrentAnimation(anim, debug) {
			if debug {
				log.Printf("animation: CurrentAnimation nil for state %q; skipping Update", anim.State)
			}
			continue
		}

		// Advance frame based on delta time
		advanceFrame(anim, dt, debug)

		// Invoke callbacks for current frame
		runCallbacks(anim)
	}
}

// ensureCurrentAnimation verifies CurrentAnimation is non-nil for the current state.
// Returns true if ready; false if not recoverable this tick.
func ensureCurrentAnimation(anim *components.Animation, debug bool) bool {
	if anim.Data == nil || anim.State == "" {
		return false
	}
	if anim.Data.CurrentAnimation != nil {
		return true
	}
	// Try to recover by playing the current state
	return state.PlayState(anim, anim.State, debug)
}

// advanceFrame updates animation progress and triggers state transition if finished.
func advanceFrame(anim *components.Animation, dt float64, debug bool) {
	anim.Data.Update(float32(dt))

	if anim.Data.AnimationFinished() {
		prev := anim.State
		next, ok := anim.FSMTransitions[anim.State]
		if !ok {
			next = anim.FSMInitial
		}
		if next != "" {
			state.SetAnimationState(anim, next, debug)
		} else {
			anim.State = prev
		}
	}
}

// runCallbacks invokes frame-specific and slice-specific callbacks for the current frame.
func runCallbacks(anim *components.Animation) {
	// Frame callbacks - fire once per frame index
	frameIndex := anim.Data.CurrentFrame - anim.Data.CurrentAnimation.From
	if cb, ok := anim.FrameCallbacks[frameIndex]; ok && cb != nil {
		cb()
		delete(anim.FrameCallbacks, frameIndex)
	}

	// Slice callbacks - fire every frame the slice is present
	for sliceName, sliceCallback := range anim.SliceCallbacks {
		if sliceCallback == nil || sliceCallback.Callback == nil {
			continue
		}

		// Try to extract slice for current frame
		rect, err := callbacks.ExtractSlice(anim, sliceName, sliceCallback.FlipX, sliceCallback.FlipY)
		if err != nil {
			// Slice not present on this frame - reset firstFrame flag
			sliceCallback.FirstFrame = true
			continue
		}

		// Invoke callback with slice rect and first-frame flag
		sliceCallback.Callback(rect.X, rect.Y, rect.W, rect.H, sliceCallback.FirstFrame)
		sliceCallback.FirstFrame = false
	}
}
