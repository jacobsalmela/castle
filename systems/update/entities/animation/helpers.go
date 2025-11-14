package animation

import (
	"game/components"
	"game/ecs"
	"game/pkg/bump"
	"game/pkg/config"
	"game/systems/update/entities/animation/callbacks"
	"game/systems/update/entities/animation/state"
)

// HasState returns true if the animation data defines the given state.
func HasState(anim *components.Animation, animState string) bool {
	if anim == nil || anim.Data == nil || animState == "" {
		return false
	}
	return anim.Data.Animation(animState) != nil
}

// Convenience wrappers for backward compatibility and ease of use.
// These functions provide simpler signatures for common cases.

// SetAnimationState is a convenience wrapper that changes animation state without debug logging.
// For debug-enabled state changes, use SetAnimationStateDebug or state.SetAnimationState directly.
func SetAnimationState(anim *components.Animation, animState string) {
	state.SetAnimationState(anim, animState, false)
}

// SetAnimationStateDebug changes animation state with debug logging enabled if configured.
// This is the recommended function to use from systems that have access to world.
func SetAnimationStateDebug(world *ecs.World, anim *components.Animation, animState string) {
	cfg := ecs.Resource[config.Config](world)
	debug := cfg != nil && cfg.Debug
	state.SetAnimationState(anim, animState, debug)
}

// PlayState is a convenience wrapper that plays an animation state without debug logging.
// For debug-enabled state playback, use PlayStateDebug or state.PlayState directly.
func PlayState(anim *components.Animation, animState string) bool {
	return state.PlayState(anim, animState, false)
}

// PlayStateDebug plays an animation state with debug logging enabled if configured.
// This is the recommended function to use from systems that have access to world.
func PlayStateDebug(world *ecs.World, anim *components.Animation, animState string) bool {
	cfg := ecs.Resource[config.Config](world)
	debug := cfg != nil && cfg.Debug
	return state.PlayState(anim, animState, debug)
}

// SetStateEffect is a convenience wrapper for state.SetStateEffect.
func SetStateEffect(anim *components.Animation, applyAndGetRestore func() func(), forStates ...string) {
	state.SetStateEffect(anim, applyAndGetRestore, forStates...)
}

// RegisterFrameCallback is a convenience wrapper for callbacks.RegisterFrameCallback.
func RegisterFrameCallback(anim *components.Animation, frame int, callback func()) {
	callbacks.RegisterFrameCallback(anim, frame, callback)
}

// RegisterSliceCallback is a convenience wrapper for callbacks.RegisterSliceCallback.
func RegisterSliceCallback(anim *components.Animation, sliceName string, flipX, flipY bool, callback func(x, y, w, h float64, firstFrame bool)) {
	callbacks.RegisterSliceCallback(anim, sliceName, flipX, flipY, callback)
}

// UnregisterSliceCallback is a convenience wrapper for callbacks.UnregisterSliceCallback.
func UnregisterSliceCallback(anim *components.Animation, sliceName string) {
	callbacks.UnregisterSliceCallback(anim, sliceName)
}

// ExtractSlice is a convenience wrapper for callbacks.ExtractSlice.
func ExtractSlice(anim *components.Animation, sliceName string, flipX, flipY bool) (bump.Rect, error) {
	return callbacks.ExtractSlice(anim, sliceName, flipX, flipY)
}
