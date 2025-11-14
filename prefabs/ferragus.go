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
	ferragusAnimFile                                     = "ferragus"
	ferragusWidth, ferragusHeight                        = 8, 15
	ferragusOffsetX, ferragusOffsetY, ferragusOffsetFlip = -2, -1, 6

	// Stats
	ferragusMaxPoise = 100
	ferragusPoise    = 100

	// Interaction
	ferragusInteractionWidthMultiplier  = 4 // Textbox interaction area width multiplier
	ferragusInteractionHeightMultiplier = 2 // Textbox interaction area height multiplier
)

// NewFerragusPrefab constructs a Ferragus boss NPC entity.
//
// Ferragus is a passive, unmovable boss NPC that displays textbox dialogue.
// Unlike Oscar, Ferragus is completely invulnerable (no Health component) but
// can show hit reactions via Poise stagger.
//
// Lifecycle:
//  1. Spawn: All components initialized, textbox indicator visible
//  2. Idle: Waits for player interaction, displays dialogue
//  3. Hit: Can be staggered via Poise but takes no damage
//  4. Boss: Never dies, stays as solid obstacle
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: Currently unused (Ferragus uses fixed 8x15 dimensions)
//   - flipX: Initial sprite orientation
//
// Returns: EntityId of the created entity, or 0 if world is nil
//
// Use SetFerragusText() to configure dialogue text after creation.
func NewFerragusPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {

	eid := world.NewEntity()

	// Transform component
	transform := &components.Transform{X: x, Y: y, W: ferragusWidth, H: ferragusHeight}
	world.AddComponent(eid, transform)

	// === ANIMATION COMPONENT ===
	// Animation component - created using Pure ECS helper
	anim := NewAnimationComponent(AnimationConfig{
		FilesName:      ferragusAnimFile,
		OX:             ferragusOffsetX,
		OY:             ferragusOffsetY,
		OXFlip:         ferragusOffsetFlip,
		Layer:          1,
		FSMInitial:     "Idle",
		FSMTransitions: make(map[string]string),
	})
	if anim == nil {
		world.DestroyEntity(eid)
		return 0
	}
	world.AddComponent(eid, anim)

	// === HITBOX COMPONENT ===
	// Hitbox component - initialized immediately with hurtbox from animation
	hitbox := &components.Hitbox{}
	addHurtboxToHitbox(anim, hitbox)
	world.AddComponent(eid, hitbox)

	// === PURE ECS STATS COMPONENTS ===

	// Poise component - for hit reactions only (Ferragus is invulnerable)
	poise := &components.Poise{
		Current:        ferragusPoise,
		Max:            ferragusMaxPoise,
		RecoverSeconds: 3.0,
	}
	world.AddComponent(eid, poise)

	// HeadHealthTimer component - controls health bar visibility
	// Note: Ferragus has no Health (invulnerable), but needs timer for UI
	headHealthTimer := &components.HeadHealthTimer{}
	world.AddComponent(eid, headHealthTimer)

	// === LEGACY COMPONENTS (to be migrated) ===

	// Textbox component with default text and interaction area
	textboxData := &components.TextboxData{
		Text:      "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		Indicator: true, // Show position indicator
		Area: func() bump.Rect {
			return bump.NewRect(x-ferragusWidth*ferragusInteractionWidthMultiplier/2, y-ferragusHeight, ferragusWidth*ferragusInteractionWidthMultiplier, ferragusHeight*ferragusInteractionHeightMultiplier)
		},
		AdvanceFlicker:      false,
		AdvanceFlickerTimer: 0,
		Active:              false,
	}
	// Store entity bounds for indicator positioning
	textboxData.EntityX = x
	textboxData.EntityY = y
	textboxData.EntityW = ferragusWidth
	textboxData.EntityH = ferragusHeight
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
	// Facing component for sprite orientation
	facing := &components.Facing{FlipX: flipX}
	world.AddComponent(eid, facing)

	// === TEAM COMPONENT ===
	// Team component - boss enemy
	world.AddComponent(eid, &components.Team{Type: components.TeamEnemy})

	// === BEHAVIOR COMPONENT ===
	ferragus := &components.Ferragus{}
	world.AddComponent(eid, ferragus)

	return eid
}

// NOTE: SetFerragusText has been removed.
// Use ui.UpdateTextboxText(world, eid, text) instead.
