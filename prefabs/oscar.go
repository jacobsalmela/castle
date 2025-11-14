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
	oscarAnimFile                               = "oscar"
	oscarWidth, oscarHeight                     = 7, 12
	oscarOffsetX, oscarOffsetY, oscarOffsetFlip = -3, -1, 6

	// Stats
	oscarMaxHealth = 200
	oscarHealth    = 80 // Oscar starts injured
	oscarMaxPoise  = 100
	oscarPoise     = 100

	// Interaction
	oscarInteractionWidthMultiplier  = 4 // Textbox interaction area width multiplier
	oscarInteractionHeightMultiplier = 2 // Textbox interaction area height multiplier
)

// NewOscarPrefab constructs a Oscar NPC entity.
//
// Oscar is a passive, unmovable NPC that displays textbox dialogue and can take damage.
// He starts injured (80/200 HP) and displays a different message when killed.
//
// Lifecycle:
//  1. Spawn: All components initialized, textbox indicator visible
//  2. Idle: Waits for player interaction, displays dialogue
//  3. Damaged: Health bar appears, plays stagger animation
//  4. Death: Swaps to dead dialogue, hides indicator, stays as solid obstacle
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: Currently unused (Oscar uses fixed 7x12 dimensions)
//   - flipX: Initial sprite orientation
//
// Returns: EntityId of the created entity, or 0 if world is nil
//
// Use SetOscarText() to configure dialogue text after creation.
// Use SetOscarDeadText() to configure death dialogue after creation.
func NewOscarPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {

	eid := world.NewEntity()

	// Transform component
	transform := &components.Transform{X: x, Y: y, W: oscarWidth, H: oscarHeight}
	world.AddComponent(eid, transform)

	// === ANIMATION COMPONENT ===
	// Animation component - created using Pure ECS helper
	anim := NewAnimationComponent(AnimationConfig{
		FilesName:      oscarAnimFile,
		OX:             oscarOffsetX,
		OY:             oscarOffsetY,
		OXFlip:         oscarOffsetFlip,
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
	// Hitbox component - initialized in systems/update/oscar.go based on Facing
	// This ensures hitbox respects sprite flip direction (left vs right)
	hitbox := &components.Hitbox{}
	world.AddComponent(eid, hitbox)

	// === PURE ECS STATS COMPONENTS ===

	// Health component - Oscar starts injured (80/200)
	health := &components.Health{
		Current: oscarHealth,
		Max:     oscarMaxHealth,
	}
	world.AddComponent(eid, health)

	// Poise component - full poise
	poise := &components.Poise{
		Current:        oscarPoise,
		Max:            oscarMaxPoise,
		RecoverSeconds: 3.0,
	}
	world.AddComponent(eid, poise)

	// HeadHealthTimer component - controls health bar visibility
	headHealthTimer := &components.HeadHealthTimer{}
	world.AddComponent(eid, headHealthTimer)

	// === LEGACY COMPONENTS (to be migrated) ===

	// Textbox component with default text and interaction area
	textboxData := &components.TextboxData{
		Text:      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		Indicator: true, // Show position indicator
		Area: func() bump.Rect {
			return bump.NewRect(x-oscarWidth*2, y-oscarHeight, oscarWidth*4, oscarHeight*2)
		},
		AdvanceFlicker:      false,
		AdvanceFlickerTimer: 0,
		Active:              false,
	}
	// Store entity bounds for indicator positioning
	textboxData.EntityX = x
	textboxData.EntityY = y
	textboxData.EntityW = oscarWidth
	textboxData.EntityH = oscarHeight
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
	// Team component - boss enemy (takes damage)
	world.AddComponent(eid, &components.Team{Type: components.TeamEnemy})

	// === BEHAVIOR COMPONENT ===
	// Oscar-specific component (pure data)
	// HitboxInited is set to false - will be initialized by systems/update/oscar.go
	oscar := &components.Oscar{
		DeadText:     "", // Set via SetOscarDeadText()
		HitboxInited: false,
		DeathHandled: false,
	}
	world.AddComponent(eid, oscar)

	return eid
}

// NOTE: SetOscarText has been removed.
// Use ui.UpdateTextboxText(world, eid, text) instead.

// SetOscarDeadText assigns death dialogue to an Oscar entity.
func SetOscarDeadText(world *ecs.World, eid entities.EntityId, deadText string) {
	if oscar := ecs.GetComponent[components.Oscar](world, eid); oscar != nil {
		oscar.DeadText = deadText
	}
}
