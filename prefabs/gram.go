package prefabs

import (
	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
)

const (
	// Visual properties
	gramAnimFile                             = "gram"
	gramWidth, gramHeight                    = 10, 12
	gramOffsetX, gramOffsetY, gramOffsetFlip = -1, -2, 6

	// Combat stats - Gram is invulnerable (no Health)
	gramMaxPoise = 100 // Maximum poise
	gramPoise    = 100 // Starting poise
)

// NewGramPrefab creates a Gram NPC entity with dialogue.
// Returns an ECS EntityId instead of core.Entity.
//
// Gram is an invulnerable passive NPC (no Health component).
// - Takes poise damage for hit reactions
// - Cannot be killed
// - Shows poise bar when hit (via HeadHealthTimer)
// - Fully Pure ECS - no Control component
//
// Dimensions: 10x12 pixels
// Combat: Poise=100/100 (invulnerable, no Health)
// Special: Unmovable passive NPC with textbox dialogue
//
// Use SetGramText() to set dialogue text after creation.
func NewGramPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
	if world == nil {
		return 0
	}

	eid := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions
	transform := &components.Transform{X: x, Y: y, W: gramWidth, H: gramHeight}
	world.AddComponent(eid, transform)

	// === ANIMATION COMPONENT ===
	// Animation component - created using Pure ECS helper
	anim := NewAnimationComponent(AnimationConfig{
		FilesName:      gramAnimFile,
		OX:             gramOffsetX,
		OY:             gramOffsetY,
		OXFlip:         gramOffsetFlip,
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
		Current:        gramPoise,
		Max:            gramMaxPoise,
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
			return bump.NewRect(x-gramWidth*2, y-gramHeight, gramWidth*4, gramHeight*2)
		},
		AdvanceFlicker:      false,
		AdvanceFlickerTimer: 0,
		Active:              false,
	}
	// Store entity bounds for indicator positioning
	textboxData.EntityX = x
	textboxData.EntityY = y
	textboxData.EntityW = gramWidth
	textboxData.EntityH = gramHeight
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
	gram := &components.Gram{}
	world.AddComponent(eid, gram)

	return eid
}

// NOTE: SetGramText has been removed.
// Use ui.UpdateTextboxText(world, eid, text) instead.
