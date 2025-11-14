package prefabs

import (
	"math/rand"

	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/pkg/config"
)

const (
	// Visual properties
	acedianAnimFile                                   = "acedian"
	acedianWidth, acedianHeight                       = 10, 12
	acedianOffsetX, acedianOffsetY, acedianOffsetFlip = -1, -2, 6

	// Combat stats - Acedian is invulnerable (no Health)
	acedianMaxPoise = 100 // Maximum poise
	acedianPoise    = 100 // Starting poise

	// Light properties - Acedian emits a soft glow
	acedianLightRadius    = 60.0 // pixels
	acedianLightIntensity = 0.8  // slightly dimmer than torch
	acedianPulseSpeed     = 8.0  // subtle pulse
)

// NewAcedianPrefab creates an Acedian NPC entity with dialogue.
// Returns an ECS EntityId instead of core.Entity.
func NewAcedianPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
	if world == nil {
		return 0
	}

	eid := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions
	transform := &components.Transform{X: x, Y: y, W: acedianWidth, H: acedianHeight}
	world.AddComponent(eid, transform)

	// === ANIMATION COMPONENT ===
	// Animation component - created using Pure ECS helper
	anim := NewAnimationComponent(AnimationConfig{
		FilesName:      acedianAnimFile,
		OX:             acedianOffsetX,
		OY:             acedianOffsetY,
		OXFlip:         acedianOffsetFlip,
		Layer:          1,
		FSMInitial:     "Idle",
		FSMTransitions: make(map[string]string),
	})
	if anim == nil {
		world.DestroyEntity(eid)
		return 0
	}
	world.AddComponent(eid, anim)

	// === COMBAT COMPONENTS ===
	// Hitbox for receiving damage - initialized immediately with hurtbox from animation
	hitbox := &components.Hitbox{}
	addHurtboxToHitbox(anim, hitbox)
	world.AddComponent(eid, hitbox)

	// Poise - hit reactions and stagger (NO HEALTH - invulnerable)
	poise := &components.Poise{
		Current:        acedianPoise,
		Max:            acedianMaxPoise,
		RecoverSeconds: 3.0, // Recover poise after 3 seconds
	}
	world.AddComponent(eid, poise)

	// HeadHealthTimer - controls when poise bar is visible
	headHealthTimer := &components.HeadHealthTimer{}
	world.AddComponent(eid, headHealthTimer)

	// === DIALOGUE COMPONENT ===
	// Textbox component with default text and interaction area
	textboxData := &components.TextboxData{
		Text:      "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		Indicator: true, // Show position indicator
		Area: func() bump.Rect {
			return bump.NewRect(x-acedianWidth*2, y-acedianHeight, acedianWidth*4, acedianHeight*2)
		},
		AdvanceFlicker:      false,
		AdvanceFlickerTimer: 0,
		Active:              false,
	}
	// Store entity bounds for indicator positioning
	textboxData.EntityX = x
	textboxData.EntityY = y
	textboxData.EntityW = acedianWidth
	textboxData.EntityH = acedianHeight
	world.AddComponent(eid, textboxData)

	// === PHYSICS COMPONENT ===
	// Static physics - boss NPC never moves
	physics := spatial.NewPhysicsStatic()
	world.AddComponent(eid, physics)

	// === COLLISION COMPONENT ===
	// Solid collider - blocks player movement
	collider := &components.Collider{
		Tags:      []string{"body", "solid"},
		QueryTags: []string{},
		Solid:     true,
		Immovable: true, // Static boss
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0, // Use Transform size
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(eid, collider)

	// === FACING COMPONENT ===
	// Facing component for sprite orientation (used by hitbox init)
	facing := &components.Facing{FlipX: flipX}
	world.AddComponent(eid, facing)

	// === BEHAVIOR COMPONENT ===
	acedian := &components.Acedian{}
	world.AddComponent(eid, acedian)

	// === LIGHT EMITTER COMPONENT ===
	// Acedian emits a soft glow (like upstream implementation)
	cfg := ecs.Resource[config.Config](world)
	radius := acedianLightRadius
	intensity := acedianLightIntensity
	if cfg != nil {
		// Check if there's a configured light source for acedian
		if source := cfg.Lighting.GetLightSource("acedian"); source.Radius > 0 {
			radius = source.Radius
			intensity = source.Intensity
		}
	}
	lightEmitter := &components.LightEmitter{
		EntityID:   eid,
		Radius:     radius,
		Active:     true,
		Intensity:  intensity,
		PulseSpeed: acedianPulseSpeed,
		PulsePhase: rand.Float64() * 10.0, // Random phase for variety
	}
	world.AddComponent(eid, lightEmitter)

	return eid
}

// NOTE: SetAcedianText has been removed.
// Use ui.UpdateTextboxText(world, eid, text) instead.
