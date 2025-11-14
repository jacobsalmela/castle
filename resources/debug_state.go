package resources

import (
	"sort"
	"sync"
)

// Debug category name constants for visualization overlays.
const (
	DebugCategoryCollision    = "Collision"
	DebugCategoryPhysics      = "Physics"
	DebugCategoryHitbox       = "Hitbox"
	DebugCategoryAI           = "AI"
	DebugCategoryBehaviorTree = "BehaviorTree"
	DebugCategoryStats        = "Stats"
	DebugCategoryAnim         = "Anim"
	DebugCategoryGrave        = "Grave"
	DebugCategoryEntityID     = "EntityID"
	DebugCategoryTile         = "Tiles"
	DebugCategoryLadder       = "Ladder"
	DebugCategoryECSOverlay   = "ECSOverlay"
	DebugCategoryProfiler     = "Profiler"
	DebugCategoryLighting     = "Lighting"
	DebugCategoryFakeWall     = "FakeWall"
)

// DebugState is a centralized Pure ECS resource for managing debug overlay toggles.
// Replaces the old pattern of atomic int32 toggles scattered across multiple component files.
//
// Usage:
//   - Toggle categories with Cmd+0 through Cmd+6, Shift+E, Shift+T, etc.
//   - Check enabled state in systems: debugState.IsEnabled(DebugCategoryX)
//
// Integration:
//   - Keybindings: systems/update/preupdate/debug_keybindings.go (registry)
//   - Input handling: systems/update/preupdate/debug_input.go (processes keys)
//   - Rendering: systems/draw/debug/world_space.go (world-space overlays)
//   - Rendering: systems/draw/debug/ecs_overlay.go (device-space ECS info)
//   - Rendering: systems/draw/debug/profiler.go (performance stats)
type DebugState struct {
	mu    sync.RWMutex
	flags map[string]bool
}

// NewDebugState creates a new debug state with all flags disabled.
func NewDebugState() *DebugState {
	return &DebugState{
		flags: make(map[string]bool),
	}
}

// IsEnabled checks if a debug category is enabled (thread-safe).
func (d *DebugState) IsEnabled(category string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.flags[category]
}

// Toggle flips a debug category on/off (thread-safe).
func (d *DebugState) Toggle(category string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.flags[category] = !d.flags[category]
}

// Set explicitly enables/disables a debug category (thread-safe).
func (d *DebugState) Set(category string, enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.flags[category] = enabled
}

// SetAll sets all debug categories to the same state.
func (d *DebugState) SetAll(enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Ensure all known categories exist
	categories := []string{
		DebugCategoryCollision, DebugCategoryPhysics, DebugCategoryHitbox, DebugCategoryAI,
		DebugCategoryBehaviorTree, DebugCategoryStats, DebugCategoryAnim, DebugCategoryGrave,
		DebugCategoryEntityID, DebugCategoryTile, DebugCategoryECSOverlay, DebugCategoryProfiler,
		DebugCategoryLighting, DebugCategoryFakeWall, DebugCategoryLadder,
	}
	for _, cat := range categories {
		d.flags[cat] = enabled
	}
}

// GetActive returns a list of currently enabled categories.
// Results are sorted alphabetically for consistent display order.
func (d *DebugState) GetActive() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var active []string
	for cat, enabled := range d.flags {
		if enabled {
			active = append(active, cat)
		}
	}
	sort.Strings(active)
	return active
}

// AnyEnabled returns true if any debug category is enabled.
func (d *DebugState) AnyEnabled() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, enabled := range d.flags {
		if enabled {
			return true
		}
	}
	return false
}
