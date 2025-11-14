package enemies

import (
	"game/entities"
)

// Bat is a flying enemy that stalks and attacks the player.
// Bats hover in the air, patrol areas, and dive-attack when targets are in range.
type Bat struct {
	RemovalTarget entities.EntityId // Entity to remove on death
	Paused        bool              // Whether AI/combat is paused
}
