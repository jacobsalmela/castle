package decision

import (
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
)

// buildKnightBehaviorTree creates the behavior tree for knight enemies.
//
// Knights have two phases:
//   - Phase 1 (100%-80% HP): Defensive, uses Attack, Throw, Backup, Idle
//   - Phase 2 (<80% HP): Aggressive, adds Dash and Shield abilities
//
// The behavior tree is built based on the knight's current phase state.
func buildKnightBehaviorTree(world *ecs.World, eid entities.EntityId, cfg *config.Config) ai.Node {
	knight := ecs.GetComponent[components.Knight](world, eid)
	if knight == nil {
		return buildKnightPhase1Tree(cfg)
	}

	// Select tree based on phase
	if knight.SecondPhase {
		return buildKnightPhase2Tree(cfg)
	}
	return buildKnightPhase1Tree(cfg)
}
