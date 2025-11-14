package world

import (
	"game/components"
	"game/ecs"
	"game/pkg/config"
	"game/pkg/tilemap"
	"game/resources"
	debugworld "game/systems/draw/debug/world"

	"github.com/hajimehoshi/ebiten/v2"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PHASE 2: WORLD RENDERING
// Render map tiles, sprites, and compose them to screen and normal map.
// ═══════════════════════════════════════════════════════════════════════════════

// Update renders the complete game world to the logical screen and normal map.
// This is a standalone package that orchestrates world rendering without importing parent.
//
// Rendering Pipeline:
//  1. Clear normal map
//  2. Render map tiles to queue
//  3. Render sprites to queue
//  4. Compose queue to screen + normal map (sorted by layer)
//  5. Draw debug overlays
//  6. Render UI elements (headbars, textboxes) to queue
//  7. Draw damage numbers (floating text VFX)
//
// Parameters:
//   - world: ECS world instance
//   - screen: Logical screen buffer
//   - normalMap: Normal map buffer for lighting
//   - vp: Viewport for tile debug overlay
//
// Note: This package should be self-contained. Import map.go, sprites.go, compose.go, etc.
// into this package instead of referencing parent systems/draw.
func Update(world *ecs.World, screen, normalMap *ebiten.Image, vp *components.ViewPort) {
	if world == nil || screen == nil || normalMap == nil {
		return
	}

	// Get resources at start for efficient access
	queue := ecs.Resource[resources.RenderQueue](world)
	if queue == nil {
		panic("RenderQueue not available - this should not happen in Phase 14.5+")
	}
	camera := ecs.Resource[resources.Camera](world)
	mapRef := ecs.Resource[resources.MapRef](world)
	var worldMap *tilemap.Map
	if mapRef != nil {
		worldMap = mapRef.Map
	}

	// === PHASE 2.1: CLEAR NORMAL MAP ===
	normalMap.Fill(components.NormalMaskColor)

	// === PHASE 2.2: RENDER MAP ===
	// Map tiles render to queue at appropriate layers
	if worldMap != nil && camera != nil {
		renderMap(queue, worldMap, camera)
	}

	// === PHASE 2.3: RENDER SPRITES ===
	// All entities with Transform+Render components
	renderStats := ecs.Resource[resources.RenderStats](world)
	spritesDrawn, _, _ := runSprites(world, world, screen)
	if renderStats != nil {
		renderStats.SpritesDrawn = spritesDrawn
	}

	// === PHASE 2.4: COMPOSE TO SCREEN + NORMAL MAP ===
	// Drain queue and render all commands (map + sprites)
	ComposeRenderQueue(screen, normalMap, queue)

	// === PHASE 2.5: PER-ENTITY DEBUG OVERLAYS ===
	// Draw stats and entity ID debug text on top of sprites
	runDebugOverlays(world, world, screen)

	// Get config from ECS world for debug checks
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// === PHASE 2.6: PURE ECS DEBUG SYSTEMS ===
	// All debug systems check their own enabled flags internally
	if cfg.Debug {
		debugworld.DrawCollisionDebug(world, screen, camera)
		debugworld.DrawBodyDebug(world, screen, camera)
		debugworld.DrawPhysicsDebug(world, screen, camera)
		debugworld.DrawHitboxDebug(world, screen, camera)
		debugworld.DrawAIDebug(world, screen, camera)
		debugworld.DrawBehaviorTreeDebug(world, screen, camera)
		debugworld.DrawStatsDebug(world, screen, camera)
		debugworld.DrawAnimDebug(world, screen, camera)
		debugworld.DrawEntityIDDebug(world, screen, camera)
		debugworld.DrawVisualPrimitives(world, screen, camera)
	}

	// === PHASE 2.7: TILE DEBUG OVERLAY ===
	// Tile coordinates and grid (if enabled)
	if cfg.Debug && vp != nil && worldMap != nil && camera != nil {
		debugworld.DrawTileDebugOverlay(world, worldMap, camera, screen, vp)
		debugworld.DrawLadderDebugOverlay(world, camera, screen, vp)
	}

	// === PHASE 2.8: DAMAGE NUMBERS ===
	// Floating damage number VFX (rendered after all world content)
	if camera != nil {
		DrawDamageNumbers(world, screen, camera)
	}

	// Note: UI elements (headbars, textboxes) are now rendered in parent draw.go
	// as part of the world rendering phase, not in the UI package phase.
	// This is because they render in world space and need to be composed with sprites.
}
