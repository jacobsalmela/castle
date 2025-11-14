package prefabs

import (
	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/systems/update/preupdate"
)

const (
	// Visual properties (player uses knight sprite)
	playerAnimFile = "knight" // Animation file base name
	playerWidth    = 8        // Collision width in pixels
	playerHeight   = 11       // Collision height in pixels

	// Sprite offset configuration
	playerOffsetX    = -10 // Sprite offset when facing right
	playerOffsetY    = -3  // Vertical sprite offset
	playerOffsetFlip = 17  // Sprite offset when facing left

	// Physics properties
	playerWeight = 1.0 // Normal weight for knockback calculations

	// Player stats (different from enemies - player is stronger)
	playerMaxHealth  = 60 // Hit points
	playerMaxStamina = 65 // Action resource (attacks, dodges, jumps)
	playerMaxHeal    = 5  // Healing charges (limited resource)

	playerDefaultMaxX       = 55
	playerDefaultSpeed      = 350
	playerDefaultJumpSpeed  = 110
	playerDefaultClimbSpeed = 5
	playerDefaultDamage     = 20
	playerDefaultPoise      = 16
	playerDefaultJumpCost   = 30
)

// NewPlayerPrefab constructs a player entity.
//
// The player is FULLY MIGRATED to Pure ECS (Phase 2 complete - no Control component).
// All combat, movement, and state management is handled through dedicated systems.
//
// The player has a complex multi-system lifecycle:
//  1. Input: Keyboard/gamepad input processed each frame
//  2. Movement: Horizontal walk, jump, climb, dash mechanics
//  3. Combat: Light/heavy attacks, blocking, parrying, healing
//  4. Animation: State-driven sprite animation with attack holds
//  5. Stats: Health, stamina, poise, healing charges with HUD display
//
// Systems involved:
//   - systems/update/player.go: Main input, movement, guard state, and attack execution (Phase 8: consolidated)
//   - systems/update/player_hurt.go: Damage reaction (Hit events)
//   - systems/update/player_block.go: Block/parry reaction (Block/ParryBlock events)
//   - systems/update/player_combat_effects.go: Camera shake, freeze, visual effects
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//
// Returns: EntityId of the created player, or 0 if world is nil
func NewPlayerPrefab(world *ecs.World, x, y float64) entities.EntityId {
	if world == nil {
		return 0
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Create ECS entity
	eid := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Position and dimensions
	transform := &components.Transform{
		X: x,
		Y: y,
		W: playerWidth,
		H: playerHeight,
	}
	world.AddComponent(eid, transform)

	// === VISUAL COMPONENT ===
	// Animation with sprite offsets
	anim := NewAnimationComponent(AnimationConfig{
		FilesName:  playerAnimFile,
		OX:         playerOffsetX,
		OY:         playerOffsetY,
		OXFlip:     playerOffsetFlip,
		Layer:      5, // Player renders above most entities
		FSMInitial: "Idle",
		FSMTransitions: map[string]string{
			"ParryBlock": "Block", // ParryBlock (1 frame) transitions to Block (hold state)
			"Block":      "Block", // Block loops to itself (hold state while button held)
			"Climb":      "Climb", // Climb loops to itself (hold state while on ladder)
		},
	})
	if anim == nil {
		world.DestroyEntity(eid)
		return 0
	}
	world.AddComponent(eid, anim)

	// === COLLISION COMPONENT ===
	// Hitbox will be initialized in systems/update/player.go from animation slice data
	hitbox := &components.Hitbox{}
	world.AddComponent(eid, hitbox)

	// === STATS COMPONENTS (Pure ECS Pattern) ===
	// Health component
	health := &components.Health{
		Current: playerMaxHealth,
		Max:     playerMaxHealth,
		Lag:     playerMaxHealth,
	}
	world.AddComponent(eid, health)

	// Stamina component (for attacks, dodges, blocking)
	stamina := &components.Stamina{
		Current:      playerMaxStamina,
		Max:          playerMaxStamina,
		Lag:          playerMaxStamina,
		RecoveryRate: cfg.Stats.RecoverRate,
	}
	world.AddComponent(eid, stamina)

	// Poise component (knockback/stagger resistance)
	poise := &components.Poise{
		Current:        playerDefaultPoise,
		Max:            playerDefaultPoise,
		Lag:            playerDefaultPoise,
		RecoverSeconds: cfg.Stats.RecoverSeconds,
	}
	world.AddComponent(eid, poise)

	// Healing component (heal charges and amounts)
	healing := &components.Healing{
		Count:      playerMaxHeal,
		MaxCount:   playerMaxHeal,
		HealAmount: cfg.Stats.HealAmount,
	}
	world.AddComponent(eid, healing)

	// Experience component (not used yet, but consistent with enemies)
	experience := &components.Experience{Points: 0}
	world.AddComponent(eid, experience)

	// Attack multiplier (bonus damage from consuming heals)
	attackMult := &components.AttackMultiplier{
		Current: 0,
		PerHeal: 0.2, // Default: +20% damage per heal consumed
	}
	world.AddComponent(eid, attackMult)

	// === NEW PHYSICS COMPONENT (Phase 2) ===
	// Player physics with custom max velocity
	physics := spatial.NewPhysics()
	physics.Weight = playerWeight
	physics.MaxVelocity.X = playerDefaultMaxX
	physics.GravityEnabled = true
	physics.FrictionEnabled = true
	world.AddComponent(eid, physics)

	// === NEW COLLIDER COMPONENT (Phase 2) ===
	// Player collider (solid, player team)
	collider := &components.Collider{
		Tags:      []string{"player", "body"},
		QueryTags: []string{"enemy", "body", "map", "solid", "passthrough"},
		Solid:     true,
		Immovable: false,
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0, // Use Transform size
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(eid, collider)

	// === TEAM COMPONENT ===
	team := &components.Team{Type: components.TeamPlayer}
	world.AddComponent(eid, team)

	// === INPUT COMPONENT ===
	// Keyboard/gamepad input state (Pure ECS: pure data component)
	// Uses configurable key bindings from config file
	input := &components.Input{
		KeyBindings: preupdate.KeyBindingsFromWorld(world),
	}
	world.AddComponent(eid, input)

	// Action tracking for input buffering
	actions := &components.ActionIntents{}
	world.AddComponent(eid, actions)

	// === FACING COMPONENT ===
	// Sprite direction (controlled by input, not AI)
	facing := &components.Facing{FlipX: false}
	world.AddComponent(eid, facing)

	// === BEHAVIOR COMPONENT ===
	// Player-specific state (Pure ECS: only data fields)
	player := &components.Player{
		// Gameplay tuning
		Speed:           playerDefaultSpeed,
		JumpSpeed:       playerDefaultJumpSpeed,
		JumpCost:        playerDefaultJumpCost,
		ClimbSpeed:      playerDefaultClimbSpeed,
		AttackDamage:    playerDefaultDamage,
		ReactForce:      0,
		AttackPushForce: 0,
		AttackLevel:     0,
	}
	world.AddComponent(eid, player)

	return eid
}
