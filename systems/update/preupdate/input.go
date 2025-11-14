package preupdate

import (
	"game/components"
	"game/ecs"
	"game/pkg/config"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const inputBufferDuration = 0.150 // seconds (150ms)

// DefaultKeyBindings returns the default key binding configuration.
// Used by player prefab and save system fallback.
// DEPRECATED: Use KeyBindingsFromConfig for configurable bindings.
func DefaultKeyBindings() [9][]ebiten.Key {
	return [9][]ebiten.Key{
		components.InputKeyRight:  {ebiten.KeyArrowRight, ebiten.KeyD},
		components.InputKeyLeft:   {ebiten.KeyArrowLeft, ebiten.KeyA},
		components.InputKeyUp:     {ebiten.KeyArrowUp, ebiten.KeyW},
		components.InputKeyDown:   {ebiten.KeyArrowDown, ebiten.KeyS},
		components.InputKeyJump:   {ebiten.KeyZ, ebiten.KeyN},
		components.InputKeyAction: {ebiten.KeyX, ebiten.KeyM},
		components.InputKeyGuard:  {ebiten.KeyC, ebiten.KeyB},
		components.InputKeyHeal:   {ebiten.KeyV, ebiten.KeyControl, ebiten.KeyControlLeft},
		components.InputKeyDash:   {ebiten.KeySpace},
	}
}

// KeyBindingsFromConfig returns key bindings from the config.
// If cfg is nil, returns default bindings.
func KeyBindingsFromConfig(cfg *config.Config) [9][]ebiten.Key {
	if cfg == nil {
		return DefaultKeyBindings()
	}
	return cfg.Input.ToEbitenKeys()
}

// KeyBindingsFromWorld returns key bindings from the world's config resource.
// Falls back to defaults if config is not available.
func KeyBindingsFromWorld(world *ecs.World) [9][]ebiten.Key {
	if world == nil {
		return DefaultKeyBindings()
	}
	cfg := ecs.Resource[config.Config](world)
	return KeyBindingsFromConfig(cfg)
}

// UpdateInput polls keyboard input and updates Input components in the ECS world.
// This system should run early in the update cycle before gameplay systems that read input.
//
// The system:
// - Syncs key bindings from config (supports hot-reload)
// - Polls physical keyboard state via ebiten
// - Updates KeyDown, KeyPressed, KeyReleased state for each logical key
// - Manages input buffering for better feel (coyote time for button presses)
// - Expires old buffered inputs based on game time
func UpdateInput(world *ecs.World, currentTime interface{}, _ float64) {
	if world == nil {
		return
	}

	// currentTime should be total elapsed seconds (float64)
	gameTime, ok := currentTime.(float64)
	if !ok {
		gameTime = 0 // Fallback if time not provided
	}

	// Get current bindings from global config (supports hot-reload via watcher)
	// The config watcher updates config.Cfg when the file changes
	currentBindings := config.Cfg.Input.ToEbitenKeys()

	// Find all entities with Input components (typically just the player)
	for _, eid := range world.EntitiesWith((*components.Input)(nil)) {
		input := ecs.GetComponent[components.Input](world, eid)
		if input == nil {
			continue
		}

		// Sync bindings from config (hot-reload support)
		// This ensures bindings are updated when config file changes
		syncInputBindings(input, currentBindings)

		updateInputState(input, gameTime)
	}
}

// syncInputBindings updates the Input component's bindings from config.
// This is called each frame to support hot-reload of key bindings.
func syncInputBindings(input *components.Input, configBindings [9][]ebiten.Key) {
	if input == nil {
		return
	}
	// Copy bindings from config to component
	input.KeyBindings = configBindings
}

// updateInputState polls the current keyboard state and updates the Input component.
func updateInputState(input *components.Input, gameTime float64) {
	if input == nil {
		return
	}

	// Update state for each logical key
	for logicalKey := 0; logicalKey < len(input.KeyBindings); logicalKey++ {
		updateLogicalKey(input, logicalKey, gameTime)
	}
}

// updateLogicalKey updates the input state for a single logical key.
func updateLogicalKey(input *components.Input, logicalKey int, gameTime float64) {
	physicalKeys := input.KeyBindings[logicalKey]
	if len(physicalKeys) == 0 {
		clearKeyState(input, logicalKey)
		return
	}

	down, pressed, released := pollPhysicalKeys(physicalKeys)

	input.KeyDown[logicalKey] = down
	input.KeyPressed[logicalKey] = pressed
	input.KeyReleased[logicalKey] = released

	// Automatically buffer key presses for better input feel
	if pressed {
		input.Buffer[logicalKey] = true
		input.BufferExpiry[logicalKey] = gameTime + inputBufferDuration
	}

	// Expire old buffered inputs
	if input.Buffer[logicalKey] && gameTime >= input.BufferExpiry[logicalKey] {
		input.Buffer[logicalKey] = false
	}
}

// clearKeyState clears all input state for a logical key.
func clearKeyState(input *components.Input, logicalKey int) {
	input.KeyDown[logicalKey] = false
	input.KeyPressed[logicalKey] = false
	input.KeyReleased[logicalKey] = false
}

// pollPhysicalKeys checks the state of all physical keys bound to a logical key.
// Returns (down, pressed, released) booleans.
func pollPhysicalKeys(physicalKeys []ebiten.Key) (bool, bool, bool) {
	down := false
	pressed := false
	released := false

	for _, physKey := range physicalKeys {
		if ebiten.IsKeyPressed(physKey) {
			down = true
		}
		if inpututil.IsKeyJustPressed(physKey) {
			pressed = true
		}
		if inpututil.IsKeyJustReleased(physKey) {
			released = true
		}
	}

	return down, pressed, released
}
