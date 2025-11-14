package state

import (
	"log"

	"game/components"
)

// SetAnimationState changes the current animation state.
// Handles state transitions, callback cleanup, and state effects.
func SetAnimationState(anim *components.Animation, state string, debug bool) {
	if anim == nil {
		return
	}

	if anim.State == state {
		// Already in this state. If the underlying animation pointer was lost, replay it.
		if anim.Data != nil && anim.Data.CurrentAnimation == nil {
			_ = PlayState(anim, state, debug)
		}
		return
	}

	prev := anim.State
	anim.State = state

	if !playState(anim, state, debug) {
		// Failed to play - revert to previous state
		anim.State = prev
		return
	}

	clearCallbacks(anim)
}

// PlayState attempts to play the given state on the aseprite data.
// Returns true if the animation exists and was successfully played.
// The debug parameter controls whether missing states are logged.
func PlayState(anim *components.Animation, state string, debug bool) bool {
	return playState(anim, state, debug)
}

// playState is the internal implementation of PlayState.
// It attempts to play the given animation state and logs errors if debug is enabled.
func playState(anim *components.Animation, state string, debug bool) bool {
	if anim == nil || anim.Data == nil {
		return false
	}
	if anim.Data.Animation(state) == nil {
		if debug && state != components.StaggerTag {
			// Be quiet for missing Stagger to avoid noisy logs on assets without that state.
			log.Printf("animation: missing state %s", state)
		}
		return false
	}
	if err := anim.Data.Play(state); err != nil {
		if debug {
			log.Printf("animation: %s", err)
		}
		return false
	}
	return true
}
