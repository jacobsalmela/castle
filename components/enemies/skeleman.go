package enemies

import (
	"game/entities"
)

// Skeleman is a sword-wielding melee enemy with short and long attack patterns.
type Skeleman struct {
	RemovalTarget entities.EntityId // ECS entity to remove after death

	// State flags
	Paused bool // Whether AI/combat is paused (during attacks/reactions)
}
