package enemies

import (
	"game/entities"
)

// Ent is a large ground enemy with powerful attacks.
// Ents have high health, deal significant damage, and use attack cooldowns.
type Ent struct {
	RemovalTarget  entities.EntityId // Entity to remove on death
	Paused         bool              // Whether AI is paused
	AttackCooldown float64           // Time remaining before next attack allowed
}
