package resources

import (
	"image/color"
	"sync"
	"time"
)

// DebugCategory represents a debug category with its configuration.
type DebugCategory struct {
	Name         string
	ConsoleColor string // ANSI color code for console output
	VisualColor  color.RGBA
	ThrottleMS   int // Milliseconds between log outputs (0 = every frame)
	LastLogTime  time.Time
}

// DebugCategories is a singleton ECS resource that stores all debug category info.
// This replaces the global map in pkg/debug with Pure ECS architecture.
//
// Usage:
//   - Get resource: cats := ecs.Resource[resources.DebugCategories](world)
//   - Get color: color, ok := cats.GetColor("Collision")
//   - Check throttle: shouldLog := cats.ShouldLog("AI")
type DebugCategories struct {
	categories map[string]*DebugCategory
	mu         sync.RWMutex
}

// NewDebugCategories creates and initializes the debug categories resource.
// Call this during game initialization and add to ECS world as a resource.
func NewDebugCategories() *DebugCategories {
	dc := &DebugCategories{
		categories: make(map[string]*DebugCategory),
	}

	// Register all categories with their colors (matching pkg/debug colors)
	// ANSI color codes:
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorCyan   = "\033[36m"
		colorWhite  = "\033[37m"
		colorOrange = "\033[38;5;208m"
		colorPink   = "\033[38;5;213m"
	)

	dc.Register("Collision", colorCyan, color.RGBA{0, 255, 255, 255}, 1000)      // Cyan - bump/collision space
	dc.Register("Body", colorGreen, color.RGBA{0, 255, 0, 255}, 1000)            // Green - body physics
	dc.Register("Hitbox", colorRed, color.RGBA{255, 0, 0, 255}, 1000)            // Red - hitboxes
	dc.Register("AI", colorYellow, color.RGBA{255, 255, 0, 255}, 1000)           // Yellow - AI logic
	dc.Register("Stats", colorPurple, color.RGBA{255, 0, 255, 255}, 1000)        // Purple - stats/health
	dc.Register("Anim", colorBlue, color.RGBA{0, 128, 255, 255}, 1000)           // Blue - animations
	dc.Register("BehaviorTree", colorBlue, color.RGBA{200, 200, 200, 255}, 0)    // Light gray - behavior trees
	dc.Register("Textbox", colorOrange, color.RGBA{255, 165, 0, 255}, 1000)      // Orange - textboxes
	dc.Register("Grave", colorPink, color.RGBA{255, 192, 203, 255}, 1000)        // Pink - graves
	dc.Register("Physics", colorWhite, color.RGBA{255, 255, 255, 255}, 5000)     // White - physics system
	dc.Register("EntityID", "\033[38;5;244m", color.RGBA{128, 128, 128, 255}, 0) // Gray - entity IDs
	dc.Register("Tiles", "\033[38;5;250m", color.RGBA{200, 200, 200, 255}, 0)    // Light gray - tiles
	dc.Register("Ladder", colorGreen, color.RGBA{0, 255, 0, 100}, 0)             // Green - ladders (semi-transparent)
	dc.Register("FakeWall", colorGreen, color.RGBA{0, 255, 0, 255}, 0)           // Green - fake wall collision debug

	return dc
}

// Register creates a new debug category with the given configuration.
func (dc *DebugCategories) Register(name, consoleColor string, visualColor color.RGBA, throttleMS int) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.categories[name] = &DebugCategory{
		Name:         name,
		ConsoleColor: consoleColor,
		VisualColor:  visualColor,
		ThrottleMS:   throttleMS,
	}
}

// Get returns a debug category by name.
func (dc *DebugCategories) Get(name string) (*DebugCategory, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	cat, ok := dc.categories[name]
	return cat, ok
}

// GetColor returns the visual color for a category.
func (dc *DebugCategories) GetColor(name string) (color.RGBA, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	if cat, ok := dc.categories[name]; ok {
		return cat.VisualColor, true
	}
	return color.RGBA{}, false
}

// GetConsoleColor returns the ANSI console color for a category.
func (dc *DebugCategories) GetConsoleColor(name string) (string, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	if cat, ok := dc.categories[name]; ok {
		return cat.ConsoleColor, true
	}
	return "", false
}

// ShouldLog checks if a category should log based on throttle settings.
// Returns true if logging is allowed, false if throttled.
// Also updates the last log time for the category.
func (dc *DebugCategories) ShouldLog(name string) bool {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	cat, ok := dc.categories[name]
	if !ok {
		return false
	}

	now := time.Now()
	if cat.ThrottleMS > 0 && !cat.LastLogTime.IsZero() {
		elapsed := now.Sub(cat.LastLogTime).Milliseconds()
		if elapsed < int64(cat.ThrottleMS) {
			return false
		}
	}

	cat.LastLogTime = now
	return true
}

// GetAll returns all registered category names.
func (dc *DebugCategories) GetAll() []string {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	names := make([]string, 0, len(dc.categories))
	for name := range dc.categories {
		names = append(names, name)
	}
	return names
}
