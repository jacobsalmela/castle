// Package renderer handles map rendering to screen and queues.
package renderer

import (
	"log"
	"math"

	"game/pkg/tilemap/types"

	"github.com/hajimehoshi/ebiten/v2"
)

// Target constants matching resources.RenderTarget
const (
	TargetScreen = 1 << 0 // resources.TargetScreen
	TargetNormal = 1 << 1 // resources.TargetNormal
)

// LayerIndex is the base z-depth for map layers.
var LayerIndex = 2

// Update advances all animation timers.
func Update(m *types.Map, dt float64) {
	for _, layers := range m.Layers {
		for _, layer := range layers {
			updateLayerAnimations(layer, dt)
		}
	}
}

// updateLayerAnimations updates all animations in a layer.
func updateLayerAnimations(layer *types.LayerData, dt float64) {
	for _, anim := range layer.Animations {
		anim.Timer += dt
		if anim.Timer >= anim.Frames[anim.Current].Duration {
			anim.Current = (anim.Current + 1) % len(anim.Frames)
			anim.Timer = 0
		}
	}
}

// DrawToQueue renders map layers to the render queue (Pure ECS rendering).
func DrawToQueue(m *types.Map, queue types.MapRenderQueue, camera types.Camera) {
	if queue == nil || camera == nil {
		return
	}

	renderObjectLayersToQueue(m, queue, camera)
	renderTileLayersToQueue(m, queue, camera)
}

// renderObjectLayersToQueue renders parallax backgrounds to the queue.
func renderObjectLayersToQueue(m *types.Map, queue types.MapRenderQueue, camera types.Camera) {
	for _, layer := range m.ObjectLayers {
		if layer.Image == nil {
			continue
		}

		bounds := camera.BoundsWithOffsetAndParallax(layer.OffsetX, layer.OffsetY, layer.ParallaxX, layer.ParallaxY)
		localImage, ok := layer.Image.SubImage(bounds).(*ebiten.Image)
		if !ok || localImage == nil {
			continue
		}

		// Object layers render to both screen and normal map for lighting
		queue.PushMapTile(localImage, TargetScreen|TargetNormal, -LayerIndex, ebiten.GeoM{})
	}
}

// renderTileLayersToQueue renders tile layers and animations to the queue.
func renderTileLayersToQueue(m *types.Map, queue types.MapRenderQueue, camera types.Camera) {
	for imageTag, layers := range m.Layers {
		targetType := getTargetType(imageTag)

		for i, layer := range layers {
			layerDepth := calculateLayerDepth(i, m.BackgroundLayersNum)

			if layer.Image == nil {
				log.Printf("WARNING: Map layer %d (%s) has nil image, skipping", i, imageTag)
				continue
			}

			renderStaticLayer(queue, camera, layer, targetType, layerDepth)
			renderLayerAnimations(m, queue, camera, layer, targetType, layerDepth)
		}
	}
}

// getTargetType converts an image tag to a render target type.
func getTargetType(imageTag string) int {
	if imageTag == "normal" {
		return TargetNormal
	}
	return TargetScreen
}

// calculateLayerDepth determines draw order depth for a layer.
func calculateLayerDepth(layerIndex, backgroundLayersNum int) int {
	if layerIndex >= backgroundLayersNum {
		return LayerIndex
	}
	return -LayerIndex
}

// renderStaticLayer renders a single static tile layer.
func renderStaticLayer(queue types.MapRenderQueue, camera types.Camera, layer *types.LayerData, targetType, layerDepth int) {
	bounds := camera.BoundsWithOffsetAndParallax(layer.OffsetX, layer.OffsetY, layer.ParallaxX, layer.ParallaxY)
	localImage, ok := layer.Image.SubImage(bounds).(*ebiten.Image)
	if !ok || localImage == nil {
		return
	}

	queue.PushMapTile(localImage, targetType, layerDepth, ebiten.GeoM{})
}

// renderLayerAnimations renders all animated tiles in a layer.
func renderLayerAnimations(m *types.Map, queue types.MapRenderQueue, camera types.Camera, layer *types.LayerData, targetType, layerDepth int) {
	cx, cy := camera.Position()

	for _, anim := range layer.Animations {
		for _, pos := range anim.Positions {
			geom := buildAnimationTransform(m, pos, cx, cy)
			queue.PushMapTile(anim.Frames[anim.Current].Image, targetType, layerDepth, geom)
		}
	}
}

// buildAnimationTransform creates the geometry transform for an animated tile.
func buildAnimationTransform(m *types.Map, pos types.AnimationPosition, cx, cy float64) ebiten.GeoM {
	var geom ebiten.GeoM
	sx, sy, dx, dy := 1.0, 1.0, 0.0, 0.0

	// Handle tile flipping
	if pos.FlipR {
		geom.Rotate(math.Pi / 2)
		sx = -1
	}
	if pos.FlipX {
		sx, dx = -1, float64(m.Data.TileWidth)
		if pos.FlipR {
			sx = 1
		}
	}
	if pos.FlipY {
		sy, dy = -1, float64(m.Data.TileHeight)
	}

	// Apply transforms
	posX := pos.X + dx
	posY := pos.Y + dy
	geom.Scale(sx, sy)
	geom.Translate(math.Ceil(posX-cx), math.Ceil(posY-cy))

	return geom
}
