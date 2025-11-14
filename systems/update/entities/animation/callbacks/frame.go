package callbacks

import "game/components"

// RegisterFrameCallback registers a callback to be invoked on a specific frame index.
// Frame index is relative to the animation start (0-based).
func RegisterFrameCallback(anim *components.Animation, frame int, callback func()) {
	if anim == nil || callback == nil {
		return
	}

	if anim.FrameCallbacks == nil {
		anim.FrameCallbacks = make(map[int]func())
	}
	anim.FrameCallbacks[frame] = callback
}
