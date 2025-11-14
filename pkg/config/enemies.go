package config

// EnemyBalance holds balance configuration for all enemy types.
// This enables data-driven tuning without code changes.
type EnemyBalance struct {
	// Per-enemy balance data
	Rat      EnemyStats `yaml:"rat" head_comment:"Rat enemy balance"`
	Bat      EnemyStats `yaml:"bat" head_comment:"Bat enemy balance"`
	Crawler  EnemyStats `yaml:"crawler" head_comment:"Crawler enemy balance"`
	Skeleman EnemyStats `yaml:"skeleman" head_comment:"Skeleman enemy balance"`
	Ghoul    EnemyStats `yaml:"ghoul" head_comment:"Ghoul enemy balance"`
	Ent      EnemyStats `yaml:"ent" head_comment:"Ent enemy balance"`
	Knight   EnemyStats `yaml:"knight" head_comment:"Knight boss balance"`
	Oscar    EnemyStats `yaml:"oscar" head_comment:"Oscar NPC balance"`
	Gram     EnemyStats `yaml:"gram" head_comment:"Gram NPC balance"`
	Ferragus EnemyStats `yaml:"ferragus" head_comment:"Ferragus NPC balance"`
}

// EnemyStats holds all configurable parameters for an enemy type.
// Mirrors prefabs.EnemyConfig but with YAML-friendly structure.
type EnemyStats struct {
	// === VISUAL PROPERTIES ===
	AnimFile   string  `yaml:"anim_file" line_comment:"Animation file name"`
	Width      float64 `yaml:"width" line_comment:"Collision width in pixels"`
	Height     float64 `yaml:"height" line_comment:"Collision height in pixels"`
	OffsetX    float64 `yaml:"offset_x" line_comment:"Sprite X offset (facing right)"`
	OffsetY    float64 `yaml:"offset_y" line_comment:"Sprite Y offset"`
	OffsetFlip float64 `yaml:"offset_flip" line_comment:"Sprite X offset (facing left)"`

	// === PHYSICS PROPERTIES ===
	Weight          float64 `yaml:"weight" line_comment:"Body weight for knockback"`
	GravityEnabled  bool    `yaml:"gravity" line_comment:"Apply gravity (false for flying)"`
	FrictionEnabled bool    `yaml:"friction" line_comment:"Apply friction"`
	MaxVelocityX    float64 `yaml:"max_velocity_x" line_comment:"Max horizontal speed"`
	MaxVelocityY    float64 `yaml:"max_velocity_y" line_comment:"Max vertical speed"`

	// === COMBAT STATS ===
	Health              int     `yaml:"health" line_comment:"Hit points"`
	Poise               int     `yaml:"poise" line_comment:"Knockback resistance"`
	Exp                 int     `yaml:"exp" line_comment:"Experience on death"`
	PoiseRecoverSeconds float64 `yaml:"poise_recover" line_comment:"Poise recovery time (0 = no recover)"`

	// === VISUAL EFFECTS ===
	FlashDuration float64      `yaml:"flash_duration" line_comment:"Hit flash duration"`
	FlashColor    [3]float32   `yaml:"flash_color" line_comment:"Flash color RGB [0-1]"`
	DieDuration   float64      `yaml:"die_duration" line_comment:"Death fade duration"`

	// === AI DETECTION ===
	DetectionFront float64 `yaml:"detection_front" line_comment:"Front detection range"`
	DetectionBack  float64 `yaml:"detection_back" line_comment:"Back detection range"`
	DetectionUp    float64 `yaml:"detection_up" line_comment:"Up detection range"`
	DetectionDown  float64 `yaml:"detection_down" line_comment:"Down detection range"`

	// === BEHAVIOR PARAMETERS ===
	ApproachSpeed    float64 `yaml:"approach_speed" line_comment:"Approach movement acceleration"`
	ApproachMaxSpeed float64 `yaml:"approach_max_speed" line_comment:"Approach max velocity"`
	ApproachMinRange float64 `yaml:"approach_min_range" line_comment:"Min distance to target"`
	BackupSpeed      float64 `yaml:"backup_speed" line_comment:"Backup movement speed"`
	BackupMaxRange   float64 `yaml:"backup_max_range" line_comment:"Max backup distance"`
}

// GetEnemyStats returns stats for a specific enemy type by name.
// Returns nil if the enemy type is not found.
func (e EnemyBalance) GetEnemyStats(name string) *EnemyStats {
	switch name {
	case "rat":
		return &e.Rat
	case "bat":
		return &e.Bat
	case "crawler":
		return &e.Crawler
	case "skeleman":
		return &e.Skeleman
	case "ghoul":
		return &e.Ghoul
	case "ent":
		return &e.Ent
	case "knight":
		return &e.Knight
	case "oscar":
		return &e.Oscar
	case "gram":
		return &e.Gram
	case "ferragus":
		return &e.Ferragus
	default:
		return nil
	}
}

// NewDefaultEnemyBalance returns default balance values for all enemies.
// These values match the current hardcoded constants in prefabs.
func NewDefaultEnemyBalance() EnemyBalance {
	return EnemyBalance{
		Rat: EnemyStats{
			// Visual
			AnimFile: "rat", Width: 11, Height: 7,
			OffsetX: -13, OffsetY: -17, OffsetFlip: 21,
			// Physics
			Weight: 0.6, GravityEnabled: true, FrictionEnabled: true,
			MaxVelocityX: 40.0, MaxVelocityY: 0,
			// Combat
			Health: 25, Poise: 10, Exp: 15, PoiseRecoverSeconds: 0,
			// Effects
			FlashDuration: 0.8, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection
			DetectionFront: 60.0, DetectionBack: 0.0, DetectionUp: 16.0, DetectionDown: 16.0,
			// Behavior (no approach/backup - uses BT)
			ApproachSpeed: 0, ApproachMaxSpeed: 0, ApproachMinRange: 0,
			BackupSpeed: 0, BackupMaxRange: 0,
		},
		Bat: EnemyStats{
			// Visual
			AnimFile: "bat", Width: 5, Height: 5,
			OffsetX: -10, OffsetY: -9, OffsetFlip: 9,
			// Physics
			Weight: 0.0, GravityEnabled: false, FrictionEnabled: false,
			MaxVelocityX: 40.0, MaxVelocityY: 40.0,
			// Combat
			Health: 45, Poise: 10, Exp: 15, PoiseRecoverSeconds: 0,
			// Effects
			FlashDuration: 1.5, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection
			DetectionFront: 7.5, DetectionBack: 7.5, DetectionUp: 0.0, DetectionDown: 80.0,
			// Behavior
			ApproachSpeed: 60.0, ApproachMaxSpeed: 40.0, ApproachMinRange: 20.0,
			BackupSpeed: 0, BackupMaxRange: 0,
		},
		Crawler: EnemyStats{
			// Visual
			AnimFile: "crawler", Width: 8, Height: 8,
			OffsetX: -8, OffsetY: -4, OffsetFlip: 12,
			// Physics
			Weight: 0.8, GravityEnabled: true, FrictionEnabled: true,
			MaxVelocityX: 35.0, MaxVelocityY: 0,
			// Combat
			Health: 35, Poise: 15, Exp: 20, PoiseRecoverSeconds: 0,
			// Effects
			FlashDuration: 0.8, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection
			DetectionFront: 50.0, DetectionBack: 10.0, DetectionUp: 20.0, DetectionDown: 20.0,
			// Behavior
			ApproachSpeed: 50.0, ApproachMaxSpeed: 35.0, ApproachMinRange: 15.0,
			BackupSpeed: 30.0, BackupMaxRange: 25.0,
		},
		Skeleman: EnemyStats{
			// Visual
			AnimFile: "skeleman", Width: 8, Height: 11,
			OffsetX: -8, OffsetY: -3, OffsetFlip: 12,
			// Physics
			Weight: 1.0, GravityEnabled: true, FrictionEnabled: true,
			MaxVelocityX: 45.0, MaxVelocityY: 0,
			// Combat
			Health: 50, Poise: 20, Exp: 25, PoiseRecoverSeconds: 0,
			// Effects
			FlashDuration: 0.8, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection
			DetectionFront: 60.0, DetectionBack: 15.0, DetectionUp: 24.0, DetectionDown: 24.0,
			// Behavior
			ApproachSpeed: 70.0, ApproachMaxSpeed: 45.0, ApproachMinRange: 20.0,
			BackupSpeed: 40.0, BackupMaxRange: 30.0,
		},
		Ghoul: EnemyStats{
			// Visual
			AnimFile: "ghoul", Width: 8, Height: 11,
			OffsetX: -8, OffsetY: -3, OffsetFlip: 12,
			// Physics
			Weight: 1.2, GravityEnabled: true, FrictionEnabled: true,
			MaxVelocityX: 50.0, MaxVelocityY: 0,
			// Combat
			Health: 65, Poise: 18, Exp: 30, PoiseRecoverSeconds: 0,
			// Effects
			FlashDuration: 0.8, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection
			DetectionFront: 70.0, DetectionBack: 20.0, DetectionUp: 24.0, DetectionDown: 24.0,
			// Behavior
			ApproachSpeed: 80.0, ApproachMaxSpeed: 50.0, ApproachMinRange: 18.0,
			BackupSpeed: 45.0, BackupMaxRange: 28.0,
		},
		Ent: EnemyStats{
			// Visual
			AnimFile: "ent", Width: 10, Height: 14,
			OffsetX: -7, OffsetY: -2, OffsetFlip: 14,
			// Physics
			Weight: 2.0, GravityEnabled: true, FrictionEnabled: true,
			MaxVelocityX: 30.0, MaxVelocityY: 0,
			// Combat
			Health: 80, Poise: 30, Exp: 35, PoiseRecoverSeconds: 0,
			// Effects
			FlashDuration: 1.0, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection
			DetectionFront: 80.0, DetectionBack: 30.0, DetectionUp: 32.0, DetectionDown: 32.0,
			// Behavior
			ApproachSpeed: 50.0, ApproachMaxSpeed: 30.0, ApproachMinRange: 25.0,
			BackupSpeed: 25.0, BackupMaxRange: 35.0,
		},
		Knight: EnemyStats{
			// Visual
			AnimFile: "knight", Width: 8, Height: 11,
			OffsetX: -10, OffsetY: -3, OffsetFlip: 17,
			// Physics
			Weight: 2.0, GravityEnabled: true, FrictionEnabled: true,
			MaxVelocityX: 60.0, MaxVelocityY: 0,
			// Combat
			Health: 180, Poise: 25, Exp: 50, PoiseRecoverSeconds: 3.0,
			// Effects
			FlashDuration: 1.5, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection
			DetectionFront: 24.0, DetectionBack: 24.0, DetectionUp: 24.0, DetectionDown: 24.0,
			// Behavior
			ApproachSpeed: 100.0, ApproachMaxSpeed: 60.0, ApproachMinRange: 25.0,
			BackupSpeed: 80.0, BackupMaxRange: 40.0,
		},
		Oscar: EnemyStats{
			// Visual
			AnimFile: "oscar", Width: 10, Height: 11,
			OffsetX: -3, OffsetY: -3, OffsetFlip: 6,
			// Physics
			Weight: 0.0, GravityEnabled: false, FrictionEnabled: false,
			MaxVelocityX: 0, MaxVelocityY: 0,
			// Combat (invulnerable NPC)
			Health: 0, Poise: 100, Exp: 0, PoiseRecoverSeconds: 3.0,
			// Effects
			FlashDuration: 0.8, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection (NPC - no detection)
			DetectionFront: 0, DetectionBack: 0, DetectionUp: 0, DetectionDown: 0,
			// Behavior (NPC - no movement)
			ApproachSpeed: 0, ApproachMaxSpeed: 0, ApproachMinRange: 0,
			BackupSpeed: 0, BackupMaxRange: 0,
		},
		Gram: EnemyStats{
			// Visual
			AnimFile: "gram", Width: 10, Height: 12,
			OffsetX: -1, OffsetY: -2, OffsetFlip: 6,
			// Physics
			Weight: 0.0, GravityEnabled: false, FrictionEnabled: false,
			MaxVelocityX: 0, MaxVelocityY: 0,
			// Combat (invulnerable NPC)
			Health: 0, Poise: 100, Exp: 0, PoiseRecoverSeconds: 3.0,
			// Effects
			FlashDuration: 0.8, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection (NPC - no detection)
			DetectionFront: 0, DetectionBack: 0, DetectionUp: 0, DetectionDown: 0,
			// Behavior (NPC - no movement)
			ApproachSpeed: 0, ApproachMaxSpeed: 0, ApproachMinRange: 0,
			BackupSpeed: 0, BackupMaxRange: 0,
		},
		Ferragus: EnemyStats{
			// Visual
			AnimFile: "ferragus", Width: 10, Height: 12,
			OffsetX: -1, OffsetY: -2, OffsetFlip: 6,
			// Physics
			Weight: 0.0, GravityEnabled: false, FrictionEnabled: false,
			MaxVelocityX: 0, MaxVelocityY: 0,
			// Combat (invulnerable NPC)
			Health: 0, Poise: 100, Exp: 0, PoiseRecoverSeconds: 3.0,
			// Effects
			FlashDuration: 0.8, FlashColor: [3]float32{1, 1, 1}, DieDuration: 1.0,
			// Detection (NPC - no detection)
			DetectionFront: 0, DetectionBack: 0, DetectionUp: 0, DetectionDown: 0,
			// Behavior (NPC - no movement)
			ApproachSpeed: 0, ApproachMaxSpeed: 0, ApproachMinRange: 0,
			BackupSpeed: 0, BackupMaxRange: 0,
		},
	}
}
