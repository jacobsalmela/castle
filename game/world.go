package game

import (
	"fmt"
	"log"
	"time"

	"game/assets/maps"
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/pkg/tilemap"
	"game/prefabs"
	"game/resources"
	"game/systems/draw/lighting"
)

// ╔════════════════════════════════════════════════════════════════════════════════╗
// ║                          WORLD LIFECYCLE MANAGEMENT                            ║
// ║                                                                                ║
// ║  This file handles ECS world initialization and configuration.                ║
// ║                                                                                ║
// ║  Responsibilities:                                                             ║
// ║  • Create and configure ECS worlds                                            ║
// ║  • Load maps and bind to worlds                                               ║
// ║  • Register resources and spawn entities                                      ║
// ║  • Apply save data to worlds                                                  ║
// ║                                                                                ║
// ║  Used by:                                                                      ║
// ║  • game.go - Initial game setup (NewGame)                                     ║
// ║  • game_reset.go - World reset after death (ResetWithMap)                    ║
// ╚════════════════════════════════════════════════════════════════════════════════╝

// WorldInitResult contains everything needed to construct a Game instance.
type WorldInitResult struct {
	World    *ecs.World
	Viewport *components.ViewPort
	SaveData *SaveData
}

// InitializeWorld creates and fully configures a new ECS world.
// This is the main entry point for world creation, used by both initial
// game setup and world resets.
//
// Initialization Flow:
//  1. Load save data (independent of world)
//  2. Load map and initialize rendering systems
//  3. Create and configure ECS world
//  4. Spawn entities and apply save state
//  5. Finalize world state (camera, time control)
//
// Parameters:
//   - cfg: Game configuration. If nil, uses default config.
//
// Returns: WorldInitResult containing configured world, viewport, and save data, or error
func InitializeWorld(cfg *config.Config) (WorldInitResult, error) {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Phase 1: Load persistent game state
	saveData, err := loadSaveData(cfg.Debug)
	if err != nil {
		return WorldInitResult{}, fmt.Errorf("initializing world: %w", err)
	}

	// Phase 2: Load map and initialize rendering
	worldMap := loadStartingMap(cfg)
	initializeRendering(worldMap, cfg)

	// Phase 3: Create ECS world and bind map
	world := createWorldWithMap(worldMap, cfg)

	// Phase 4: Register all ECS resources
	registerWorldResources(world)

	// Phase 5: Create viewport and register it
	viewport := createAndRegisterViewport(world)

	// Phase 6: Spawn entities from map
	loadEntitiesECSOnly(worldMap, world)

	// Phase 7: Apply save data (spawn player, restore stats, opened state)
	applySaveDataToWorld(world, saveData, worldMap)

	// Phase 8: Finalize world state (camera tracking, time control)
	finalizeWorldState(world)

	return WorldInitResult{
		World:    world,
		Viewport: viewport,
		SaveData: saveData,
	}, nil
}

// InitializeWorldWithMap creates a new ECS world with a specific map and save data.
// This is used for world resets where we want to preserve or change the map.
//
// Parameters:
//   - cfg: Game configuration. If nil, extracts from existing world or uses default.
//   - worldMap: Map to use (if nil, will preserve existing or load default)
//   - saveData: Save data to apply (must not be nil)
//   - existingWorld: Optional existing world to extract map from
//   - existingViewport: Optional existing viewport to preserve (avoids 0-dimension issues)
//
// Returns: Configured world and viewport
func InitializeWorldWithMap(cfg *config.Config, worldMap *tilemap.Map, saveData *SaveData, existingWorld *ecs.World, existingViewport *components.ViewPort) (*ecs.World, *components.ViewPort) {
	// Extract config from existing world if not provided
	if cfg == nil && existingWorld != nil {
		cfg = ecs.Resource[config.Config](existingWorld)
	}
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	if saveData == nil {
		log.Println("Warning: InitializeWorldWithMap called with nil saveData, loading defaults")
		saveData = newDefaultSaveData()
	}

	// Phase 1: Determine which map to use
	worldMap = determineMapToUse(worldMap, existingWorld)
	initializeRendering(worldMap, cfg)

	// Phase 2: Create ECS world and bind map
	world := createWorldWithMap(worldMap, cfg)

	// Phase 3: Register all ECS resources
	registerWorldResources(world)

	// Phase 4: Create or reuse viewport and register it
	viewport := determineViewportToUse(existingViewport, world)

	// Phase 5: Spawn entities from map
	loadEntitiesECSOnly(worldMap, world)

	// Phase 6: Apply save data (spawn player, restore stats, opened state)
	applySaveDataToWorld(world, saveData, worldMap)

	// Phase 7: Finalize world state (camera tracking, time control)
	finalizeWorldState(world)

	return world, viewport
}

// ============================================================================
// PHASE 1: SAVE DATA & MAP LOADING
// ============================================================================

// loadSaveData loads save data from disk.
func loadSaveData(debug bool) (*SaveData, error) {
	saveData, err := LoadSaveData(debug)
	if err != nil {
		return nil, fmt.Errorf("loading save data: %w", err)
	}
	return saveData, nil
}

// loadStartingMap loads the starting map from config.
// TODO: Pass config as parameter to avoid global access (called during initialization)
func loadStartingMap(cfg *config.Config) *tilemap.Map {
	mapPath := cfg.World.StartingMap
	return tilemap.NewMap(mapPath, 1, maps.FS, config.PipelineScreenTag, config.PipelineNormalMapTag)
}

// determineMapToUse selects which map to use for world initialization.
// Priority: provided map > existing world's map > default map from config
func determineMapToUse(worldMap *tilemap.Map, existingWorld *ecs.World) *tilemap.Map {
	// Use provided map if available
	if worldMap != nil {
		return worldMap
	}

	// Try to get map from existing world
	if existingWorld != nil {
		if mapRef := ecs.Resource[resources.MapRef](existingWorld); mapRef != nil {
			worldMap = mapRef.Map
		}
	}

	// Fall back to default map
	if worldMap == nil {
		log.Println("Warning: No map provided, loading default map")
		cfg := ecs.Resource[config.Config](existingWorld)
		if cfg == nil {
			cfg = config.NewDefaultConfig()
		}
		worldMap = loadStartingMap(cfg)
	}

	return worldMap
}

// ============================================================================
// PHASE 2: RENDERING INITIALIZATION
// ============================================================================

// initializeRendering sets up the lighting system with map data.
func initializeRendering(worldMap *tilemap.Map, cfg *config.Config) {
	lighting.LoadLighting(worldMap, cfg.Entities.Torch)
}

// ============================================================================
// PHASE 3: WORLD CREATION
// ============================================================================

// createWorldWithMap creates a new ECS world and binds the map to it.
func createWorldWithMap(worldMap *tilemap.Map, cfg *config.Config) *ecs.World {
	world := ecs.NewWorld(cfg)

	// Bind map via resource system (NewWorld already created MapRef)
	world.BindMap(worldMap, "rooms", "collisions")

	return world
}

// ============================================================================
// PHASE 4: RESOURCE REGISTRATION
// ============================================================================

// registerWorldResources registers all ECS resources needed for game systems.
func registerWorldResources(world *ecs.World) {
	registerCoreResources(world)
	registerDebugResources(world)
	registerSpatialResources(world)
}

// registerCoreResources registers core game system resources.
func registerCoreResources(world *ecs.World) {
	world.SetResource(resources.NewTimeControl())
	world.SetResource(resources.NewEventQueue()) // Unified event queue (hits, impacts, triggers)
	world.SetResource(resources.NewRenderQueue())
	world.SetResource(resources.NewTransitionManager())
	world.SetResource(&resources.GameSignals{})
	world.SetResource(&resources.RenderStats{})
	world.SetResource(resources.NewSpikeCooldown()) // Spike damage cooldown tracking
}

// registerDebugResources registers debug-specific resources.
func registerDebugResources(world *ecs.World) {
	world.SetResource(resources.NewDebugCategories())
	world.SetResource(resources.NewDebugState()) // Pure ECS debug state for toggle flags
	world.SetResource(resources.NewCollisionEventQueue(2 * time.Second))
	world.SetResource(resources.NewHitboxEventQueue(1 * time.Second))
}

// registerSpatialResources registers collision space and camera resources.
// NOTE: Spatial resources (Camera, Space) are now initialized in ecs.NewWorld(),
// so this function is a no-op. Kept for backward compatibility with the resource
// registration phase.
func registerSpatialResources(world *ecs.World) {
	// No-op: Camera and Space are already registered as resources in ecs.NewWorld()
}

// ============================================================================
// PHASE 5: VIEWPORT CREATION
// ============================================================================

// createAndRegisterViewport creates a viewport and registers it as a resource.
func createAndRegisterViewport(world *ecs.World) *components.ViewPort {
	viewport := prefabs.NewViewport()
	world.SetResource(viewport)
	return viewport
}

// determineViewportToUse selects viewport to use for world initialization.
// Priority: existing viewport (preserves dimensions) > new viewport
// This is critical during reset to avoid 0-dimension viewports before LayoutF runs.
func determineViewportToUse(existingViewport *components.ViewPort, world *ecs.World) *components.ViewPort {
	if existingViewport != nil {
		// Reuse existing viewport to preserve dimensions
		world.SetResource(existingViewport)
		return existingViewport
	}

	// Create new viewport (will have 0 dimensions until LayoutF is called)
	return createAndRegisterViewport(world)
}

// ============================================================================
// PHASE 6: ENTITY SPAWNING (from game_vars.go)
// ============================================================================
// loadEntitiesECSOnly is imported from game_vars.go - no changes needed here

// ============================================================================
// PHASE 7: SAVE DATA APPLICATION
// ============================================================================

// applySaveDataToWorld applies loaded save data to the ECS world.
// This is a mid-level orchestrator that coordinates three main phases:
//  1. Spawn player at the correct position
//  2. Restore player stats from save
//  3. Restore opened state for interactable entities
func applySaveDataToWorld(world *ecs.World, sd *SaveData, worldMap *tilemap.Map) {
	if world == nil || sd == nil {
		return
	}

	// Get config from world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Phase 1: Determine spawn position and create player entity
	spawnPosition := determinePlayerSpawnPosition(sd, worldMap, cfg)
	playerID := spawnPlayerAtPosition(world, spawnPosition)

	// Phase 2: Restore player stats from save data
	restorePlayerStats(world, playerID, sd)

	// Phase 3: Restore opened state for interactable entities
	restoreOpenedEntities(world, sd)
}

// spawnPosition holds player spawn coordinates.
type spawnPosition struct {
	X, Y float64
}

// determinePlayerSpawnPosition gets player spawn position from save or map default.
func determinePlayerSpawnPosition(sd *SaveData, worldMap *tilemap.Map, cfg *config.Config) spawnPosition {
	// Use saved position if valid
	if sd.PlayerData.X != 0 || sd.PlayerData.Y != 0 {
		return spawnPosition{X: sd.PlayerData.X, Y: sd.PlayerData.Y}
	}

	// Fall back to map default spawn point
	if worldMap != nil {
		obj, err := tilemap.FindObjectFromTileID(worldMap, cfg.Entities.Player, "entities")
		if err == nil && obj != nil {
			return spawnPosition{X: obj.X, Y: obj.Y}
		}
	}

	return spawnPosition{X: 0, Y: 0}
}

// spawnPlayerAtPosition creates the player entity.
func spawnPlayerAtPosition(world *ecs.World, pos spawnPosition) entities.EntityId {
	playerID := prefabs.NewPlayerPrefab(world, pos.X, pos.Y)
	return playerID
}

// restorePlayerStats restores all player stats from save data.
func restorePlayerStats(world *ecs.World, playerID entities.EntityId, sd *SaveData) {
	if playerID == 0 {
		return
	}

	restoreHealthComponent(world, playerID, sd)
	restoreStaminaComponent(world, playerID, sd)
	restorePoiseComponent(world, playerID, sd)
	restoreHealingComponent(world, playerID, sd)
	restoreExperienceComponent(world, playerID, sd)
	restoreKeyBindings(world, playerID, sd)
}

// restoreHealthComponent restores the Health component from save data.
func restoreHealthComponent(world *ecs.World, playerID entities.EntityId, sd *SaveData) {
	health := ecs.GetComponent[components.Health](world, playerID)
	if health == nil || sd.PlayerData.Health.Max <= 0 {
		return
	}

	health.Current = sd.PlayerData.Health.Current
	health.Max = sd.PlayerData.Health.Max

	// Ensure current health is never zero (prevents instant death on load)
	if health.Current <= 0 {
		health.Current = health.Max
	}
}

// restoreStaminaComponent restores the Stamina component from save data.
func restoreStaminaComponent(world *ecs.World, playerID entities.EntityId, sd *SaveData) {
	stamina := ecs.GetComponent[components.Stamina](world, playerID)
	if stamina == nil || sd.PlayerData.Stamina.Max <= 0 {
		return
	}

	stamina.Current = sd.PlayerData.Stamina.Current
	stamina.Max = sd.PlayerData.Stamina.Max
}

// restorePoiseComponent restores the Poise component from save data.
func restorePoiseComponent(world *ecs.World, playerID entities.EntityId, sd *SaveData) {
	poise := ecs.GetComponent[components.Poise](world, playerID)
	if poise == nil || sd.PlayerData.Poise.Max <= 0 {
		return
	}

	poise.Current = sd.PlayerData.Poise.Current
	poise.Max = sd.PlayerData.Poise.Max
}

// restoreHealingComponent restores the Healing component from save data.
func restoreHealingComponent(world *ecs.World, playerID entities.EntityId, sd *SaveData) {
	healing := ecs.GetComponent[components.Healing](world, playerID)
	if healing == nil || sd.PlayerData.Healing.MaxCount <= 0 {
		return
	}

	healing.Count = sd.PlayerData.Healing.Count
	healing.MaxCount = sd.PlayerData.Healing.MaxCount
	healing.HealAmount = sd.PlayerData.Healing.HealAmount
}

// restoreExperienceComponent restores the Experience component from save data.
func restoreExperienceComponent(world *ecs.World, playerID entities.EntityId, sd *SaveData) {
	exp := ecs.GetComponent[components.Experience](world, playerID)
	if exp == nil {
		return
	}

	exp.Points = sd.PlayerData.Experience.Points
}

// restoreKeyBindings restores key bindings to the Input component from save data.
func restoreKeyBindings(world *ecs.World, playerID entities.EntityId, sd *SaveData) {
	input := ecs.GetComponent[components.Input](world, playerID)
	if input == nil {
		return
	}

	// Pure ECS: inline LoadBindings
	input.KeyBindings = deserializeInputBindings(sd.Pad)
}

// restoreOpenedEntities restores the opened state for doors, chests, and walls.
func restoreOpenedEntities(world *ecs.World, sd *SaveData) {
	for _, openedID := range sd.Opened {
		eid := world.GetEntityByID(openedID)
		if eid == 0 {
			continue
		}

		// Try each type of openable entity
		if restoreOpenedDoor(world, eid) {
			continue
		}
		if restoreOpenedChest(world, eid) {
			continue
		}
		restoreOpenedFakeWall(world, eid)
	}
}

// restoreOpenedDoor restores a door's opened state. Returns true if entity has Door component.
func restoreOpenedDoor(world *ecs.World, eid entities.EntityId) bool {
	door := ecs.GetComponent[components.Door](world, eid)
	if door == nil {
		return false
	}

	door.Opened = true
	return true
}

// restoreOpenedChest restores a chest's opened state. Returns true if entity has Chest component.
func restoreOpenedChest(world *ecs.World, eid entities.EntityId) bool {
	chest := ecs.GetComponent[components.Chest](world, eid)
	if chest == nil {
		return false
	}

	chest.Opened = true
	chest.AnimationStage = 2 // Fully open state

	// Remove hitbox so chest can't be hit again
	if ecs.GetComponent[components.Hitbox](world, eid) != nil {
		ecs.RemoveComponent[components.Hitbox](world, eid)
	}

	// Update animation to open state if Animation component exists
	if anim := ecs.GetComponent[components.Animation](world, eid); anim != nil && anim.Data != nil {
		// Set animation to final frame of "activate" tag (fully open)
		if err := anim.Data.Play("activate"); err == nil {
			anim.State = "activate"
			// Jump to last frame of animation (fully open)
			if anim.Data.CurrentAnimation != nil {
				anim.Frame = anim.Data.CurrentAnimation.To - anim.Data.CurrentAnimation.From
			}
		}
	}

	return true
}

// restoreOpenedFakeWall restores a fake wall's opened state. Returns true if entity has FakeWall component.
func restoreOpenedFakeWall(world *ecs.World, eid entities.EntityId) bool {
	fakeWall := ecs.GetComponent[components.FakeWall](world, eid)
	if fakeWall == nil {
		return false
	}

	fakeWall.Opened = true
	return true
}

// ============================================================================
// PHASE 8: WORLD FINALIZATION
// ============================================================================

// finalizeWorldState completes world setup (camera tracking, time control).
func finalizeWorldState(world *ecs.World) {
	setupCameraTracking(world)
	initializeTimeControl(world)
}

// setupCameraTracking makes the camera follow the player entity.
// **FIX FOR ISSUE #3 (REVISED)**: Set camera position AND room borders before follow
func setupCameraTracking(world *ecs.World) {
	// Get player entity
	players := world.EntitiesWith((*components.Player)(nil))
	if len(players) == 0 {
		return
	}
	playerID := players[0]

	playerTransform := ecs.GetComponent[components.Transform](world, playerID)
	if playerTransform == nil {
		return
	}

	// Get player center position
	px, py, pw, ph := playerTransform.Rect()
	playerCenterX := px + pw/2
	playerCenterY := py + ph/2

	// Get camera dimensions from viewport
	viewport := ecs.Resource[components.ViewPort](world)
	if viewport == nil {
		// Fallback: just follow without initial positioning
		world.Follow(playerTransform)
		return
	}

	// Calculate camera position centered on player
	cam := ecs.Resource[resources.Camera](world)
	if cam != nil {
		// Camera should be centered on player, with camera dimensions from viewport
		camW := viewport.LW
		camH := viewport.LH

		// Center camera on player BEFORE calling Follow
		initialX := playerCenterX - camW/2
		initialY := playerCenterY - camH/2
		cam.SetPosition(initialX, initialY)
	}

	// Now follow - this will call SetRoomBorders(false) which won't trigger transition
	world.Follow(playerTransform)
}

// initializeTimeControl sets up the time control system with default values.
func initializeTimeControl(world *ecs.World) {
	timeControl := ecs.Resource[resources.TimeControl](world)
	if timeControl == nil {
		return
	}

	timeControl.SetSpeed(1)
}
