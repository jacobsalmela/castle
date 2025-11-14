//go:build !release

package preupdate

import (
	"game/ecs"
	"game/resources"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// UpdateDebugInput handles debug keyboard shortcuts using the central key binding registry.
// This system runs in Phase 1 (pre-update) to capture debug input before game logic.
//
// All debug shortcuts are defined in debug_keybindings.go - the SINGLE SOURCE OF TRUTH.
// This function processes those bindings and updates the DebugState resource.
func UpdateDebugInput(world *ecs.World) {
	if world == nil {
		return
	}

	debugState := ecs.Resource[resources.DebugState](world)
	if debugState == nil {
		return
	}

	// Check Cmd+0 first (toggle all - special case)
	handleToggleAll(debugState)

	// Process all registered key bindings
	bindings := GetKeyBindings()
	for _, binding := range bindings {
		if checkBinding(binding) {
			debugState.Toggle(binding.Category)
		}
	}
}

// handleToggleAll handles Cmd+0 (toggle all world-space overlays).
// If any category is enabled, disables all. Otherwise, enables all.
func handleToggleAll(debugState *resources.DebugState) {
	cmdOrCtrl := ebiten.IsKeyPressed(ebiten.KeyMeta) || ebiten.IsKeyPressed(ebiten.KeyControl)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	// Cmd+0 (without Shift) = toggle all
	if cmdOrCtrl && !shift {
		if inpututil.IsKeyJustPressed(ebiten.KeyDigit0) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad0) {
			// Check if any category is enabled
			anyEnabled := debugState.AnyEnabled()

			// Toggle all to opposite state
			debugState.SetAll(!anyEnabled)
		}
	}
}

// checkBinding checks if a key binding's requirements are met.
// Returns true if the key was just pressed with the correct modifiers.
func checkBinding(binding DebugKeyBinding) bool {
	// Check modifiers
	cmdOrCtrl := ebiten.IsKeyPressed(ebiten.KeyMeta) || ebiten.IsKeyPressed(ebiten.KeyControl)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	// Verify Cmd modifier requirement
	if binding.Modifiers.Cmd && !cmdOrCtrl {
		return false // Cmd required but not pressed
	}
	if !binding.Modifiers.Cmd && cmdOrCtrl {
		return false // Cmd is pressed but not required
	}

	// Verify Shift modifier requirement
	if binding.Modifiers.Shift && !shift {
		return false // Shift required but not pressed
	}
	if !binding.Modifiers.Shift && shift && binding.Modifiers.Cmd {
		return false // Shift is pressed but not required (prevents Cmd+Shift+1 triggering Cmd+1)
	}

	// Check if key was just pressed (primary or alt key)
	keyPressed := inpututil.IsKeyJustPressed(binding.Key)
	if binding.AltKey != 0 {
		keyPressed = keyPressed || inpututil.IsKeyJustPressed(binding.AltKey)
	}

	return keyPressed
}
