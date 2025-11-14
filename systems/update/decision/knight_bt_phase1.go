package decision

import (
	"game/components/ai"
	"game/pkg/config"
	"game/systems/update/decision/nodes"
)

// buildKnightPhase1Tree creates the behavior tree for Phase 1 knight (100%-80% health).
//
// Phase 1 Behavior Pattern:
//   1. Idle until ready
//   2. Approach target to minimum range
//   3. Brief pause before action
//   4. Randomly choose action (Attack 48%, Throw 17%, Backup 12%, Idle 19%)
//   5. Repeat forever
//
// This creates a defensive, methodical combat pattern.
func buildKnightPhase1Tree(cfg *config.Config) ai.Node {
	// Pre-wrap actions with human-friendly names for debug visualization
	idle := ai.WrapForDebug(nodes.Idle(), "Idle", 2)
	approach := ai.WrapForDebug(newKnightApproach(), "Approach", 2)
	pauseBrief := ai.WrapForDebug(nodes.Wait(0.1), "PreAttackPause(0.1s)", 2)

	// Action choices
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

	// Random action selector with weights matching legacy system
	randomAction := ai.WrapForDebug(nodes.NewRandomSelector(
		nodes.WeightedChoice{Weight: 2.0, Node: attack},    // ~48% chance
		nodes.WeightedChoice{Weight: 0.7, Node: throwRock}, // ~17% chance
		nodes.WeightedChoice{Weight: 0.5, Node: backup},    // ~12% chance
		nodes.WeightedChoice{Weight: 0.8, Node: idlePause}, // ~19% chance
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

	return ai.WrapForDebug(tree, "KnightPhase1BT", 0)
}
