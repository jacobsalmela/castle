package enemies

import (
	"game/entities"
)

// Ghoul is an intelligent enemy that can approach, attack, jump-attack, and throw rocks.
// They support two AI modes: aggressive (melee) and poacher (ranged).
type Ghoul struct {
	RemovalTarget entities.EntityId // Target entity for removal queue
	Rocks         int               // Number of rocks available to throw
	Poacher       bool              // If true, uses ranged AI (throws rocks, backs away)
	ThrowCooldown float64           // Cooldown timer between rock throws
	Paused        bool              // Movement paused during attacks
}
