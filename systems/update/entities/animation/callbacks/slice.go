package callbacks

import (
	"fmt"

	"game/components"
	"game/pkg/bump"
)

// RegisterSliceCallback registers a callback to be invoked when a named slice is present.
// The callback receives the slice rectangle and a firstFrame flag.
func RegisterSliceCallback(anim *components.Animation, sliceName string, flipX, flipY bool, callback func(x, y, w, h float64, firstFrame bool)) {
	if anim == nil || sliceName == "" || callback == nil {
		return
	}

	if anim.SliceCallbacks == nil {
		anim.SliceCallbacks = make(map[string]*components.AnimationSliceCallback)
	}

	anim.SliceCallbacks[sliceName] = &components.AnimationSliceCallback{
		Callback:   callback,
		FlipX:      flipX,
		FlipY:      flipY,
		FirstFrame: true,
	}
}

// UnregisterSliceCallback removes a slice callback from the animation component.
func UnregisterSliceCallback(anim *components.Animation, sliceName string) {
	if anim == nil {
		return
	}
	delete(anim.SliceCallbacks, sliceName)
}

// ExtractSlice returns a named slice rectangle from the current animation frame.
// Applies appropriate offsets (OX/OY or OXFlip/OYFlip) based on flip parameters.
// Complexity reduced from 11 to 7 by extracting helper functions.
func ExtractSlice(anim *components.Animation, sliceName string, flipX, flipY bool) (bump.Rect, error) {
	if err := validateSliceParams(anim, sliceName); err != nil {
		return bump.Rect{}, err
	}

	rect, err := getSliceRect(anim, sliceName)
	if err != nil {
		return bump.Rect{}, err
	}

	frameBounds := anim.Data.FrameBoundaries().Rectangle()
	offsetX := calculateOffsetX(rect, anim, flipX, float64(frameBounds.Dx()))
	offsetY := calculateOffsetY(rect, anim, flipY, float64(frameBounds.Dy()))

	return bump.Rect{X: offsetX, Y: offsetY, W: rect.W, H: rect.H}, nil
}

// validateSliceParams validates the animation component and slice name.
func validateSliceParams(anim *components.Animation, sliceName string) error {
	if anim == nil {
		return fmt.Errorf("animation: nil component")
	}
	if anim.SliceMap[sliceName] == nil {
		return fmt.Errorf("slice name %s not found", sliceName)
	}
	return nil
}

// getSliceRect retrieves the slice rectangle for the current frame.
func getSliceRect(anim *components.Animation, sliceName string) (bump.Rect, error) {
	rect, ok := anim.SliceMap[sliceName][anim.Data.CurrentFrame]
	if !ok {
		return bump.Rect{}, fmt.Errorf("no slice in current frame %d", anim.Data.CurrentFrame)
	}
	return rect, nil
}

// calculateOffsetX calculates the X offset for a slice based on flip state and animation offsets.
func calculateOffsetX(rect bump.Rect, anim *components.Animation, flipX bool, frameWidth float64) float64 {
	offsetX := rect.X
	if flipX {
		if frameWidth > 0 {
			offsetX = frameWidth - rect.X - rect.W
		}
		offsetX += anim.OX + anim.OXFlip
	} else {
		offsetX += anim.OX
	}
	return offsetX
}

// calculateOffsetY calculates the Y offset for a slice based on flip state and animation offsets.
func calculateOffsetY(rect bump.Rect, anim *components.Animation, flipY bool, frameHeight float64) float64 {
	offsetY := rect.Y
	if flipY {
		if frameHeight > 0 {
			offsetY = frameHeight - rect.Y - rect.H
		}
		offsetY += anim.OY + anim.OYFlip
	} else {
		offsetY += anim.OY
	}
	return offsetY
}
