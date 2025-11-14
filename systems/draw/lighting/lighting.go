package lighting

import (
	_ "embed" // Embed is used to embed the shader file.
	"game/components"
	"game/ecs"
	"game/pkg/config"
	"game/pkg/tilemap"
	"game/resources"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PHASE 3: LIGHTING
// Apply lighting shader to create dynamic lighting effects.
// ═══════════════════════════════════════════════════════════════════════════════

// Constants removed - now using config values:
//   - Light radius: config.Cfg.Screen.LightRadius (default 96)
//   - Ambient brightness: config.Cfg.Screen.AmbientBrightness (default 80)

var (
	// Lazy-initialized images - created on first Update() call
	fallbackNormalMap *ebiten.Image
	diffuseImage      *ebiten.Image
	imagesInitialized bool

	lights         []lightSource
	shaderTime     float32
	lightingShader *ebiten.Shader
	//go:embed light.kage
	shaderData []byte
)

type lightSource struct {
	x, y          float64
	size          float64 // If 0, will use config.Cfg.Lighting.Sources["torch"].Radius at render time
	intensity     float64 // Brightness multiplier (1.0 = normal, 2.0 = twice as bright)
	useConfigSize bool    // If true, read radius from config each frame
}

// getFloatOrFallback returns primary value if non-zero, otherwise fallback.
// Used for backwards compatibility when migrating from Screen config to Lighting config.
func getFloatOrFallback(primary, fallback float32) float32 {
	if primary != 0 {
		return primary
	}
	return fallback
}

// initLightingImages creates the lighting buffers on first use.
// This is called lazily to avoid using config values at package init time.
func initLightingImages(cfg *config.Config) {
	if imagesInitialized {
		return
	}

	width := int(cfg.Screen.Width)
	height := int(cfg.Screen.Height)

	fallbackNormalMap = ebiten.NewImage(width, height)
	fallbackNormalMap.Fill(color.RGBA{128, 128, 255, 255})

	diffuseImage = ebiten.NewImage(width, height)

	imagesInitialized = true
}

// LoadLighting initializes the lighting shader and extracts light positions from the map.
// Called from game package during map load.
func LoadLighting(worldMap *tilemap.Map, lightTileGID uint32) {
	// Always load the shader and lights, regardless of current toggle state
	// This allows runtime toggle with L key
	var err error
	if lightingShader, err = ebiten.NewShader(shaderData); err != nil {
		log.Fatal(err)
	}
	positions := tilemap.FindTilePosition(worldMap, lightTileGID)
	lights = make([]lightSource, len(positions)+1)
	lights[len(lights)-1] = lightSource{x: 0, y: 0, size: 0, intensity: 1.0, useConfigSize: false}

	// Center lights on 8x8 torch tiles (upstream uses +4, +4 for centering)
	const torchTileHalfSize = 4.0

	for i, pos := range positions {
		// Store position but mark to use config radius at render time for hot-reload
		lights[i] = lightSource{
			x:             pos[0] + torchTileHalfSize,
			y:             pos[1] + torchTileHalfSize,
			size:          0,   // Will use config value
			intensity:     1.0, // Default brightness (can be overridden per-light later)
			useConfigSize: true,
		}
	}
}

// UpdateTime advances the lighting shader animation time.
func UpdateTime(dt float64) {
	shaderTime += float32(dt)
}

// Update renders the lighting shader to the screen.
// This composites the normal map with light sources to create dynamic lighting.
// A minimum ambient brightness is maintained so rooms without lights remain visible.
//
// The lighting can be toggled with the L key (DebugCategoryLighting).
// When the debug toggle is enabled, it DISABLES lighting (turns it off).
//
// Logic:
//   - Config true + debug off = Lighting ON
//   - Config true + debug on  = Lighting OFF (L key pressed)
//   - Config false + debug off = Lighting OFF
//   - Config false + debug on = Lighting ON (L key pressed to override config)
//
// Light sources come from two sources:
//  1. Static map-based lights (from LoadLighting)
//  2. Dynamic ECS entities with LightEmitter component
//
// Parameters:
//   - world: ECS world instance (for camera position and debug state)
//   - screen: Logical screen buffer to apply lighting to
//   - normalMap: Normal map buffer (falls back to a flat map when nil)
func Update(world *ecs.World, screen, normalMap *ebiten.Image) {
	if screen == nil {
		return
	}

	// Get config from world resource
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Initialize lighting images on first use (lazy initialization)
	initLightingImages(cfg)

	// Check if lighting is enabled via debug toggle
	// Debug toggle acts as XOR - it flips the config value
	debugState := ecs.Resource[resources.DebugState](world)

	// Use the new config.Lighting.Enabled as the authoritative setting
	// (config.Screen.Lighting is deprecated but kept for reference)
	lightingEnabled := cfg.Lighting.Enabled

	// If debug category is enabled, flip the lighting state
	if debugState != nil && debugState.IsEnabled(resources.DebugCategoryLighting) {
		lightingEnabled = !lightingEnabled
	}

	if !lightingEnabled {
		return
	}

	// Get camera position from world
	var cx, cy float64
	camera := ecs.Resource[resources.Camera](world)
	if camera != nil {
		cx, cy = camera.Position()
	}

	// Start with pure black for lighting buffer (upstream behavior)
	// This creates high-contrast lighting where only lit areas are visible
	// The ambient brightness is applied via the shader's internal ambient value (0.5)
	// instead of pre-filling the buffer with gray
	diffuseImage.Fill(color.Black)

	// Determine normal map source for shader input
	normalSource := normalMap
	if normalSource == nil {
		normalSource = fallbackNormalMap
	}

	// Build combined light list from static map lights and dynamic ECS emitters
	allLights := buildLightList(world)

	// Render each light source
	for i, light := range allLights {
		// Convert to screen-space coordinates (relative to camera)
		x, y := light.x-cx, light.y-cy
		w, h := float64(cfg.Screen.Width), float64(cfg.Screen.Height)

		// Skip lights that are far off-screen (optimization)
		if x < -2*w || y < -2*h || x > 3*w || y > 3*h {
			continue
		}

		// Determine light size: use config if flagged, otherwise use stored size
		lightSize := light.size
		if light.useConfigSize {
			// Use "torch" light source from config (fallback to old Screen.LightRadius for compatibility)
			torchSource := cfg.Lighting.GetLightSource("torch")
			lightSize = torchSource.Radius
			light.intensity = torchSource.Intensity

			// Fallback to old config if new lighting config not set
			if lightSize == 0 {
				lightSize = cfg.Screen.LightRadius
			}
			if light.intensity == 0 {
				light.intensity = 1.0
			}
		}

		op := &ebiten.DrawRectShaderOptions{
			Uniforms: map[string]any{
				"LightPosSize":          []float32{float32(x), float32(y), float32(lightSize)},
				"LightIntensity":        float32(light.intensity),
				"Time":                  shaderTime + float32(10*i),
				"AmbientBrightness":     getFloatOrFallback(cfg.Lighting.AmbientBrightness, cfg.Screen.AmbientBrightness),
				"FlickerAmount":         getFloatOrFallback(cfg.Lighting.FlickerAmount, cfg.Screen.LightFlickerAmount),
				"FlickerSpeed":          getFloatOrFallback(cfg.Lighting.FlickerSpeed, cfg.Screen.LightFlickerSpeed),
				"LightResolution":       getFloatOrFallback(cfg.Lighting.Resolution, cfg.Screen.LightResolution),
				"LightResolutionOffset": getFloatOrFallback(cfg.Lighting.ResolutionOffset, cfg.Screen.LightResolutionOffset),
				"DitherIntensity":       getFloatOrFallback(cfg.Lighting.DitherIntensity, cfg.Screen.LightDitherIntensity),
			},
			Images: [4]*ebiten.Image{normalSource},
			Blend:  ebiten.BlendLighter,
		}
		op.Blend.BlendOperationRGB = ebiten.BlendOperationMax
		op.Blend.BlendOperationAlpha = ebiten.BlendOperationMax
		diffuseImage.DrawRectShader(int(cfg.Screen.Width), int(cfg.Screen.Height), lightingShader, op)
	}

	// Always apply lighting composite with ambient minimum
	// Multiply mode: screen_color * diffuse_brightness
	// With ambient at config.AmbientBrightness/255, darkest areas remain at that brightness level
	screen.DrawImage(diffuseImage, &ebiten.DrawImageOptions{CompositeMode: ebiten.CompositeModeMultiply})

	// DEBUG: Draw light origin markers (enable when debugging light positions)
	if cfg.Debug {
		drawLightDebugMarkers(world, screen, allLights, cx, cy, cfg)
	}
}

// drawLightDebugMarkers draws small circles at each light source origin for debugging.
// Enable by setting debug: true in config.yml
func drawLightDebugMarkers(world *ecs.World, screen *ebiten.Image, lights []lightSource, cx, cy float64, cfg *config.Config) {
	if screen == nil {
		return
	}

	// Draw a small red circle at each light position
	for _, light := range lights {
		x, y := light.x-cx, light.y-cy

		// Skip off-screen lights
		if x < -20 || y < -20 || x > float64(cfg.Screen.Width)+20 || y > float64(cfg.Screen.Height)+20 {
			continue
		}

		// Draw a 3x3 red square at the light origin
		markerSize := 3.0
		markerImg := ebiten.NewImage(int(markerSize), int(markerSize))
		markerImg.Fill(color.RGBA{255, 0, 0, 255}) // Bright red

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x-markerSize/2, y-markerSize/2) // Center the marker
		screen.DrawImage(markerImg, op)
	}
}

// buildLightList combines static map lights with dynamic ECS LightEmitter entities.
// This allows both background tile torches and interactive torch entities to emit light.
func buildLightList(world *ecs.World) []lightSource {
	// Start with static map lights (from LoadLighting)
	lightList := make([]lightSource, 0, len(lights)+16)
	lightList = append(lightList, lights...)

	// Add dynamic ECS light emitters
	if world != nil {
		entities := world.EntitiesWith((*components.LightEmitter)(nil), (*components.Transform)(nil))
		for _, eid := range entities {
			emitter := ecs.GetComponent[components.LightEmitter](world, eid)
			transform := ecs.GetComponent[components.Transform](world, eid)

			// Skip inactive emitters
			if emitter == nil || !emitter.Active || transform == nil {
				continue
			}

			// Create light source at entity position (centered)
			lightList = append(lightList, lightSource{
				x:             transform.X + transform.W/2,
				y:             transform.Y + transform.H/2,
				size:          emitter.Radius,
				intensity:     emitter.Intensity,
				useConfigSize: false, // ECS entities use their own radius
			})
		}
	}

	return lightList
}
