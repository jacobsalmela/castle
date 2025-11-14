package world

import (
	"game/entities"

	"github.com/hajimehoshi/ebiten/v2"
)

// Object represents a destructible environmental object (crate, barrel, pot, etc).
// Pure data - no methods. All logic in systems/update/world/object.go.
type Object struct {
	// TileImage and NormalImage are pre-constructed from tilemap data
	TileImage   *ebiten.Image
	NormalImage *ebiten.Image

	// Offset from Transform position (for multi-tile objects)
	OffsetX float64
	OffsetY float64

	// Reward drops when destroyed
	Reward int

	// Entities that spawned from this object (for cleanup)
	SpawnedEntities []entities.EntityId
}
