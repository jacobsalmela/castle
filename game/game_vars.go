package game

import (
	"fmt"
	"log"
	"strconv"

	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/pkg/tilemap"
	"game/prefabs"
)

// ECSEntityConstructor creates an ECS entity directly, returning its EntityId.
type ECSEntityConstructor func(world *ecs.World, x, y, w, h float64, props *tilemap.Properties) entities.EntityId

// Entity constructor functions - extracted from inline declarations for maintainability

// Player constructor
func createPlayer(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewPlayerPrefab(world, x, y)
}

// Enemy constructors
func createKnight(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewKnightPrefab(world, x, y, w, h, p.FlipX)
}

func createGhoul(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	eid := prefabs.NewGhoulPrefab(world, x, y, w, h, p.FlipX)

	// Apply custom properties from Tiled
	// Note: View field removed in Phase 5 (unused in system)
	if rocksStr, ok := p.Custom["rocks"]; ok {
		if rocks, err := strconv.Atoi(rocksStr); err == nil {
			prefabs.SetGhoulRocks(world, eid, rocks)
		}
	}
	if p.Custom["ai"] == "poacher" {
		prefabs.SetGhoulPoacher(world, eid, true)
	}

	return eid
}

func createSkeleman(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewSkelemanPrefab(world, x, y, w, h, p.FlipX)
}

func createCrawler(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewCrawlerPrefab(world, x, y, w, h, p.FlipX)
}

func createRat(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewRatPrefab(world, x, y, w, h, p.FlipX)
}

func createBat(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewBatPrefab(world, x, y, w, h, p.FlipX)
}

func createEnt(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewEntPrefab(world, x, y, w, h, p.FlipX)
}

// Boss constructors
func createGram(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewGramPrefab(world, x, y, w, h, p.FlipX)
}

func createAcedian(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewAcedianPrefab(world, x, y, w, h, p.FlipX)
}

func createFerragus(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewFerragusPrefab(world, x, y, w, h, p.FlipX)
}

func createOscar(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewOscarPrefab(world, x, y, w, h, p.FlipX)
}

// Interactive object constructors
func createTorch(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewTorchPrefab(world, x, y, w, h, p)
}

func createChest(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewChestPrefab(world, x, y, w, h, p)
}

func createGrave(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	text := p.Custom["text"]
	return prefabs.NewGravePrefab(world, x, y, w, h, text)
}

func createDoor(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewDoorPrefab(world, x, y, w, h, p)
}

func createSpike(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewSpikePrefab(world, x, y, w, h, p)
}

func createFakeWall(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewFakeWallPrefab(world, x, y, w, h, p)
}

func createBlock(world *ecs.World, x, y, w, h float64, p *tilemap.Properties) entities.EntityId {
	return prefabs.NewBlockPrefab(world, x, y, w, h)
}

// ecsEntityBinds maps Tiled GIDs to ECS entity constructors.
// This is package-private and accessed through GetEntityConstructor.
// The GID values are loaded from config.yml and can be changed at runtime.
var ecsEntityBinds = make(map[uint32]ECSEntityConstructor)

// GetEntityConstructor returns the constructor function for the given GID.
// Returns nil if no constructor is registered for that GID.
func GetEntityConstructor(gid uint32) ECSEntityConstructor {
	return ecsEntityBinds[gid]
}

// InitializeEntityBindings builds the ecsEntityBinds map from config values.
// This must be called after config is loaded but before any maps are loaded.
func InitializeEntityBindings(cfg *config.Config) {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	entities := cfg.Entities

	ecsEntityBinds = map[uint32]ECSEntityConstructor{
		// Player (note: torch is handled separately in lighting system)
		entities.Player: createPlayer,

		// Enemies
		entities.Knight:   createKnight,
		entities.Ghoul:    createGhoul,
		entities.Skeleman: createSkeleman,
		entities.Crawler:  createCrawler,
		entities.Rat:      createRat,
		entities.Bat:      createBat,
		entities.Ent:      createEnt,

		// Bosses
		entities.Gram:     createGram,
		entities.Ferragus: createFerragus,
		entities.Oscar:    createOscar,
		entities.Acedian:  createAcedian,

		// Interactive objects
		entities.Torch:    createTorch,
		entities.Chest:    createChest,
		entities.Grave:    createGrave,
		entities.Door:     createDoor,
		entities.Spike:    createSpike,
		entities.FakeWall: createFakeWall,
		entities.Block:    createBlock,
	}
}

// ValidateEntityBindings checks that all GID constants have corresponding constructors.
// This should be called during game initialization to catch configuration errors early.
//
// Returns an error if any GID is missing a constructor, if the bindings map is empty,
// or if any config GID values are invalid (zero or duplicated).
func ValidateEntityBindings(cfg *config.Config) error {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Initialize entity bindings from config
	InitializeEntityBindings(cfg)

	entities := cfg.Entities

	if len(ecsEntityBinds) == 0 {
		return fmt.Errorf("ecsEntityBinds map is empty - no entity constructors registered")
	}

	// List of all GIDs that should be registered (now from config)
	requiredGIDs := []struct {
		gid  uint32
		name string
	}{
		// Player
		{entities.Player, "Player"},
		// Enemies
		{entities.Knight, "Knight"},
		{entities.Ghoul, "Ghoul"},
		{entities.Skeleman, "Skeleman"},
		{entities.Crawler, "Crawler"},
		{entities.Rat, "Rat"},
		{entities.Bat, "Bat"},
		{entities.Ent, "Ent"},
		// Bosses
		{entities.Gram, "Gram"},
		{entities.Acedian, "Acedian"},
		{entities.Ferragus, "Ferragus"},
		{entities.Oscar, "Oscar"},
		// Interactive objects
		{entities.Torch, "Torch"},
		{entities.Chest, "Chest"},
		{entities.Grave, "Grave"},
		{entities.Door, "Door"},
		{entities.Spike, "Spike"},
		{entities.FakeWall, "FakeWall"},
		{entities.Block, "Block"},
	}

	// Check for zero GIDs (invalid config)
	var zeroGIDs []string
	for _, req := range requiredGIDs {
		if req.gid == 0 {
			zeroGIDs = append(zeroGIDs, req.name)
		}
	}
	if len(zeroGIDs) > 0 {
		return fmt.Errorf("config has zero GID values for: %v", zeroGIDs)
	}

	// Check for duplicate GIDs (config collision)
	gidUsage := make(map[uint32][]string)
	for _, req := range requiredGIDs {
		gidUsage[req.gid] = append(gidUsage[req.gid], req.name)
	}
	var duplicates []string
	for gid, names := range gidUsage {
		if len(names) > 1 {
			duplicates = append(duplicates, fmt.Sprintf("GID %d used by: %v", gid, names))
		}
	}
	if len(duplicates) > 0 {
		return fmt.Errorf("duplicate GIDs in config: %v", duplicates)
	}

	// Check each required GID has a constructor
	var missing []string
	for _, req := range requiredGIDs {
		if _, exists := ecsEntityBinds[req.gid]; !exists {
			missing = append(missing, fmt.Sprintf("%s (GID %d)", req.name, req.gid))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing entity bindings for: %v", missing)
	}

	if cfg.DebugConsole {
		log.Printf("  ✓ Validated %d entity bindings", len(ecsEntityBinds))
	}
	return nil
}
