package ui

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/resources"
)

// RenderPlayerStatsHUD renders the player's stats HUD using decomposed stats components.
func RenderPlayerStatsHUD(world *ecs.World, queue *resources.RenderQueue, playerID entities.EntityId) {
	if world == nil || queue == nil || playerID == 0 {
		return
	}

	// Build HUDData from decomposed components
	hudData := BuildHUDDataFromComponents(world, playerID)
	if hudData == nil {
		return
	}

	// Use existing HUD rendering system
	RenderHUD(world, queue, hudData)
}

// RenderEntityHeadbar renders an entity's headbar using decomposed stats components.
func RenderEntityHeadbar(world *ecs.World, queue *resources.RenderQueue, entityID entities.EntityId, camX, camY float64) {
	if world == nil || queue == nil || entityID == 0 {
		return
	}

	// Build HeadbarData from decomposed components
	headbarData := BuildHeadbarDataFromComponents(world, entityID)
	if headbarData == nil {
		return
	}

	// Check if headbar should be visible
	// Show when: ShowTimer > 0 && Health < MaxHealth && Health > 0
	if !(headbarData.ShowTimer > 0 && headbarData.Health < headbarData.MaxHealth && headbarData.Health > 0) {
		return
	}

	// Get transform for positioning
	transform := ecs.GetComponent[components.Transform](world, entityID)
	if transform == nil {
		return
	}

	// Convert to camera-relative coordinates
	relativeX := transform.X - camX
	relativeY := transform.Y - camY

	// Render using position-based headbar rendering (no component required)
	RenderHeadbarAtPosition(world, queue, headbarData, relativeX, relativeY)
}

// BuildHUDDataFromComponents constructs a HUDData struct from decomposed stats components.
// This allows the existing RenderHUD function to work with the new component architecture.
func BuildHUDDataFromComponents(world *ecs.World, entityID entities.EntityId) *components.HUDData {
	if world == nil || entityID == 0 {
		return nil
	}

	health := ecs.GetComponent[components.Health](world, entityID)
	stamina := ecs.GetComponent[components.Stamina](world, entityID)
	healing := ecs.GetComponent[components.Healing](world, entityID)
	exp := ecs.GetComponent[components.Experience](world, entityID)
	attackMod := ecs.GetComponent[components.AttackModifier](world, entityID)
	attackMult := ecs.GetComponent[components.AttackMultiplier](world, entityID)

	// Require at least health to render
	if health == nil {
		return nil
	}

	hudData := &components.HUDData{
		Health:    health.Current,
		MaxHealth: health.Max,
		HealthLag: health.Lag,
	}

	// Optional components
	if stamina != nil {
		hudData.Stamina = stamina.Current
		hudData.MaxStamina = stamina.Max
		hudData.StaminaLag = stamina.Lag
	}

	if healing != nil {
		hudData.Heal = healing.Count
	}

	if exp != nil {
		hudData.Exp = exp.Points
	}

	// Support both AttackModifier and AttackMultiplier (player-specific)
	if attackMod != nil {
		hudData.AttackMult = attackMod.Multiplier
		// Only show attack mult if above minimum threshold (inline logic from IsActive)
		hudData.ShowAttackMult = attackMod.Multiplier >= 0.1
	} else if attackMult != nil {
		hudData.AttackMult = attackMult.Current
		// Only show attack mult if above minimum threshold
		hudData.ShowAttackMult = attackMult.Current >= 0.1
	}

	return hudData
}

// BuildHeadbarDataFromComponents constructs a HeadbarData struct from decomposed components.
// This allows the existing RenderHeadbar function to work with the new component architecture.
//
// For invulnerable entities (no Health component), uses Poise as a substitute to show hit reactions.
func BuildHeadbarDataFromComponents(world *ecs.World, entityID entities.EntityId) *components.HeadbarData {
	if world == nil || entityID == 0 {
		return nil
	}

	health := ecs.GetComponent[components.Health](world, entityID)
	poise := ecs.GetComponent[components.Poise](world, entityID)
	transform := ecs.GetComponent[components.Transform](world, entityID)
	timer := ecs.GetComponent[components.HeadHealthTimer](world, entityID)

	// Require transform
	if transform == nil {
		return nil
	}

	// Get show timer from component if available
	showTimer := 0.0
	if timer != nil {
		showTimer = timer.Timer
	}

	// If entity has Health, use it for the bar
	if health != nil {
		return &components.HeadbarData{
			Health:      health.Current,
			MaxHealth:   health.Max,
			HealthLag:   health.Lag,
			ShowTimer:   showTimer,
			EntityWidth: transform.W,
		}
	}

	// For invulnerable entities (no Health), use Poise as substitute
	// This allows showing hit reactions for NPCs like Ferragus
	if poise != nil {
		return &components.HeadbarData{
			Health:      poise.Current,
			MaxHealth:   poise.Max,
			HealthLag:   0, // Poise doesn't have lag
			ShowTimer:   showTimer,
			EntityWidth: transform.W,
		}
	}

	// No health or poise - can't render a bar
	return nil
}

// UpdateHeadbarDataFromHealth synchronizes a HeadbarData component with a Health component.
// This is used to update existing HeadbarData when health changes.
func UpdateHeadbarDataFromHealth(headbar *components.HeadbarData, health *components.Health, timer *components.HeadHealthTimer) {
	if headbar == nil || health == nil {
		return
	}

	// Update health values
	headbar.Health = health.Current
	headbar.MaxHealth = health.Max
	headbar.HealthLag = health.Lag

	// Update show timer from component if available
	if timer != nil {
		headbar.ShowTimer = timer.Timer
	}
}

// CreateHeadbarDataFromHealth creates a new HeadbarData component from Health and Transform.
// This is used during entity initialization to set up headbars.
func CreateHeadbarDataFromHealth(health *components.Health, transform *components.Transform) *components.HeadbarData {
	if health == nil || transform == nil {
		return nil
	}

	return &components.HeadbarData{
		Health:      health.Current,
		MaxHealth:   health.Max,
		HealthLag:   health.Lag,
		ShowTimer:   0, // Don't show initially
		EntityWidth: transform.W,
	}
}
