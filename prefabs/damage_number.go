package prefabs

import (
	"fmt"
	"math/rand"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/tween"
)

const (
	// Animation timing
	damageNumberDuration = 1.5 // Total animation duration in seconds

	// Movement configuration
	damageNumberMinDriftX = -20.0 // Minimum X drift (pixels)
	damageNumberMaxDriftX = 20.0  // Maximum X drift (pixels)
	damageNumberMinDriftY = -40.0 // Minimum Y drift (upward, pixels)
	damageNumberMaxDriftY = -20.0 // Maximum Y drift (upward, pixels)

	// Visual configuration
	damageNumberLayer = 100 // High layer to render above everything
)

// NewDamageNumberPrefab constructs a damage number VFX entity.
//
// Damage numbers are floating text particles that display combat damage:
//  1. Spawn centered on the damaged entity's sprite
//  2. Drift upward/outward in random directions
//  3. Fade out over 1.5 seconds
//  4. Use larger, colored text for critical hits (>50% max health)
//
// The number displays the actual damage dealt, with visual emphasis for high damage.
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Center position of the damaged entity's sprite
//   - damage: Damage amount to display
//   - maxHealth: Maximum health of the damaged entity (for critical detection)
//
// Returns: EntityId of the created damage number, or 0 if world is nil
func NewDamageNumberPrefab(world *ecs.World, x, y, damage, maxHealth float64) entities.EntityId {
	if world == nil {
		return 0
	}

	// Create entity
	entityID := world.NewEntity()

	// Calculate random drift trajectory (upward arc with horizontal variance)
	driftX := damageNumberMinDriftX + rand.Float64()*(damageNumberMaxDriftX-damageNumberMinDriftX)
	driftY := damageNumberMinDriftY + rand.Float64()*(damageNumberMaxDriftY-damageNumberMinDriftY)

	// Determine if this is a critical hit (damage >= 50% of max health)
	critical := damage >= (maxHealth * 0.5)

	// === SPATIAL COMPONENT ===
	// Position at center of damaged entity's sprite
	transform := &components.Transform{
		X: x,
		Y: y,
		W: 0, // Text has no collision
		H: 0,
	}
	world.AddComponent(entityID, transform)

	// === BEHAVIOR COMPONENT ===
	// Damage number animation state
	animTween := tween.New(0, 1, damageNumberDuration, tween.EaseOutCubic)
	damageNumber := &components.DamageNumber{
		Tween:    animTween,
		StartX:   x,
		StartY:   y,
		TargetX:  driftX,
		TargetY:  driftY,
		Damage:   damage,
		Critical: critical,
		EntityID: entityID,
	}
	world.AddComponent(entityID, damageNumber)

	// === VISUAL COMPONENT ===
	// Render as text (actual rendering handled by draw system)
	// Using high layer to ensure numbers appear above all sprites
	render := &components.Render{
		Layer: damageNumberLayer,
	}
	world.AddComponent(entityID, render)

	return entityID
}

// FormatDamageText returns the formatted damage text for display.
// Pure helper function for consistent formatting across systems.
func FormatDamageText(damage float64) string {
	return fmt.Sprintf("%.0f", damage)
}
