package game

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
	"syscall"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/systems/update/preupdate"

	"github.com/fsnotify/fsnotify"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	Persistent = true // Enable persistent save to disk
	SavePath   = "save.json"
	fileMode   = 0666
)

type Opener interface {
	Open()
	Opened() bool
}

type PlayerData struct {
	X   float64 `json:"x"`
	Y   float64 `json:"y"`
	Exp int     `json:"exp"` // Kept for backwards compatibility, use Experience instead

	// Pure ECS stat components
	Health struct {
		Current float64 `json:"current"`
		Max     float64 `json:"max"`
	} `json:"health"`

	Stamina struct {
		Current float64 `json:"current"`
		Max     float64 `json:"max"`
	} `json:"stamina"`

	Poise struct {
		Current float64 `json:"current"`
		Max     float64 `json:"max"`
	} `json:"poise"`

	Healing struct {
		Count      int     `json:"count"`
		MaxCount   int     `json:"max_count"`
		HealAmount float64 `json:"heal_amount"`
	} `json:"healing"`

	Experience struct {
		Points int `json:"points"`
	} `json:"experience"`
}

type SaveData struct {
	PlayerData PlayerData `json:"player_data"`
	Pad        [9][]int   `json:"keys"` // Stored as int array for JSON compatibility
	Opened     []uint     `json:"opened"`
}

// LoadSaveData reads save data from disk, creating default save if missing (like config pattern).
// This function is independent of ECS world and can be called during initialization.
func LoadSaveData(debug bool) (*SaveData, error) {
	// Try to read existing save file
	data, err := os.ReadFile(SavePath)
	if err != nil {
		if os.IsNotExist(err) || (!Persistent && errors.Is(err, syscall.ENOSYS)) {
			// File doesn't exist, create default save
			defaultSave := newDefaultSaveData()

			// Optionally write default save to disk (like config does)
			if Persistent {
				if saveErr := saveSaveData(defaultSave, debug); saveErr != nil {
					log.Printf("Warning: failed to create default save file: %v", saveErr)
				} else {
					log.Println("Created default save.json")
				}
			}

			return defaultSave, nil
		}
		return nil, err
	}

	// Parse existing save file
	var saveData SaveData
	if err := json.Unmarshal(data, &saveData); err != nil {
		return nil, err
	}

	// Migrate old save data to new format if needed
	migrateSaveData(&saveData)

	return &saveData, nil
}

// newDefaultSaveData creates default save data with default key bindings.
// Player position will be set from map spawn point during world initialization.
func newDefaultSaveData() *SaveData {
	return &SaveData{
		Pad: serializeInputBindings(preupdate.DefaultKeyBindings()),
		// PlayerData fields will be populated from map spawn point later
	}
}

// saveSaveData writes save data to disk.
func saveSaveData(sd *SaveData, debug bool) error {
	if !Persistent {
		return nil
	}

	// Serialize to JSON
	data, err := json.Marshal(sd)
	if err != nil {
		return err
	}

	// Pretty-print in debug mode
	if debug {
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err != nil {
			return err
		}
		data = buf.Bytes()
	}

	// Write to file
	return os.WriteFile(SavePath, data, fileMode)
}

// Save writes the current game state to disk.
// This updates the in-memory save data with current world state, then writes to file.
func (g *Game) Save() error {
	// Update save data from current game state
	g.populateSaveData(g.saveData)

	// Get debug flag from config
	debug := false
	if g.world != nil {
		if cfg := ecs.Resource[config.Config](g.world); cfg != nil {
			debug = cfg.Debug
		}
	}

	// Write to disk
	return saveSaveData(g.saveData, debug)
}

// ReloadSave updates the game state from externally modified save data.
// This is typically called by WatchSave when the save file changes on disk.
func (g *Game) ReloadSave(sd *SaveData) {
	if sd == nil || g.world == nil {
		return
	}

	// Update the in-memory save data
	g.saveData = sd

	// Apply the new save data to the current world
	// Note: This only updates stats, not position (to avoid teleporting the player)
	g.ApplySaveData(sd)

	log.Println("  ✓ Save data reloaded into game")
}

// migrateSaveData fills in missing fields from old save files with default values.
// This prevents the "zero health death loop" when loading saves created before new fields were added.
func migrateSaveData(sd *SaveData) {
	if sd == nil {
		return
	}

	// Check if save data needs migration (missing new stat fields)
	needsMigration := sd.PlayerData.Health.Max == 0 &&
		sd.PlayerData.Stamina.Max == 0 &&
		sd.PlayerData.Poise.Max == 0

	if !needsMigration {
		return
	}

	log.Println("Migrating old save data to new format...")

	// Apply default player stats (from prefabs/player.go constants)
	const (
		defaultMaxHealth  = 60.0
		defaultMaxStamina = 65.0
		defaultMaxPoise   = 100.0
		defaultMaxHeal    = 5
		defaultHealAmount = 20.0
	)

	// Set health defaults (if missing)
	if sd.PlayerData.Health.Max == 0 {
		sd.PlayerData.Health.Current = defaultMaxHealth
		sd.PlayerData.Health.Max = defaultMaxHealth
	}

	// Set stamina defaults (if missing)
	if sd.PlayerData.Stamina.Max == 0 {
		sd.PlayerData.Stamina.Current = defaultMaxStamina
		sd.PlayerData.Stamina.Max = defaultMaxStamina
	}

	// Set poise defaults (if missing)
	if sd.PlayerData.Poise.Max == 0 {
		sd.PlayerData.Poise.Current = defaultMaxPoise
		sd.PlayerData.Poise.Max = defaultMaxPoise
	}

	// Set healing defaults (if missing)
	if sd.PlayerData.Healing.MaxCount == 0 {
		sd.PlayerData.Healing.Count = defaultMaxHeal
		sd.PlayerData.Healing.MaxCount = defaultMaxHeal
		sd.PlayerData.Healing.HealAmount = defaultHealAmount
	}

	log.Printf("  ✓ Migrated save: Health=%v, Stamina=%v, Poise=%v, Healing=%v, Exp=%v",
		sd.PlayerData.Health.Max,
		sd.PlayerData.Stamina.Max,
		sd.PlayerData.Poise.Max,
		sd.PlayerData.Healing.MaxCount,
		sd.PlayerData.Experience.Points)
}

// ApplySaveData restores saved game state to the ECS world.
// This is a high-level orchestrator for applying save data during hot-reload.
// NOTE: This must be called AFTER the player entity is created.
func (g *Game) ApplySaveData(sd *SaveData) {
	if sd == nil {
		return
	}

	// Phase 1: Restore player stats from save data
	g.applySavePlayerStats(sd)

	// Phase 2: Restore opened state for interactive entities
	g.applySaveOpenedEntities(sd)
}

// applySavePlayerStats restores all player component stats from save data.
func (g *Game) applySavePlayerStats(sd *SaveData) {
	// Get player entity
	players := g.world.EntitiesWith((*components.Player)(nil))
	if len(players) == 0 {
		return
	}
	playerID := players[0]

	// Restore each player component from save data
	g.applySaveHealthComponent(playerID, sd)
	g.applySaveStaminaComponent(playerID, sd)
	g.applySavePoiseComponent(playerID, sd)
	g.applySaveHealingComponent(playerID, sd)
	g.applySaveExperienceComponent(playerID, sd)
	g.applySaveKeyBindings(playerID, sd)
}

// applySaveHealthComponent restores the Health component from save data.
func (g *Game) applySaveHealthComponent(playerID entities.EntityId, sd *SaveData) {
	health := ecs.GetComponent[components.Health](g.world, playerID)
	if health == nil || sd.PlayerData.Health.Max <= 0 {
		return
	}

	health.Current = sd.PlayerData.Health.Current
	health.Max = sd.PlayerData.Health.Max

	// Ensure current health is never zero (prevents instant death on hot-reload)
	if health.Current <= 0 {
		health.Current = health.Max
	}
}

// applySaveStaminaComponent restores the Stamina component from save data.
func (g *Game) applySaveStaminaComponent(playerID entities.EntityId, sd *SaveData) {
	stamina := ecs.GetComponent[components.Stamina](g.world, playerID)
	if stamina == nil || sd.PlayerData.Stamina.Max <= 0 {
		return
	}

	stamina.Current = sd.PlayerData.Stamina.Current
	stamina.Max = sd.PlayerData.Stamina.Max
}

// applySavePoiseComponent restores the Poise component from save data.
func (g *Game) applySavePoiseComponent(playerID entities.EntityId, sd *SaveData) {
	poise := ecs.GetComponent[components.Poise](g.world, playerID)
	if poise == nil || sd.PlayerData.Poise.Max <= 0 {
		return
	}

	poise.Current = sd.PlayerData.Poise.Current
	poise.Max = sd.PlayerData.Poise.Max
}

// applySaveHealingComponent restores the Healing component from save data.
func (g *Game) applySaveHealingComponent(playerID entities.EntityId, sd *SaveData) {
	healing := ecs.GetComponent[components.Healing](g.world, playerID)
	if healing == nil || sd.PlayerData.Healing.MaxCount <= 0 {
		return
	}

	healing.Count = sd.PlayerData.Healing.Count
	healing.MaxCount = sd.PlayerData.Healing.MaxCount
	healing.HealAmount = sd.PlayerData.Healing.HealAmount
}

// applySaveExperienceComponent restores the Experience component from save data.
func (g *Game) applySaveExperienceComponent(playerID entities.EntityId, sd *SaveData) {
	exp := ecs.GetComponent[components.Experience](g.world, playerID)
	if exp == nil {
		return
	}

	exp.Points = sd.PlayerData.Experience.Points
}

// applySaveKeyBindings restores key bindings to the Input component from save data.
func (g *Game) applySaveKeyBindings(playerID entities.EntityId, sd *SaveData) {
	input := ecs.GetComponent[components.Input](g.world, playerID)
	if input == nil {
		return
	}

	// Pure ECS: inline LoadBindings
	input.KeyBindings = deserializeInputBindings(sd.Pad)
}

// applySaveOpenedEntities restores the opened state for doors, chests, and walls.
func (g *Game) applySaveOpenedEntities(sd *SaveData) {
	for _, openedID := range sd.Opened {
		eid := g.world.GetEntityByID(openedID)
		if eid == 0 {
			continue
		}

		// Try each type of openable entity
		if g.applySaveOpenedDoor(eid) {
			continue
		}
		if g.applySaveOpenedChest(eid) {
			continue
		}
		g.applySaveOpenedFakeWall(eid)
	}
}

// applySaveOpenedDoor restores a door's opened state. Returns true if entity has Door component.
func (g *Game) applySaveOpenedDoor(eid entities.EntityId) bool {
	door := ecs.GetComponent[components.Door](g.world, eid)
	if door == nil {
		return false
	}

	door.Opened = true
	return true
}

// applySaveOpenedChest restores a chest's opened state. Returns true if entity has Chest component.
func (g *Game) applySaveOpenedChest(eid entities.EntityId) bool {
	chest := ecs.GetComponent[components.Chest](g.world, eid)
	if chest == nil {
		return false
	}

	chest.Opened = true
	chest.AnimationStage = 2 // Fully open state

	// Remove hitbox so chest can't be hit again
	if ecs.GetComponent[components.Hitbox](g.world, eid) != nil {
		ecs.RemoveComponent[components.Hitbox](g.world, eid)
	}

	// Update animation to open state if Animation component exists
	if anim := ecs.GetComponent[components.Animation](g.world, eid); anim != nil && anim.Data != nil {
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

// applySaveOpenedFakeWall restores a fake wall's opened state. Returns true if entity has FakeWall component.
func (g *Game) applySaveOpenedFakeWall(eid entities.EntityId) bool {
	fakeWall := ecs.GetComponent[components.FakeWall](g.world, eid)
	if fakeWall == nil {
		return false
	}

	fakeWall.Opened = true
	return true
}

// populateSaveData updates save data with current game state.
// This is a mid-level orchestrator that coordinates saving player and entity state.
func (g *Game) populateSaveData(sd *SaveData) {
	// Phase 1: Save player state (position and all component stats)
	g.populatePlayerData(sd)

	// Phase 2: Save opened entity state (doors, chests, walls)
	sd.Opened = sd.Opened[:0]
	g.saveOpened(sd)
}

// populatePlayerData saves all player component data to save data.
func (g *Game) populatePlayerData(sd *SaveData) {
	// Get player entity
	players := g.world.EntitiesWith((*components.Player)(nil))
	if len(players) == 0 {
		return
	}
	playerID := players[0]

	// Save each component's data
	g.populatePlayerTransform(playerID, sd)
	g.populatePlayerHealth(playerID, sd)
	g.populatePlayerStamina(playerID, sd)
	g.populatePlayerPoise(playerID, sd)
	g.populatePlayerHealing(playerID, sd)
	g.populatePlayerExperience(playerID, sd)
	g.populatePlayerKeyBindings(playerID, sd)
}

// populatePlayerTransform saves player position from Transform component.
func (g *Game) populatePlayerTransform(playerID entities.EntityId, sd *SaveData) {
	transform := ecs.GetComponent[components.Transform](g.world, playerID)
	if transform == nil {
		return
	}

	sd.PlayerData.X = transform.X
	sd.PlayerData.Y = transform.Y
}

// populatePlayerHealth saves Health component data.
func (g *Game) populatePlayerHealth(playerID entities.EntityId, sd *SaveData) {
	health := ecs.GetComponent[components.Health](g.world, playerID)
	if health == nil {
		return
	}

	sd.PlayerData.Health.Current = health.Current
	sd.PlayerData.Health.Max = health.Max
}

// populatePlayerStamina saves Stamina component data.
func (g *Game) populatePlayerStamina(playerID entities.EntityId, sd *SaveData) {
	stamina := ecs.GetComponent[components.Stamina](g.world, playerID)
	if stamina == nil {
		return
	}

	sd.PlayerData.Stamina.Current = stamina.Current
	sd.PlayerData.Stamina.Max = stamina.Max
}

// populatePlayerPoise saves Poise component data.
func (g *Game) populatePlayerPoise(playerID entities.EntityId, sd *SaveData) {
	poise := ecs.GetComponent[components.Poise](g.world, playerID)
	if poise == nil {
		return
	}

	sd.PlayerData.Poise.Current = poise.Current
	sd.PlayerData.Poise.Max = poise.Max
}

// populatePlayerHealing saves Healing component data.
func (g *Game) populatePlayerHealing(playerID entities.EntityId, sd *SaveData) {
	healing := ecs.GetComponent[components.Healing](g.world, playerID)
	if healing == nil {
		return
	}

	sd.PlayerData.Healing.Count = healing.Count
	sd.PlayerData.Healing.MaxCount = healing.MaxCount
	sd.PlayerData.Healing.HealAmount = healing.HealAmount
}

// populatePlayerExperience saves Experience component data.
func (g *Game) populatePlayerExperience(playerID entities.EntityId, sd *SaveData) {
	exp := ecs.GetComponent[components.Experience](g.world, playerID)
	if exp == nil {
		return
	}

	sd.PlayerData.Experience.Points = exp.Points
	sd.PlayerData.Exp = exp.Points // Backwards compatibility
}

// populatePlayerKeyBindings saves Input component key bindings.
func (g *Game) populatePlayerKeyBindings(playerID entities.EntityId, sd *SaveData) {
	input := ecs.GetComponent[components.Input](g.world, playerID)
	if input != nil {
		// Pure ECS: inline SerializableBindings
		sd.Pad = serializeInputBindings(input.KeyBindings)
	} else {
		// Fallback to default bindings if Input component not found
		sd.Pad = serializeInputBindings(preupdate.DefaultKeyBindings())
	}
}

// saveOpened records which interactive entities (doors, chests, walls) have been opened.
func (g *Game) saveOpened(sd *SaveData) {
	if sd == nil {
		return
	}

	seen := map[uint]struct{}{}
	appendIfOpened := func(eid entities.EntityId, id uint) {
		if id == 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		// Check Door component
		if door := ecs.GetComponent[components.Door](g.world, eid); door != nil && door.Opened {
			seen[id] = struct{}{}
			sd.Opened = append(sd.Opened, id)
			return
		}
		// Check Chest component
		if chest := ecs.GetComponent[components.Chest](g.world, eid); chest != nil && chest.Opened {
			seen[id] = struct{}{}
			sd.Opened = append(sd.Opened, id)
			return
		}
		// Check FakeWall component
		if fakeWall := ecs.GetComponent[components.FakeWall](g.world, eid); fakeWall != nil && fakeWall.Opened {
			seen[id] = struct{}{}
			sd.Opened = append(sd.Opened, id)
			return
		}
	}

	// Record opened items by ID from ECS state
	if g.world != nil {
		g.world.RangeEntities(appendIfOpened)
		g.world.RangeRemoved(appendIfOpened)
	}
}

// serializeInputBindings converts ebiten.Key bindings to int arrays for JSON serialization.
func serializeInputBindings(bindings [9][]ebiten.Key) [9][]int {
	var result [9][]int
	for i, keys := range bindings {
		result[i] = make([]int, len(keys))
		for j, key := range keys {
			result[i][j] = int(key)
		}
	}
	return result
}

// deserializeInputBindings converts int arrays from JSON back to ebiten.Key bindings.
func deserializeInputBindings(bindings [9][]int) [9][]ebiten.Key {
	var result [9][]ebiten.Key
	for i, keys := range bindings {
		result[i] = make([]ebiten.Key, len(keys))
		for j, key := range keys {
			result[i][j] = ebiten.Key(key)
		}
	}
	return result
}

// WatchSave watches the save file for external changes and reloads it.
// This allows external tools or manual edits to be reflected in the running game.
// Call this in a goroutine from main.go like: go game.WatchSave(savePath, onReload)
func WatchSave(savePath string, onReload func(*SaveData)) {
	watcher, err := createSaveFileWatcher()
	if err != nil {
		log.Printf("Failed to create save file watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Start the event loop in a goroutine
	go watchSaveEventLoop(watcher, savePath, onReload)

	// Set up file watching
	if err := setupSaveFileWatch(watcher, savePath); err != nil {
		return
	}

	select {} // block forever
}

// createSaveFileWatcher creates and returns a new file system watcher.
func createSaveFileWatcher() (*fsnotify.Watcher, error) {
	return fsnotify.NewWatcher()
}

// watchSaveEventLoop runs the event loop for processing file change events.
func watchSaveEventLoop(watcher *fsnotify.Watcher, savePath string, onReload func(*SaveData)) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			handleSaveFileEvent(event, savePath, onReload)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Save watcher error: %v", err)
		}
	}
}

// handleSaveFileEvent processes a single file system event.
func handleSaveFileEvent(event fsnotify.Event, savePath string, onReload func(*SaveData)) {
	// Only respond to write events for our save file
	if event.Op&fsnotify.Write == 0 || event.Name != savePath {
		return
	}

	// Reload save data from disk (debug=false for hot-reload)
	saveData, err := LoadSaveData(false)
	if err != nil {
		log.Printf("Save file reload error: %v", err)
		return
	}

	log.Println("(Hot) reloaded save:", savePath)

	// Call the reload callback to apply the new save data
	if onReload != nil {
		onReload(saveData)
	}
}

// setupSaveFileWatch configures the watcher to monitor the save file.
func setupSaveFileWatch(watcher *fsnotify.Watcher, savePath string) error {
	// Check if file exists before trying to watch it
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		log.Printf("Save file %s does not exist yet, skipping watch", savePath)
		return err
	}

	// Add the file to the watcher
	if err := watcher.Add(savePath); err != nil {
		log.Printf("Failed to watch save file: %v", err)
		return err
	}

	return nil
}
