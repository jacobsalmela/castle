package buffers

import (
	"game/components"
	"game/ecs"

	"github.com/hajimehoshi/ebiten/v2"
)

// ═══════════════════════════════════════════════════════════════════════════════
// PHASE 1: BUFFER MANAGEMENT
// Create and manage rendering buffers (logical screen, normal map).
// ═══════════════════════════════════════════════════════════════════════════════

// Update manages all rendering buffers for the current frame.
// Returns the logical screen and normal map buffers, or nil if viewport is missing.
//
// Buffers:
//   - Logical screen: Main render target (scaled to device screen later)
//   - Normal map: Used by lighting shader for directional lighting
//
// Parameters:
//   - world: ECS world instance
//   - vp: Viewport with logical dimensions
//
// Returns:
//   - logicalScreen: Main render target
//   - normalMap: Normal map for lighting
func Update(world *ecs.World, vp *components.ViewPort) (logicalScreen, normalMap *ebiten.Image) {
	if world == nil || vp == nil {
		return nil, nil
	}

	logicalScreen = getLogicalScreen(world, vp)
	normalMap = getNormalMap(vp)

	// Clear logical buffer for this frame
	if logicalScreen != nil {
		logicalScreen.Clear()
	}

	return logicalScreen, normalMap
}

// getLogicalScreen returns or creates the logical screen buffer.
// Stored as a resource to persist between frames.
func getLogicalScreen(world *ecs.World, vp *components.ViewPort) *ebiten.Image {
	// Try to get existing buffer from resources
	type LogicalScreenResource struct {
		Image *ebiten.Image
	}

	lsr := ecs.Resource[LogicalScreenResource](world)
	if lsr != nil && lsr.Image != nil {
		// Check if size matches viewport
		if lsr.Image.Bounds().Dx() == int(vp.LW) && lsr.Image.Bounds().Dy() == int(vp.LH) {
			return lsr.Image
		}
		// Size changed, deallocate old buffer
		lsr.Image.Deallocate()
	}

	// Create new buffer
	img := ebiten.NewImage(int(vp.LW), int(vp.LH))
	world.SetResource(&LogicalScreenResource{Image: img})
	return img
}

// getNormalMap returns or creates the normal map buffer.
// This is a module-level cache since it's not tied to world state.
var normalMapCache *ebiten.Image

func getNormalMap(vp *components.ViewPort) *ebiten.Image {
	if normalMapCache == nil ||
		normalMapCache.Bounds().Dx() != int(vp.LW) ||
		normalMapCache.Bounds().Dy() != int(vp.LH) {
		if normalMapCache != nil {
			normalMapCache.Deallocate()
		}
		normalMapCache = ebiten.NewImage(int(vp.LW), int(vp.LH))
	}
	return normalMapCache
}
