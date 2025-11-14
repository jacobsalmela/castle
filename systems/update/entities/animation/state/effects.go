package state

import (
	"slices"

	"game/components"
)

// SetStateEffect applies a temporary state modification with a restore function.
// The effect persists across the specified states and is restored when leaving them.
func SetStateEffect(anim *components.Animation, applyAndGetRestore func() func(), forStates ...string) {
	if anim == nil || applyAndGetRestore == nil {
		return
	}

	// Clear previous effect if exists
	if anim.StateEffect != nil {
		anim.StateEffect.Restore()
	}

	// Apply new effect
	restore := applyAndGetRestore()
	anim.StateEffect = &components.AnimationStateEffect{
		Restore: restore,
		States:  forStates,
	}
}

// clearCallbacks clears frame and slice callbacks when leaving a state.
// Also restores state effects if the new state is not in the effect's state list.
// This is called internally by SetAnimationState.
func clearCallbacks(anim *components.Animation) {
	if anim == nil {
		return
	}

	// Restore state effect if leaving affected states
	if anim.StateEffect != nil && !slices.Contains(anim.StateEffect.States, anim.State) {
		anim.StateEffect.Restore()
		anim.StateEffect = nil
	}

	// Clear all callbacks
	anim.FrameCallbacks = make(map[int]func())
	// Clear slice callbacks - they are per-attack/per-state, not persistent
	anim.SliceCallbacks = make(map[string]*components.AnimationSliceCallback)
}
