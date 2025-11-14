package enemies

import (
	"game/entities"
)

// Rat is a small, fast enemy with patrol and chase behavior.
type Rat struct {
	RemovalTarget entities.EntityId // Entity to remove on death
	Paused        bool              // Whether AI is paused
}
