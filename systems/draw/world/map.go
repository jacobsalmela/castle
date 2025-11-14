package world

import (
	"game/pkg/tilemap"
	"game/resources"
)

// renderMap renders map tiles to the render queue.
//
// Pure ECS Queue-Based Rendering:
// - Map renders to queue via tilemap.DrawToQueue()
// - Uses RenderQueue.PushMapTile() to avoid import cycles
// - Proper layer ordering: map tiles → sprites → UI (all in queue)
//
// Implementation:
//  1. tilemap.MapRenderQueue interface avoids import cycles
//  2. RenderQueue.PushMapTile() satisfies the interface
//  3. tilemap.DrawToQueue() function handles map rendering
//  4. Layer ordering: background (negative) → sprites (0) → foreground (positive) → UI (10)
func renderMap(queue *resources.RenderQueue, tiledMap *tilemap.Map, cam *resources.Camera) {
	if queue == nil || tiledMap == nil || cam == nil {
		return
	}

	// Use the new DrawToQueue function
	tilemap.DrawToQueue(tiledMap, queue, cam)
}
