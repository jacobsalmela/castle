package player

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// InputKey represents a logical game control key.
type InputKey int

const (
	InputKeyRight InputKey = iota
	InputKeyLeft
	InputKeyUp
	InputKeyDown
	InputKeyJump
	InputKeyAction
	InputKeyGuard
	InputKeyHeal
	InputKeyDash
)

// Input is a component that holds input state for an entity.
// Typically attached to the player entity.
type Input struct {
	// KeyBindings maps logical keys to physical keyboard keys.
	// Each logical key can be bound to multiple physical keys (e.g., Jump = Z or N).
	// Future: extend to support ebiten.GamepadButton, mouse buttons, etc.
	KeyBindings [9][]ebiten.Key

	// Current frame state (updated each frame by input system)
	KeyDown     [9]bool // True if key is currently held down
	KeyPressed  [9]bool // True if key was just pressed this frame
	KeyReleased [9]bool // True if key was just released this frame

	// Buffered input state (for better input feel - "coyote time" for button presses)
	// A buffered key press can be consumed within a time window after the press.
	Buffer       [9]bool    // True if this key press is buffered
	BufferExpiry [9]float64 // Game time when buffer expires (seconds since game start)
}
