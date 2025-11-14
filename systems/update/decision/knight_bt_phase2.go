package decision

import (
	"game/components/ai"
	"game/pkg/config"
	"game/systems/update/decision/nodes"
)

// buildKnightPhase2Tree creates the behavior tree for Phase 2 knight (<80% health).
//
// Phase 2 Behavior Pattern:
//   1. Idle until ready
//   2. Approach target to minimum range
//   3. Brief pause before action
//   4. Randomly choose action including Phase 2 abilities:
//      - Attack: 38% chance
//      - Throw: 13% chance
//      - Backup: 10% chance
//      - Idle: 15% chance
//      - Dash: 19% chance (if cooldown ready)
//      - Shield: 17% chance
//   5. Repeat forever
//
// This creates an aggressive, unpredictable combat pattern with defensive options.
func buildKnightPhase2Tree(cfg *config.Config) ai.Node {
	// Pre-wrap actions with human-friendly names for debug visualization
	idle := ai.WrapForDebug(nodes.Idle(), "Idle", 2)
	approach := ai.WrapForDebug(newKnightApproach(), "Approach", 2)
	pauseBrief := ai.WrapForDebug(nodes.Wait(0.1), "PreAttackPause(0.1s)", 2)

	// Action choices (Phase 1 actions)
	attack := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newKnightAttack(), "Attack", 4),
		},
	}, "AttackSequence", 3)

	throwRock := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newKnightThrow(), "ThrowRock", 4),
			ai.WrapForDebug(nodes.Wait(knightThrowRecover), "ThrowRecover(0.4s)", 4),
		},
	}, "ThrowSequence", 3)

	backup := ai.WrapForDebug(&ai.Timeout{
		Duration: 1.0,
		Child:    newKnightBackup(),
	}, "Backup(1.0s)", 3)

	idlePause := ai.WrapForDebug(nodes.Wait(0.2), "IdlePause(0.2s)", 3)

	// Phase 2 exclusive: Dash with cooldown check
	dash := ai.WrapForDebug(&ai.Sequence{
		Children: []ai.Node{
			ai.WrapForDebug(newKnightDashReady(), "DashReady?", 4),
			ai.WrapForDebug(&ai.Sequence{
				Children: []ai.Node{
					ai.WrapForDebug(newKnightDash(cfg), "Dash", 5),
					ai.WrapForDebug(nodes.Wait(0.1), "DashRecover(0.1s)", 5),
				},
			}, "DashSequence", 4),
		},
	}, "ConditionalDash", 3)

	// Phase 2 exclusive: Shield defense
	shield := ai.WrapForDebug(newKnightShield(cfg), "Shield", 3)

	// Random action selector with weights for Phase 2
	randomAction := ai.WrapForDebug(nodes.NewRandomSelector(
		nodes.WeightedChoice{Weight: 2.0, Node: attack},    // ~38% chance
		nodes.WeightedChoice{Weight: 0.7, Node: throwRock}, // ~13% chance
		nodes.WeightedChoice{Weight: 0.5, Node: backup},    // ~10% chance
		nodes.WeightedChoice{Weight: 0.8, Node: idlePause}, // ~15% chance
		nodes.WeightedChoice{Weight: 1.0, Node: dash},      // ~19% chance (if cooldown ready)
		nodes.WeightedChoice{Weight: 0.9, Node: shield},    // ~17% chance
	), "ChooseAction", 2)

	// Assemble the tree
	tree := &ai.Repeat{
		Count: 0, // Infinite loop
		Child: &ai.Sequence{
			Children: []ai.Node{
				idle,
				approach,
				pauseBrief,
				randomAction,
			},
		},
	}

	return ai.WrapForDebug(tree, "KnightPhase2BT", 0)
}
