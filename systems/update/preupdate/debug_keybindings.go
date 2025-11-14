//go:build !release

package preupdate

import (
	"game/resources"

	"github.com/hajimehoshi/ebiten/v2"
)

// KeyModifiers represents modifier key requirements for a debug shortcut.
type KeyModifiers struct {
	Cmd   bool
	Shift bool
}

// DebugKeyBinding maps a key combination to a debug category.
// This is the central registry for all debug keyboard shortcuts.
type DebugKeyBinding struct {
	Category    string // Debug category name (matches resources.DebugCategory* constants)
	Key         ebiten.Key
	AltKey      ebiten.Key // Numpad alternative (0 if none)
	Modifiers   KeyModifiers
	Description string // Human-readable description for help text
}

// debugKeyBindings is the SINGLE SOURCE OF TRUTH for debug keyboard shortcuts.
//
// To add a new debug category:
//  1. Add entry to debugKeyBindings array in this file (single source of truth)
//  2. Add category constant to resources/debug_state.go (if new category)
//  3. Add draw logic to appropriate systems/draw/debug/world_space.go function
//  4. Describe binding in debug overlay (systems/draw/debug/ecs_overlay.go)
//
// That's it! No need to create new component files or update multiple places.
var debugKeyBindings = []DebugKeyBinding{
	// ═══════════════════════════════════════════════════════════════════════
	// Cmd+Number (world-space debug overlays)
	// ═══════════════════════════════════════════════════════════════════════
	{
		Category:    resources.DebugCategoryCollision,
		Key:         ebiten.KeyDigit1,
		AltKey:      ebiten.KeyNumpad1,
		Modifiers:   KeyModifiers{Cmd: true},
		Description: "Collision boxes and events",
	},
	{
		Category:    resources.DebugCategoryPhysics,
		Key:         ebiten.KeyDigit2,
		AltKey:      ebiten.KeyNumpad2,
		Modifiers:   KeyModifiers{Cmd: true},
		Description: "Physics bodies and velocity",
	},
	{
		Category:    resources.DebugCategoryHitbox,
		Key:         ebiten.KeyDigit3,
		AltKey:      ebiten.KeyNumpad3,
		Modifiers:   KeyModifiers{Cmd: true},
		Description: "Combat hitboxes",
	},
	{
		Category:    resources.DebugCategoryAI,
		Key:         ebiten.KeyDigit4,
		AltKey:      ebiten.KeyNumpad4,
		Modifiers:   KeyModifiers{Cmd: true},
		Description: "AI state and targeting",
	},
	{
		Category:    resources.DebugCategoryStats,
		Key:         ebiten.KeyDigit5,
		AltKey:      ebiten.KeyNumpad5,
		Modifiers:   KeyModifiers{Cmd: true},
		Description: "Health/Stamina/Poise",
	},
	{
		Category:    resources.DebugCategoryAnim,
		Key:         ebiten.KeyDigit6,
		AltKey:      ebiten.KeyNumpad6,
		Modifiers:   KeyModifiers{Cmd: true},
		Description: "Animation states",
	},
	{
		Category:    resources.DebugCategoryBehaviorTree,
		Key:         ebiten.KeyDigit7,
		AltKey:      ebiten.KeyNumpad7,
		Modifiers:   KeyModifiers{Cmd: true},
		Description: "Behavior tree visualization",
	},
	// Cmd+8 reserved for future use
	// Cmd+9 reserved for future use
	// Cmd+0 is "toggle all" - handled separately

	// ═══════════════════════════════════════════════════════════════════════
	// Shift+Letter
	// ═══════════════════════════════════════════════════════════════════════
	{
		Category:    resources.DebugCategoryEntityID,
		Key:         ebiten.KeyE,
		AltKey:      0,
		Modifiers:   KeyModifiers{Shift: true},
		Description: "Entity IDs",
	},
	{
		Category:    resources.DebugCategoryTile,
		Key:         ebiten.KeyT,
		AltKey:      0,
		Modifiers:   KeyModifiers{Shift: true},
		Description: "Tile grid",
	},
	{
		Category:    resources.DebugCategoryLadder,
		Key:         ebiten.KeyL,
		AltKey:      0,
		Modifiers:   KeyModifiers{Shift: true},
		Description: "Ladder tiles",
	},
	{
		Category:    resources.DebugCategoryFakeWall,
		Key:         ebiten.KeyF,
		AltKey:      0,
		Modifiers:   KeyModifiers{Shift: true},
		Description: "Fake wall collision debug",
	},

	// ═══════════════════════════════════════════════════════════════════════
	// No modifier keys (simple toggles)
	// ═══════════════════════════════════════════════════════════════════════
	{
		Category:    resources.DebugCategoryLighting,
		Key:         ebiten.KeyL,
		AltKey:      0,
		Modifiers:   KeyModifiers{},
		Description: "Dynamic lighting shader",
	},

	// ═══════════════════════════════════════════════════════════════════════
	// Cmd+Shift+Number (device-space overlays)
	// ═══════════════════════════════════════════════════════════════════════
	{
		Category:    resources.DebugCategoryProfiler,
		Key:         ebiten.KeyDigit0,
		AltKey:      ebiten.KeyNumpad0,
		Modifiers:   KeyModifiers{Cmd: true, Shift: true},
		Description: "Performance profiler",
	},
	{
		Category:    resources.DebugCategoryECSOverlay,
		Key:         ebiten.KeyDigit1,
		AltKey:      ebiten.KeyNumpad1,
		Modifiers:   KeyModifiers{Cmd: true, Shift: true},
		Description: "ECS sprite count",
	},
}

// GetKeyBindings returns the complete key binding registry.
func GetKeyBindings() []DebugKeyBinding {
	return debugKeyBindings
}

// GetBindingForCategory finds the key binding for a specific category.
// Returns nil if not found.
func GetBindingForCategory(category string) *DebugKeyBinding {
	for i := range debugKeyBindings {
		if debugKeyBindings[i].Category == category {
			return &debugKeyBindings[i]
		}
	}
	return nil
}
