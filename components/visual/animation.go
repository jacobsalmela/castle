// TODO: make pure data, move logic to systems.
package visual

import (
	"game/pkg/bump"
	"image/color"
	"math"

	"github.com/damienfamed75/aseprite"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// Animation State Tags
// These constants define the standard animation states used throughout the game.
const (
	IdleTag       = "Idle"
	WalkTag       = "Walk"
	AttackTag     = "Attack"
	BlockTag      = "Block"
	ParryBlockTag = "ParryBlock"
	StaggerTag    = "Stagger"
	ClimbTag      = "Climb"
	ConsumeTag    = "Consume"
)

// Animation Slice Names
// These constants define the standard slice names in Aseprite files for hitboxes, hurtboxes, etc.
const (
	HurtboxSliceName = "hurtbox"
	HitboxSliceName  = "hitbox"
	BlockSliceName   = "blockbox"
	HealSliceName    = "healbox"
)

// UberColor is a high-range color type used for flash effects.
type UberColor struct{ R, G, B, A uint32 }

func (c UberColor) RGBA() (uint32, uint32, uint32, uint32) { return c.R, c.G, c.B, c.A }

var (
	// WhiteScalerColor is a high-intensity white used for flash effects.
	WhiteScalerColor               = UberColor{0xffff * 4, 0xffff * 4, 0xffff * 4, 0xffff * 4}
	legacyAnimNormalMaskColor      = color.NRGBA{127, 127, 255, 255}
	legacyAnimFillNormalMaskColorM = colorm.ColorM{}
)

func init() {
	legacyAnimFillNormalMaskColorM.Scale(0, 0, 0, 1)
	r, g, b := float64(legacyAnimNormalMaskColor.R)/math.MaxUint8, float64(legacyAnimNormalMaskColor.G)/math.MaxUint8, float64(legacyAnimNormalMaskColor.B)/math.MaxUint8
	legacyAnimFillNormalMaskColorM.Translate(r, g, b, 0)
}

// FillNormalMaskColorM re-exports the legacy anim preset used for writing to the normal map.
// Keeping it here lets draw systems depend only on the components package during migration.
var FillNormalMaskColorM = legacyAnimFillNormalMaskColorM

// NormalMaskColor re-exports the base normal mask color used by the shader path.
var NormalMaskColor = legacyAnimNormalMaskColor

// Animation is a Pure ECS animation component (data only).
// All animation logic is handled by systems/update/animation.go.
//
// Animation handles sprite sheet playback from Aseprite files, including:
//   - Frame-by-frame advancement based on delta time
//   - State machine transitions (idle → walk → attack → idle)
//   - Slice extraction for hitboxes/hurtboxes
//   - Frame-specific and slice-specific callbacks
//   - Rendering offsets and layer configuration
type Animation struct {
	// === ASSET REFERENCES ===
	// FilesName is the base filename (without extension) for the Aseprite files.
	// Example: "knight" loads "knight.png" and "knight.json"
	FilesName string

	// Image is the loaded sprite sheet texture.
	Image *ebiten.Image

	// Data is the parsed Aseprite metadata (animations, frames, slices).
	Data *aseprite.File

	// === PLAYBACK STATE ===
	// State is the current animation state tag (e.g., "idle", "walk", "attack").
	State string

	// Frame is the current frame index within the animation.
	Frame int

	// Timer is the time accumulator for frame advancement (in seconds).
	Timer float64

	// === FSM (FINITE STATE MACHINE) ===
	// FSMInitial is the default state to return to after animations finish.
	FSMInitial string

	// FSMTransitions maps current state → next state when animation completes.
	// Example: {"attack": "idle", "walk": "idle"}
	// If no transition is defined, returns to FSMInitial.
	FSMTransitions map[string]string

	// === RENDERING ===
	// OX, OY are sprite offsets when not flipped (normal facing).
	OX, OY float64

	// OXFlip, OYFlip are sprite offsets when flipped (opposite facing).
	OXFlip, OYFlip float64

	// Layer determines render order (higher = drawn later/on top).
	Layer int

	// ColorScale is the color modulation/tint applied to the sprite.
	// nil = no tint (normal rendering)
	ColorScale color.Color

	// === COMPUTED PROPERTIES (set by system) ===
	// Width, Height are the dimensions of the current animation frame.
	// Updated by the animation system each frame.
	Width, Height float64

	// === CALLBACK MANAGEMENT (Pure ECS approach) ===
	// SliceCallbacks maps slice names to callback data.
	// System invokes callbacks when named slices are present on current frame.
	// Example: "hitbox" → callback fired when hitbox slice exists on frame
	SliceCallbacks map[string]*AnimationSliceCallback

	// FrameCallbacks maps frame indices (relative to animation start) to callbacks.
	// System invokes callbacks when specific frames are reached.
	// Example: frame 5 → play sound effect
	FrameCallbacks map[int]func()

	// === STATE EFFECTS (temporary overrides) ===
	// StateEffect stores a temporary modification (e.g., animation speed change)
	// with a restore function and list of states it applies to.
	StateEffect *AnimationStateEffect

	// === SLICE CACHE (precomputed for fast lookup) ===
	// SliceMap caches slice rectangles by name and frame for fast extraction.
	// Built once during initialization, used by systems for hitbox/hurtbox queries.
	// map[sliceName]map[frameIndex]Rect
	SliceMap map[string]map[int]bump.Rect
}

// AnimationSliceCallback holds callback data for slice-based events.
type AnimationSliceCallback struct {
	// Callback is the function to invoke when slice is present.
	Callback func(x, y, w, h float64, firstFrame bool)

	// FlipX, FlipY indicate whether to apply flipped offsets.
	FlipX, FlipY bool

	// FirstFrame tracks if this is the first frame the slice appeared
	// (for one-time triggers like playing sounds).
	FirstFrame bool
}

// AnimationStateEffect represents a temporary state modification.
type AnimationStateEffect struct {
	// Restore is the function to call when exiting the affected states.
	Restore func()

	// States is the list of animation states this effect applies to.
	// When transitioning away from these states, Restore is called.
	States []string
}
