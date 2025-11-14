package world

import (
	"image/color"
	"math"

	"game/components"
	"game/ecs"
	"game/resources"
	debugworld "game/systems/draw/debug/world"

	"github.com/hajimehoshi/ebiten/v2"
)

type SpriteDrawState interface {
	CameraInFrameRecter(resources.Recter, float64, float64) bool
	CameraPosition() (float64, float64)
}

// isTransformInFrame checks if a Transform is visible within the camera frame
func isTransformInFrame(t *components.Transform, cx, cy float64, state SpriteDrawState) bool {
	if t == nil {
		return false
	}
	// Use ECS-native camera culling directly with Transform (implements camera.Recter)
	return state.CameraInFrameRecter(t, 0.1, 0.1)
}

// command represents a single draw submission with enough data to target screen or normal.
type command struct {
	img        *ebiten.Image
	geom       ebiten.GeoM
	layer      int
	colorScale color.Color
	toNormal   bool // when true, draw into normal map target using FillNormalMaskColorM
}

// toRenderCommand converts an internal command to a resources.RenderCommand for the ECS queue.
func (c command) toRenderCommand() resources.RenderCommand {
	target := resources.TargetScreen
	if c.toNormal {
		target = resources.TargetNormal
	}
	return resources.RenderCommand{
		Image:      c.img,
		TargetType: target,
		Layer:      c.layer,
		GeoM:       c.geom,
		ColorScale: c.colorScale,
	}
}

// pushToQueue submits commands to the ECS RenderQueue.
func pushToQueue(queue *resources.RenderQueue, cmds []command) {
	if queue == nil {
		return
	}
	for _, c := range cmds {
		queue.Push(c.toRenderCommand())
	}
}

func accumulateCounts(byLayerScreen, byLayerNormal map[int]int, cmds []command) {
	for _, cmd := range cmds {
		if cmd.toNormal {
			byLayerNormal[cmd.layer]++
			continue
		}
		byLayerScreen[cmd.layer]++
	}
}

// runSprites renders all active sprites using the ECS Transform+Render system.
// Sprites are rendered to the RenderQueue. This is the only sprite rendering path.
// Debug overlays draw directly to screen after sprites are rendered.
func runSprites(world *ecs.World, state SpriteDrawState, screen *ebiten.Image) (count int, byLayerScreen map[int]int, byLayerNormal map[int]int) {
	return runECSSprites(world, state, screen)
}

// runECSSprites uses ECS Transform+Render iteration for drawing.
// All rendering goes through the RenderQueue.
func runECSSprites(world *ecs.World, state SpriteDrawState, screen *ebiten.Image) (count int, byLayerScreen map[int]int, byLayerNormal map[int]int) {
	if world == nil || state == nil {
		return 0, nil, nil
	}

	// Get RenderQueue from ECS resources
	queue := ecs.Resource[resources.RenderQueue](world)
	if queue == nil {
		panic("RenderQueue not available - this should not happen in Phase 14.5+")
	}

	byLayerScreen = map[int]int{}
	byLayerNormal = map[int]int{}
	cx, cy := state.CameraPosition()
	var entityPos ebiten.GeoM

	// Draw entities with Transform+Render components
	for _, eid := range world.EntitiesWith((*components.Transform)(nil), (*components.Render)(nil)) {
		transform := ecs.GetComponent[components.Transform](world, eid)
		render := ecs.GetComponent[components.Render](world, eid)
		if transform == nil || render == nil {
			continue
		}

		// Camera culling using Transform
		if !isTransformInFrame(transform, cx, cy, state) {
			continue
		}

		// Build entity position matrix
		entityPos.Reset()
		entityPos.Translate(math.Ceil(transform.X-cx), math.Ceil(transform.Y-cy))

		// Draw the Render component
		cmds := buildECSRenderCommands(render, entityPos)

		// Push to render queue
		pushToQueue(queue, cmds)

		accumulateCounts(byLayerScreen, byLayerNormal, cmds)
		count++
	}

	return count, byLayerScreen, byLayerNormal
}

// buildECSRenderCommands coordinates rendering based on component flags.
// Delegates to specialized functions for animated vs static sprites.
func buildECSRenderCommands(c *components.Render, entityPos ebiten.GeoM) []command {
	if c == nil || c.Image == nil {
		return nil
	}

	w, h := imageSize(c.Image)

	if c.FromAnim {
		return buildAnimatedRenderCommands(c, w, h, entityPos)
	}
	return buildStaticRenderCommands(c, w, h, entityPos)
}

// buildAnimatedRenderCommands handles sprite sheet animations.
func buildAnimatedRenderCommands(c *components.Render, w, h float64, entityPos ebiten.GeoM) []command {
	geom := animGeom(c, w, h, entityPos)

	// Normal-only rendering
	if c.Normal {
		return []command{{
			img:      pickNormalImage(c),
			geom:     geom,
			layer:    c.Layer,
			toNormal: true,
		}}
	}

	// Screen rendering
	cs := defaultColor(c.ColorScale)
	cmds := []command{{
		img:        c.Image,
		geom:       geom,
		layer:      c.Layer,
		colorScale: cs,
		toNormal:   false,
	}}

	// Add normal map if fully opaque
	if alphaOf(cs) == 255 {
		cmds = append(cmds, command{
			img:      pickNormalImage(c),
			geom:     geom,
			layer:    c.Layer,
			toNormal: true,
		})
	}

	return cmds
}

// buildStaticRenderCommands handles non-animated sprites.
func buildStaticRenderCommands(c *components.Render, w, h float64, entityPos ebiten.GeoM) []command {
	geom := defaultGeom(c, w, h, entityPos)

	// Normal-only rendering
	if c.Normal {
		return []command{{
			img:      pickNormalImage(c),
			geom:     geom,
			layer:    c.Layer,
			toNormal: true,
		}}
	}

	// Both screen and normal rendering
	cs := defaultColor(c.ColorScale)
	return []command{
		{
			img:        c.Image,
			geom:       geom,
			layer:      c.Layer,
			colorScale: cs,
			toNormal:   false,
		},
		{
			img:      pickNormalImage(c),
			geom:     geom,
			layer:    c.Layer,
			toNormal: true,
		},
	}
}

// helpers to reduce complexity
func defaultColor(col color.Color) color.Color {
	if col == nil {
		return color.White
	}
	return col
}

func imageSize(img *ebiten.Image) (float64, float64) {
	b := img.Bounds().Size()
	return float64(b.X), float64(b.Y)
}

func pickNormalImage(c *components.Render) *ebiten.Image {
	if c.NormalImage != nil {
		return c.NormalImage
	}
	return c.Image
}

// alphaOf extracts the alpha channel from a color.Color as an 8-bit value.
func alphaOf(col color.Color) uint8 {
	if col == nil {
		return 255
	}
	r, g, b, a := col.RGBA()
	// RGBA returns 16-bit per channel; convert to 8-bit
	_ = r
	_ = g
	_ = b
	return uint8(a >> 8)
}

// runDebugOverlays draws debug overlays on top of composed sprites.
// Call this AFTER composeRenderQueue to ensure overlays appear on top of sprites.
// This renders body/hitbox/stats debug info for all entities with Transform components.
func runDebugOverlays(world *ecs.World, state SpriteDrawState, screen *ebiten.Image) {
	if world == nil || state == nil || screen == nil {
		return
	}

	cx, cy := state.CameraPosition()
	var entityPos ebiten.GeoM

	// Iterate all entities with Transform (same as sprite rendering)
	for _, eid := range world.EntitiesWith((*components.Transform)(nil)) {
		transform := ecs.GetComponent[components.Transform](world, eid)
		if transform == nil {
			continue
		}

		// Skip dead entities
		if health := ecs.GetComponent[components.Health](world, eid); health != nil && health.Current <= 0 {
			continue
		}

		// Camera culling
		if !isTransformInFrame(transform, cx, cy, state) {
			continue
		}

		// Build entity position matrix
		entityPos.Reset()
		entityPos.Translate(math.Ceil(transform.X-cx), math.Ceil(transform.Y-cy))

		// Draw hitbox debug overlays (collision boxes)
		// TODO: Implement hitbox debug rendering in systems/draw/hitbox_debug.go

		// FIXME: Relocate to debug packages
		// Draw stats debug text (health/stamina/poise numbers) - Pure ECS version
		debugworld.DrawStatsDebugText(world, screen, eid, entityPos)

		// Draw entity ID debug text (centered on entity)
		debugState := ecs.Resource[resources.DebugState](world)
		if debugState != nil && debugState.IsEnabled(resources.DebugCategoryEntityID) {
			debugworld.DrawEntityIDText(screen, eid, transform, entityPos)
		}

		// AI debug rendering is now handled by systems/draw/debug/world_space.go
		// No longer called here - debug systems handle all AI visualization
	}

	// Draw debug overlays (AI target ranges, sensor boxes, etc.)
	// This is called after the entity loop so overlays appear on top of all entities
	drawDebugOverlaysIfEnabled(world, screen)
}

// drawDebugOverlaysIfEnabled is a placeholder for the parent draw package function.
// Debug overlay implementations are now centralized in systems/draw/debug/world_space.go
func drawDebugOverlaysIfEnabled(world *ecs.World, screen *ebiten.Image) {
	// Placeholder - debug systems are called from systems/draw/debug/
	_ = world
	_ = screen
}
