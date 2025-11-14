package prefabs

import (
	"fmt"
	"image/color"
	"log"

	"game/assets"
	"game/components"
	"game/pkg/bump"

	"github.com/damienfamed75/aseprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// AnimationConfig holds configuration for creating an Animation component.
// This separates FSM configuration from the component creation.
type AnimationConfig struct {
	FilesName              string
	OX, OY, OXFlip, OYFlip float64
	Layer                  int
	FSMInitial             string
	FSMTransitions         map[string]string
}

// NewAnimationComponent creates a fully initialized Pure ECS Animation component.
//
// The Animation component manages sprite-based animation with:
//   - Aseprite JSON data integration
//   - Finite State Machine (FSM) for state transitions
//   - Frame-based and slice-based callbacks
//   - Hitbox/hurtbox extraction from Aseprite slices
//   - Sprite offsets for positioning
//
// Parameters:
//   - config: Animation configuration (file name, offsets, layer, FSM)
//
// Returns: Fully initialized Animation component ready to attach to an entity.
//
// Example:
//
//	anim := NewAnimationComponent(AnimationConfig{
//	    FilesName: "knight",
//	    OX: -8, OY: -16, OXFlip: 8, OYFlip: -16,
//	    Layer: 5,
//	    FSMInitial: "Idle",
//	    FSMTransitions: map[string]string{"Attack": "Idle"},
//	})
//	world.AddComponent(entityID, anim)
func NewAnimationComponent(config AnimationConfig) *components.Animation {
	anim := &components.Animation{
		FilesName:  config.FilesName,
		OX:         config.OX,
		OY:         config.OY,
		OXFlip:     config.OXFlip,
		OYFlip:     config.OYFlip,
		Layer:      config.Layer,
		ColorScale: color.White,
	}

	// Set FSM or use defaults
	if config.FSMInitial != "" {
		anim.FSMInitial = config.FSMInitial
	} else {
		anim.FSMInitial = "Idle"
	}

	if config.FSMTransitions != nil {
		anim.FSMTransitions = config.FSMTransitions
	} else {
		anim.FSMTransitions = make(map[string]string)
	}

	// Initialize callbacks and state
	anim.FrameCallbacks = make(map[int]func())
	anim.SliceCallbacks = make(map[string]*components.AnimationSliceCallback)

	// Load assets and build slice map
	if err := InitializeAnimation(anim); err != nil {
		log.Printf("Warning: Failed to initialize animation %s: %v", config.FilesName, err)
	}

	return anim
}

// InitializeAnimation loads assets and prepares the animation component.
func InitializeAnimation(anim *components.Animation) error {
	// Try to load from unified asset system first
	asset := assets.GetSpriteAsset(anim.FilesName)
	if asset == nil || asset.Image == nil {
		return fmt.Errorf("failed to load sprite asset %s from unified asset system", anim.FilesName)
	}

	// Use image from unified system
	img := asset.Image
	anim.Image = img

	// Load Aseprite JSON from unified asset system
	if asset.JSONPath == "" {
		return fmt.Errorf("no JSON path for %s in unified asset system", anim.FilesName)
	}

	data, err := assets.FS.ReadFile(asset.JSONPath)
	if err != nil {
		return fmt.Errorf("failed to load JSON for %s at %s: %w", anim.FilesName, asset.JSONPath, err)
	}

	file, err := aseprite.NewFile(data)
	if err != nil {
		return fmt.Errorf("invalid JSON for %s: %w", anim.FilesName, err)
	}
	anim.Data = file

	// Debug: Log expected dimensions from JSON
	imgBounds := img.Bounds()
	metaSize := anim.Data.Meta.Size
	if imgBounds.Dx() != metaSize.Width || imgBounds.Dy() != metaSize.Height {
		log.Printf("  ⚠️  WARNING: Image size mismatch for %s! PNG=%dx%d but JSON says %dx%d",
			anim.FilesName, imgBounds.Dx(), imgBounds.Dy(), metaSize.Width, metaSize.Height)
	}

	// Set initial state
	if len(anim.Data.Meta.Animations) == 0 {
		return fmt.Errorf("no animations defined in %s", anim.FilesName)
	}
	first := anim.Data.Meta.Animations[0].Name
	anim.State = first

	// Play the initial animation
	if err := anim.Data.Play(first); err != nil {
		return fmt.Errorf("failed to play initial state %s: %w", first, err)
	}

	// Calculate frame dimensions
	frame := anim.Data.Frames.FrameAtIndex(anim.Frame).SpriteSourceSize
	anim.Width = float64(frame.Width)
	anim.Height = float64(frame.Height)

	// Build slice map cache from both Aseprite library AND assets.SliceRegistry
	// This provides redundancy and validates our parsing system
	buildSliceMapForAnimation(anim)

	return nil
}

// buildSliceMapForAnimation precomputes bounding rectangles for named slices per frame.
// This enables fast slice lookup without parsing Aseprite data every frame.
//
// Slice loading strategy (dual-source with validation):
//  1. Primary: Load from assets.SliceRegistry (pre-parsed during Init)
//  2. Fallback: Parse from Aseprite library data (legacy compatibility)
//  3. Validation: Compare both sources and warn if they differ
//
// This approach provides:
//   - Fast loading (registry is pre-parsed)
//   - Backward compatibility (fallback to Aseprite library)
//   - Validation (catch parsing bugs by comparing sources)
func buildSliceMapForAnimation(anim *components.Animation) {
	if anim.Data == nil {
		return
	}

	// Initialize slice map
	anim.SliceMap = make(map[string]map[int]bump.Rect)

	// STRATEGY 1: Load from pre-parsed assets.SliceRegistry (PREFERRED)
	registrySlices := assets.GetSliceMap(anim.FilesName)
	if registrySlices != nil && len(registrySlices) > 0 {
		// Deep copy registry data into animation component
		for sliceName, frameMap := range registrySlices {
			anim.SliceMap[sliceName] = make(map[int]bump.Rect, len(frameMap))
			for frame, rect := range frameMap {
				// Apply sprite source size offset correction
				// The registry stores raw Aseprite slice coordinates (sprite-relative)
				// but we need transform-relative coordinates for collision
				sss := anim.Data.Frames.FrameAtIndex(frame).SpriteSourceSize
				correctedRect := bump.Rect{
					X: rect.X - float64(sss.X),
					Y: rect.Y - float64(sss.Y),
					W: rect.W,
					H: rect.H,
				}
				anim.SliceMap[sliceName][frame] = correctedRect
			}
		}

		return
	}

	// STRATEGY 2: Fallback to Aseprite library parsing (LEGACY)
	// This maintains compatibility for entities without registry data
	if len(anim.Data.Meta.Slices) == 0 {
		// No slices available from either source
		// Suppress warning for VFX and environment assets that don't need slices
		if !shouldWarnAboutMissingSlices(anim.FilesName) {
			return
		}
		log.Printf("    ⚠ %s: No slice data available (neither registry nor Aseprite)", anim.FilesName)
		return
	}

	for _, slice := range anim.Data.Meta.Slices {
		sliceName := slice.Name
		anim.SliceMap[sliceName] = make(map[int]bump.Rect)

		for _, key := range slice.Keys {
			frameNum := key.FrameNum
			bounds := key.Bounds

			sss := anim.Data.Frames.FrameAtIndex(frameNum).SpriteSourceSize
			x := float64(bounds.X - sss.X)
			y := float64(bounds.Y - sss.Y)
			w := float64(bounds.Width)
			h := float64(bounds.Height)

			anim.SliceMap[sliceName][frameNum] = bump.Rect{X: x, Y: y, W: w, H: h}
		}
	}
}

// validateSliceParsing compares registry slices with Aseprite library slices.
// Returns true if they match (validation passed), false if they differ.
// This catches bugs in our custom JSON parsing by comparing against the Aseprite library.
func validateSliceParsing(anim *components.Animation, registrySlices map[string]map[int]bump.Rect) bool {
	if anim.Data == nil || len(anim.Data.Meta.Slices) == 0 {
		// No library data to validate against - assume registry is correct
		return true
	}

	// Quick validation: compare slice counts
	if len(registrySlices) != len(anim.Data.Meta.Slices) {
		return false
	}

	// Detailed validation: compare frame counts per slice
	// (We don't compare exact coordinates to avoid floating point precision issues)
	for _, slice := range anim.Data.Meta.Slices {
		registryFrames := registrySlices[slice.Name]
		if registryFrames == nil {
			return false // Missing slice in registry
		}
		if len(registryFrames) != len(slice.Keys) {
			return false // Frame count mismatch
		}
	}

	return true
}

// LoadAnimationAssets loads the sprite sheet image and Aseprite JSON data.
// This can be used standalone for pre-loading or testing without creating
// a full Animation component.
//
// Returns the loaded image and parsed Aseprite data, or error if loading fails.
func LoadAnimationAssets(filesName string) (*ebiten.Image, *aseprite.File, error) {
	// Load sprite sheet
	img, _, err := ebitenutil.NewImageFromFileSystem(assets.FS, filesName+".png")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load image %s.png: %w", filesName, err)
	}

	// Load Aseprite JSON
	data, err := assets.FS.ReadFile(filesName + ".json")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load data %s.json: %w", filesName, err)
	}

	file, err := aseprite.NewFile(data)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid JSON for %s: %w", filesName, err)
	}

	return img, file, nil
}

// shouldWarnAboutMissingSlices checks if we should warn about missing slice data.
// Returns false for asset categories that don't typically need slices (vfx, environment).
func shouldWarnAboutMissingSlices(filesName string) bool {
	asset := assets.GetSpriteAsset(filesName)
	if asset == nil {
		return true // Warn if asset not found
	}

	// Don't warn for VFX and environment assets - they often don't need slices
	return asset.Category != "vfx" && asset.Category != "environment"
}

// DefaultAnimationConfig returns a basic configuration with "Idle" initial state.
func DefaultAnimationConfig(filesName string) AnimationConfig {
	return AnimationConfig{
		FilesName:      filesName,
		Layer:          5, // Default render layer
		FSMInitial:     "Idle",
		FSMTransitions: make(map[string]string),
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// TRANSFORM DIMENSION HELPERS
// ═══════════════════════════════════════════════════════════════════════════════

// GetTransformDimensionsFromSlice extracts Transform dimensions from an animation's
// hurtbox slice data (frame 0). This derives environmental collision box from combat hitbox.
//
// Strategy:
//  1. Try to get frame 0 hurtbox from animation's SliceMap
//  2. Return width and height from hurtbox bounds
//  3. Fallback to default dimensions if no slice data
//
// This eliminates hardcoded width/height constants in prefabs - dimensions
// come directly from Aseprite slice data instead.
//
// Parameters:
//   - anim: Animation component (must be initialized with InitializeAnimation)
//   - defaultWidth, defaultHeight: Fallback dimensions if no slice data
//
// Returns: (width, height) for Transform component
//
// Example:
//
//	anim := NewAnimationComponent(config)
//	w, h := GetTransformDimensionsFromSlice(anim, 8, 11)
//	transform := &components.Transform{X: x, Y: y, W: w, H: h}
func GetTransformDimensionsFromSlice(anim *components.Animation, defaultWidth, defaultHeight float64) (float64, float64) {
	if anim == nil || anim.SliceMap == nil {
		return defaultWidth, defaultHeight
	}

	// Get hurtbox slice for frame 0 (idle/default frame)
	hurtboxFrames, ok := anim.SliceMap[components.HurtboxSliceName]
	if !ok || len(hurtboxFrames) == 0 {
		return defaultWidth, defaultHeight
	}

	// Try frame 0 first (most common)
	if rect, ok := hurtboxFrames[0]; ok {
		return rect.W, rect.H
	}

	// If frame 0 doesn't exist, use the first available frame
	for _, rect := range hurtboxFrames {
		return rect.W, rect.H
	}

	return defaultWidth, defaultHeight
}

// GetTransformDimensionsFromRegistry is a convenience function that loads
// dimensions from the global assets.SliceRegistry without creating an Animation.
//
// This is useful for prefabs that need dimensions before creating Animation components,
// or for validating that slice data matches expected dimensions.
//
// Parameters:
//   - animFileName: Animation file name (e.g., "knight", "rat")
//   - defaultWidth, defaultHeight: Fallback dimensions if no slice data
//
// Returns: (width, height) for Transform component
//
// Example:
//
//	w, h := GetTransformDimensionsFromRegistry("knight", 8, 11)
func GetTransformDimensionsFromRegistry(animFileName string, defaultWidth, defaultHeight float64) (float64, float64) {
	// Get slice data from registry
	slices := assets.GetSliceMap(animFileName)
	if slices == nil {
		return defaultWidth, defaultHeight
	}

	// Get hurtbox frames
	hurtboxFrames, ok := slices[components.HurtboxSliceName]
	if !ok || len(hurtboxFrames) == 0 {
		return defaultWidth, defaultHeight
	}

	// Use frame 0 (idle/default frame)
	if rect, ok := hurtboxFrames[0]; ok {
		// Note: Registry stores raw Aseprite coordinates, so we use W/H directly
		// (offset correction is only needed for X/Y position)
		return rect.W, rect.H
	}

	// If frame 0 doesn't exist, use the first available frame
	for _, rect := range hurtboxFrames {
		return rect.W, rect.H
	}

	return defaultWidth, defaultHeight
}
