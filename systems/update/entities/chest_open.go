package entities

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/prefabs"
	"game/resources"
)

// ApplyChestOpen processes combat events to open chests when hit by the player.
// This system handles the complete chest opening sequence including animation
// stages and reward particle spawning.
func ApplyChestOpen(world *ecs.World, events []resources.HitEvent) {
	if world == nil || len(events) == 0 {
		return
	}

	for i := range events {
		event := &events[i]
		if event.Handled {
			continue
		}

		// Only process hit contacts (not overlap/proximity)
		if event.Contact != int(components.Hit) {
			continue
		}

		// Check if target is a chest
		chest := ecs.GetComponent[components.Chest](world, event.Target)
		if chest == nil {
			continue
		}

		// Skip if chest is already open
		if chest.Opened {
			continue
		}

		// Only player can open chests
		if !isPlayerAttacker(world, event.Attacker) {
			continue
		}

		// Trigger chest opening sequence
		openChest(world, event.Target, chest)
		event.Handled = true
	}
}

// UpdateChestAnimation advances chest opening animations over time.
// Handles transition between animation stages (closed → semi-open → fully open).
func UpdateChestAnimation(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	chests := world.EntitiesWith((*components.Chest)(nil), (*components.Animation)(nil))
	for _, eid := range chests {
		chest := ecs.GetComponent[components.Chest](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		if chest == nil || anim == nil || !chest.Opened {
			continue
		}

		// Advance animation timer
		chest.AnimationTimer += dt

		// Process animation stage transitions
		processChestAnimation(world, eid, chest, anim)
	}
}

// isPlayerAttacker checks if the attacker entity belongs to the player team.
func isPlayerAttacker(world *ecs.World, attackerID entities.EntityId) bool {
	if attackerID == 0 {
		return false
	}
	team := ecs.GetComponent[components.Team](world, attackerID)
	return team != nil && team.Type == components.TeamPlayer
}

// openChest initiates the chest opening sequence.
// Sets the chest to opened state and begins animation.
func openChest(world *ecs.World, chestID entities.EntityId, chest *components.Chest) {
	// Mark as opened (prevents re-opening)
	chest.Opened = true
	chest.AnimationStage = 1 // Start with animation stage 1
	chest.AnimationTimer = 0.0

	// Pure ECS: Remove Hitbox component so chest can't be hit again
	// The component removal ensures the chest is no longer detectable in hitbox queries
	if hitbox := ecs.GetComponent[components.Hitbox](world, chestID); hitbox != nil {
		ecs.RemoveComponent[components.Hitbox](world, chestID)
	}

	// Trigger the "activate" animation tag to play opening sequence
	anim := ecs.GetComponent[components.Animation](world, chestID)
	if anim != nil && anim.Data != nil {
		if err := anim.Data.Play("activate"); err == nil {
			anim.State = "activate"
			anim.Frame = 0
			anim.Timer = 0
			// Animation system will automatically transition to "open" state when "activate" finishes
			// (configured via FSMTransitions: "activate" -> "open")
		}
	}
}

// processChestAnimation handles animation stage transitions based on animation completion.
func processChestAnimation(world *ecs.World, chestID entities.EntityId, chest *components.Chest, anim *components.Animation) {
	switch chest.AnimationStage {
	case 1: // Opening animation playing
		// Check if animation has finished (reached last frame)
		if isAnimationComplete(anim) && chest.AnimationTimer >= prefabs.ChestFullOpenDelay {
			// Animation complete - spawn rewards
			chest.AnimationStage = 2
			spawnChestReward(world, chestID, chest.Reward)
		}
	case 2:
		// Fully open - animation complete, nothing more to do
	}
}

// isAnimationComplete checks if the animation has reached its last frame.
func isAnimationComplete(anim *components.Animation) bool {
	if anim == nil || anim.Data == nil || anim.Data.CurrentAnimation == nil {
		return false
	}

	// Get the frame range for current animation
	currentAnim := anim.Data.CurrentAnimation
	lastFrame := currentAnim.To - currentAnim.From

	// Check if we've reached or passed the last frame
	return anim.Frame >= lastFrame
}

// spawnChestReward creates flake particles at the chest's position.
func spawnChestReward(world *ecs.World, chestID entities.EntityId, rewardCount int) {
	// Get chest position for particle spawning
	transform := ecs.GetComponent[components.Transform](world, chestID)
	if transform == nil {
		return
	}

	// Get player entity for flake targeting
	players := world.EntitiesWith((*components.Player)(nil))
	if len(players) == 0 {
		return
	}
	playerID := players[0]

	// Spawn reward particles using flake system
	for range rewardCount {
		// Center the flake on the chest
		centerX := transform.X + transform.W/2
		centerY := transform.Y + transform.H/2

		// Create flake targeting player (Pure ECS: pass world explicitly)
		flakeID := prefabs.NewFlakePrefab(world, centerX, centerY, 0, playerID)
		if flakeID != 0 {
			world.QueueInit(flakeID)
		}
	}
}
