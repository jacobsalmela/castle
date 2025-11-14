package game

import (
	"game/pkg/tilemap"
)

// Reset creates a fresh ECS world and spawns all entities.
// This is called by the game when restarting (e.g., after player death).
// It preserves the current map but rebuilds all entities with fresh save data.
func (g *Game) Reset() {
	g.ResetWithMap(nil)
}

// ResetWithMap creates a fresh ECS world with the specified map.
// If worldMap is nil, it attempts to preserve the existing map or loads a default.
//
// This delegates to world.go for all initialization logic, keeping Reset as a thin wrapper.
func (g *Game) ResetWithMap(worldMap *tilemap.Map) {
	// Use world initialization with map preservation
	// IMPORTANT: Pass existing viewport to preserve dimensions (avoids 0-size panic)
	// Config is nil - will be extracted from existing world
	world, viewport := InitializeWorldWithMap(nil, worldMap, g.saveData, g.world, g.viewport)

	// Update game state with new world
	g.world = world
	g.viewport = viewport

	// Apply saved game data to the newly created world
	// This restores stats and opened state for all entities
	g.ApplySaveData(g.saveData)
}
