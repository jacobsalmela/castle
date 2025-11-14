package prefabs

import (
	"game/components"
	"game/components/spatial"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"reflect"
)

// EnemyConfigFromStats creates an EnemyConfig from config.EnemyStats.
// This converts the YAML-friendly config structure to the prefab format.
func EnemyConfigFromStats(stats *config.EnemyStats) EnemyConfig {
	if stats == nil {
		return EnemyConfig{}
	}

	cfg := EnemyConfig{
		// Animation
		AnimFile:   stats.AnimFile,
		OffsetX:    stats.OffsetX,
		OffsetY:    stats.OffsetY,
		OffsetFlip: stats.OffsetFlip,

		// Dimensions
		Width:  stats.Width,
		Height: stats.Height,

		// Physics
		Weight:          stats.Weight,
		GravityEnabled:  stats.GravityEnabled,
		FrictionEnabled: stats.FrictionEnabled,
		MaxVelocityX:    stats.MaxVelocityX,
		MaxVelocityY:    stats.MaxVelocityY,

		// Combat
		Health:              int(stats.Health),
		Poise:               int(stats.Poise),
		Exp:                 int(stats.Exp),
		PoiseRecoverSeconds: stats.PoiseRecoverSeconds,

		// Effects
		FlashDuration: stats.FlashDuration,
		FlashColor:    stats.FlashColor,
		DieDuration:   stats.DieDuration,

		// Detection
		DetectionFront: stats.DetectionFront,
		DetectionBack:  stats.DetectionBack,
		DetectionUp:    stats.DetectionUp,
		DetectionDown:  stats.DetectionDown,
	}

	// Add behavior components if configured
	if stats.ApproachSpeed > 0 || stats.ApproachMaxSpeed > 0 {
		cfg.ApproachBehavior = &components.ApproachBehavior{
			Speed:           stats.ApproachSpeed,
			MaxSpeed:        stats.ApproachMaxSpeed,
			MinRange:        stats.ApproachMinRange,
			RangeAdjustment: 0.0,
		}
	}

	if stats.BackupSpeed > 0 || stats.BackupMaxRange > 0 {
		cfg.BackupBehavior = &components.BackupBehavior{
			Speed:    stats.BackupSpeed,
			MaxRange: stats.BackupMaxRange,
		}
	}

	return cfg
}

// GetEnemyConfig retrieves enemy configuration from the ECS world.
// Falls back to defaults if world or config is nil.
func GetEnemyConfig(world *ecs.World, enemyName string) EnemyConfig {
	if world == nil {
		return EnemyConfigFromStats(config.NewDefaultEnemyBalance().GetEnemyStats(enemyName))
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		return EnemyConfigFromStats(config.NewDefaultEnemyBalance().GetEnemyStats(enemyName))
	}

	stats := cfg.EnemyBalance.GetEnemyStats(enemyName)
	if stats == nil {
		return EnemyConfig{}
	}

	return EnemyConfigFromStats(stats)
}

// EnemyConfig defines all configurable parameters for enemy prefabs.
// This eliminates code duplication across 14+ enemy types by capturing
// the varying constant values in a single config struct.
type EnemyConfig struct {
	// === ANIMATION SETTINGS ===
	AnimFile   string  // Animation file base name (e.g., "rat", "knight")
	OffsetX    float64 // Sprite offset when facing right
	OffsetY    float64 // Vertical sprite offset
	OffsetFlip float64 // Sprite offset when facing left

	// === ANIMATION BEHAVIOR (optional) ===
	// If specified, overrides default animation FSM settings
	AnimLayer          int               // Animation layer (default: 0)
	AnimFSMInitial     string            // Initial FSM state (default: "")
	AnimFSMTransitions map[string]string // FSM state transitions (default: nil)

	// === SPATIAL DIMENSIONS ===
	Width  float64 // Collision width in pixels
	Height float64 // Collision height in pixels

	// === PHYSICS PROPERTIES ===
	Weight          float64 // Body weight for knockback calculations
	GravityEnabled  bool    // Apply gravity (true for ground enemies, false for flying)
	FrictionEnabled bool    // Apply friction (typically true for ground enemies)
	MaxVelocityX    float64 // Max horizontal velocity (0 = use config default)
	MaxVelocityY    float64 // Max vertical velocity (0 = use config default)

	// === COMBAT STATS ===
	Health              int     // Hit points
	Poise               int     // Knockback resistance
	Exp                 int     // Experience points on death
	PoiseRecoverSeconds float64 // Poise recovery time (0 = no recovery, knight uses 3.0)

	// === VISUAL EFFECTS ===
	FlashDuration float64    // Flash effect duration in seconds
	FlashColor    [3]float32 // Flash color RGB (e.g., [1,1,1] for white, [222,0,0] for red)
	DieDuration   float64    // Death fade duration in seconds

	// === AI DETECTION ===
	DetectionFront float64 // Front detection distance
	DetectionBack  float64 // Back detection distance
	DetectionUp    float64 // Up detection distance
	DetectionDown  float64 // Down detection distance

	// === OPTIONAL BEHAVIOR COMPONENTS ===
	// Nil = don't add component
	ApproachBehavior    *components.ApproachBehavior
	BackupBehavior      *components.BackupBehavior
	MeleeAttackBehavior *components.MeleeAttackBehavior
}

// NewEnemyPrefab constructs a generic enemy entity from configuration.
// This factory eliminates ~1000 lines of duplicate code across enemy types.
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - flipX: Initial sprite facing direction (true = left, false = right)
//   - config: Enemy configuration (stats, behavior, visuals)
//   - behaviorComponent: Enemy-specific component (e.g., &components.Rat{}, &components.Knight{})
//
// Returns: EntityId of the created enemy, or 0 if world is nil
func NewEnemyPrefab(
	world *ecs.World,
	x, y float64,
	flipX bool,
	config EnemyConfig,
	behaviorComponent interface{},
) entities.EntityId {
	if world == nil {
		return 0
	}

	// Create ECS entity
	eid := world.NewEntity()

	// === SPATIAL COMPONENT ===
	transform := &components.Transform{
		X: x,
		Y: y,
		W: config.Width,
		H: config.Height,
	}
	world.AddComponent(eid, transform)

	// === VISUAL COMPONENT ===
	animConfig := AnimationConfig{
		FilesName: config.AnimFile,
		OX:        config.OffsetX,
		OY:        config.OffsetY,
		OXFlip:    config.OffsetFlip,
	}

	// Optional animation settings
	if config.AnimLayer != 0 {
		animConfig.Layer = config.AnimLayer
	}
	if config.AnimFSMInitial != "" {
		animConfig.FSMInitial = config.AnimFSMInitial
	}
	if config.AnimFSMTransitions != nil {
		animConfig.FSMTransitions = config.AnimFSMTransitions
	}

	anim := NewAnimationComponent(animConfig)
	world.AddComponent(eid, anim)

	// === COMBAT COMPONENT ===
	hitbox := &components.Hitbox{}
	addHurtboxToHitbox(anim, hitbox)
	world.AddComponent(eid, hitbox)

	// === STAT COMPONENTS ===
	health := &components.Health{
		Current: float64(config.Health),
		Max:     float64(config.Health),
		Lag:     float64(config.Health),
	}
	world.AddComponent(eid, health)

	poise := &components.Poise{
		Current: float64(config.Poise),
		Max:     float64(config.Poise),
		Lag:     float64(config.Poise),
	}

	// Optional poise recovery (used by knight)
	if config.PoiseRecoverSeconds > 0 {
		poise.RecoverSeconds = config.PoiseRecoverSeconds
	}

	world.AddComponent(eid, poise)

	exp := &components.Experience{Points: config.Exp}
	world.AddComponent(eid, exp)

	headHealthTimer := &components.HeadHealthTimer{}
	world.AddComponent(eid, headHealthTimer)

	// === AI COMPONENT ===
	ai := &components.AI{TargetID: 0}
	world.AddComponent(eid, ai)

	detectionRange := &components.DetectionRange{
		FrontDistance: config.DetectionFront,
		BackDistance:  config.DetectionBack,
		UpDistance:    config.DetectionUp,
		DownDistance:  config.DetectionDown,
		TeamFilter:    "player",
	}
	world.AddComponent(eid, detectionRange)

	// === OPTIONAL MOVEMENT BEHAVIOR COMPONENTS ===
	if config.ApproachBehavior != nil {
		world.AddComponent(eid, *config.ApproachBehavior)
	}

	if config.BackupBehavior != nil {
		world.AddComponent(eid, config.BackupBehavior)
	}

	if config.MeleeAttackBehavior != nil {
		world.AddComponent(eid, config.MeleeAttackBehavior)
	}

	// === PHYSICS COMPONENT ===
	physics := spatial.NewPhysics()
	physics.Weight = config.Weight
	physics.GravityEnabled = config.GravityEnabled
	physics.FrictionEnabled = config.FrictionEnabled

	// Optional max velocity overrides
	if config.MaxVelocityX > 0 {
		physics.MaxVelocity.X = config.MaxVelocityX
	}
	if config.MaxVelocityY > 0 {
		physics.MaxVelocity.Y = config.MaxVelocityY
	}

	world.AddComponent(eid, physics)

	// === COLLIDER COMPONENT ===
	collider := &components.Collider{
		Tags:      []string{"enemy", "body"},
		QueryTags: []string{"player", "body", "map", "solid"},
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
	world.AddComponent(eid, &components.Team{Type: components.TeamEnemy})

	// === FACING COMPONENT ===
	facing := &components.Facing{FlipX: flipX}
	world.AddComponent(eid, facing)

	// === VISUAL EFFECTS COMPONENTS ===
	visualEffects := &components.VisualEffects{
		FlashTimer:    0,
		FlashDuration: config.FlashDuration,
		FlashColor:    config.FlashColor,
	}
	world.AddComponent(eid, visualEffects)

	deathState := &components.DeathState{
		DieTimer:    config.DieDuration,
		DieDuration: config.DieDuration,
	}
	world.AddComponent(eid, deathState)

	// === BEHAVIOR-SPECIFIC COMPONENT ===
	// Add enemy-specific component (Rat, Bat, Knight, etc.)
	if behaviorComponent != nil {
		// Set RemovalTarget field if it exists (common pattern in enemy components)
		setRemovalTarget(behaviorComponent, eid)
		world.AddComponent(eid, behaviorComponent)
	}

	return eid
}

// setRemovalTarget sets the RemovalTarget field on a component if it exists.
// Many enemy behavior components (Rat, Crawler, Skeleman, etc.) have this field
// which needs to be set to the entity's ID for proper cleanup.
func setRemovalTarget(component interface{}, eid entities.EntityId) {
	if component == nil {
		return
	}

	v := reflect.ValueOf(component)
	// Dereference pointer if needed
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Only process structs
	if v.Kind() != reflect.Struct {
		return
	}

	// Try to find and set RemovalTarget field
	field := v.FieldByName("RemovalTarget")
	if field.IsValid() && field.CanSet() {
		// Check if field is the correct type (EntityId)
		if field.Type() == reflect.TypeOf(eid) {
			field.Set(reflect.ValueOf(eid))
		}
	}
}
