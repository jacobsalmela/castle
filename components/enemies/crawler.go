package enemies

import (
	"game/entities"
)

// Crawler is a ground-based melee enemy that patrols and attacks.
// Crawlers walk back and forth in a patrol area, attacking when targets are in range.
type Crawler struct {
	RemovalTarget entities.EntityId // ECS entity to remove after death
	AiMode        string            // AI behavior mode ("patrol" or "wall" for wall-climbing)
	Paused        bool              // Whether AI/combat is paused
}
