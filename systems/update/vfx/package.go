// Package vfx contains Phase 8 systems for visual effects and particles.
// This phase updates particle systems, flash effects, and other visual effects
// that don't affect gameplay logic.
//
// Systems in this phase:
//   - Smoke: Particle smoke effects (death clouds, ambient atmosphere)
//   - Flake: Death particles that home to player for XP collection
//   - DamageNumber: Floating damage numbers with fade animations
//   - Visual Effects: Flash effects on damage, screen shake, etc.
//
// Order: This phase runs after state updates, before UI rendering.
//   - Game state is finalized (health updated, deaths processed)
//   - VFX can spawn based on current frame state
//   - Draw systems will render these effects next frame
//
// Performance: ~0.5ms per frame (particle systems can be expensive with many particles)
package vfx

import (
	"game/ecs"
)

// Update runs all Phase 8 systems: particle effects and visual feedback.
//
// This is the entry point for the VFX phase. It updates all visual effects
// including particles (smoke, flakes, damage numbers), combat flash effects,
// and other purely visual systems that don't affect gameplay.
//
// Order within phase:
//  1. Smoke - Update smoke particle tweens
//  2. Flake - Update death loot particle animations and homing
//  3. DamageNumber - Update floating damage number animations
//  4. VisualEffects - Update combat flash timers
func Update(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	UpdateSmoke(world, world, dt)
	UpdateFlake(world, world, dt)
	UpdateDamageNumber(world, dt)
	UpdateVisualEffects(world, dt)
}
