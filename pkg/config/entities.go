package config

// Entities holds entity binding configuration for Tiled GID mappings.
type Entities struct {
	// Player and environment
	Player uint32 `yaml:"player" line_comment:"Player entity GID from Tiled tileset"`
	Torch  uint32 `yaml:"torch" line_comment:"Torch/light source GID from Tiled tileset"`

	// Enemies
	Knight   uint32 `yaml:"knight" line_comment:"Knight entity GID from Tiled tileset"`
	Ghoul    uint32 `yaml:"ghoul" line_comment:"Ghoul entity GID from Tiled tileset"`
	Skeleman uint32 `yaml:"skeleman" line_comment:"Skeleman entity GID from Tiled tileset"`
	Crawler  uint32 `yaml:"crawler" line_comment:"Crawler entity GID from Tiled tileset"`
	Rat      uint32 `yaml:"rat" line_comment:"Rat entity GID from Tiled tileset"`
	Bat      uint32 `yaml:"bat" line_comment:"Bat entity GID from Tiled tileset"`
	Ent      uint32 `yaml:"ent" line_comment:"Ent entity GID from Tiled tileset"`

	// Bosses
	Gram     uint32 `yaml:"gram" line_comment:"Gram boss GID from Tiled tileset"`
	Ferragus uint32 `yaml:"ferragus" line_comment:"Ferragus boss GID from Tiled tileset"`
	Oscar    uint32 `yaml:"oscar" line_comment:"Oscar boss GID from Tiled tileset"`
	Acedian  uint32 `yaml:"acedian" line_comment:"Acedian boss GID from Tiled tileset"`

	// Interactive objects
	Chest    uint32 `yaml:"chest" line_comment:"Chest object GID from Tiled tileset"`
	Grave    uint32 `yaml:"grave" line_comment:"Grave object GID from Tiled tileset"`
	Door     uint32 `yaml:"door" line_comment:"Door object GID from Tiled tileset"`
	Spike    uint32 `yaml:"spike" line_comment:"Spike trap GID from Tiled tileset"`
	FakeWall uint32 `yaml:"fake_wall" line_comment:"Fake wall object GID from Tiled tileset"`
	Block    uint32 `yaml:"block" line_comment:"Block object GID from Tiled tileset"`
}

// Default entity GID values (matching current hardcoded constants)
const (
	defaultPlayer = 25
	defaultTorch  = 378

	defaultKnight   = 26
	defaultGhoul    = 27
	defaultSkeleman = 28
	defaultCrawler  = 29
	defaultRat      = 30
	defaultBat      = 31
	defaultEnt      = 32

	defaultGram     = 87
	defaultFerragus = 88
	defaultOscar    = 89
	defaultAcedian  = 90

	defaultChest    = 149
	defaultGrave    = 150
	defaultDoor     = 151
	defaultSpike    = 152
	defaultFakeWall = 153
	defaultBlock    = 154
)

// NewDefaultEntities returns default entity GID mappings.
func NewDefaultEntities() Entities {
	return Entities{
		// Player and environment
		Player: defaultPlayer,
		Torch:  defaultTorch,
		// Enemies
		Knight:   defaultKnight,
		Ghoul:    defaultGhoul,
		Skeleman: defaultSkeleman,
		Crawler:  defaultCrawler,
		Rat:      defaultRat,
		Bat:      defaultBat,
		Ent:      defaultEnt,
		// Bosses
		Gram:     defaultGram,
		Ferragus: defaultFerragus,
		Oscar:    defaultOscar,
		Acedian:  defaultAcedian,

		// Interactive objects
		Chest:    defaultChest,
		Grave:    defaultGrave,
		Door:     defaultDoor,
		Spike:    defaultSpike,
		FakeWall: defaultFakeWall,
		Block:    defaultBlock,
	}
}
