package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

const (
	ghoulDieDuration      = 1.0
	ghoulFlashDuration    = 0.05
	ghoulApproachMinRange = 20.0 // Minimum target range for approach behavior
)

func UpdateGhoul(world *ecs.World, dt float64) {
	entities := world.EntitiesWith((*components.Ghoul)(nil))
	for _, eid := range entities {
		ghoul := ecs.GetComponent[components.Ghoul](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		facing := ecs.GetComponent[components.Facing](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		ai := ecs.GetComponent[components.AI](world, eid)
		visualEffects := ecs.GetComponent[components.VisualEffects](world, eid)
		deathState := ecs.GetComponent[components.DeathState](world, eid)

		// Initialize ghoul-specific kinematics (one-time setup)
		ghoulInitKinematics(world, eid)

		// Initialize AI behavior tree when target is first detected
		if ai != nil && ai.TargetID != 0 && ai.BehaviorTree == nil {
			ai.BehaviorTree = buildGhoulBehaviorTree(world, eid)
		}

		updateGhoul(world, eid, ghoul, health, facing, anim, ai, visualEffects, deathState, dt)
	}
}

func ghoulInitKinematics(world *ecs.World, eid entities.EntityId) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}

	physics.MaxVelocity.X = ghoulMaxSpeed // Imported from ghoul_actions.go
}

// updateGhoul handles per-frame updates.
func updateGhoul(world *ecs.World, eid entities.EntityId, ghoul *components.Ghoul, health *components.Health, facing *components.Facing, anim *components.Animation, ai *components.AI, visualEffects *components.VisualEffects, deathState *components.DeathState, dt float64) {
	// Use generic enemy update (handles death, flash, AI, animations)
	// Note: Passing nil for state since ghoul doesn't use QueueRemoval
	if UpdateGenericEnemy(world, nil, eid, &ghoul.Paused, ghoul.RemovalTarget, dt) {
		return // Enemy is dead
	}

	// Ghoul-specific logic: Throw cooldown
	ghoulUpdateThrowCooldown(ghoul, dt)
}

func ghoulUpdateThrowCooldown(ghoul *components.Ghoul, dt float64) {
	if ghoul == nil || ghoul.ThrowCooldown <= 0 {
		return
	}
	ghoul.ThrowCooldown -= dt
	if ghoul.ThrowCooldown < 0 {
		ghoul.ThrowCooldown = 0
	}
}
