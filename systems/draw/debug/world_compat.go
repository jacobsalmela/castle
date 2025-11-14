//go:build !release

// This file provides backward compatibility for the refactored debug system.
// The world-space debug functions have been moved to the debug/world subpackage.
// This file re-exports them for any existing code that imports from the debug package.

package debug

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/tilemap"
	"game/systems/draw/debug/primitives"
	"game/systems/draw/debug/world"

	"github.com/hajimehoshi/ebiten/v2"
)

// CameraProvider is re-exported from primitives for backward compatibility.
type CameraProvider = primitives.CameraProvider

// Re-export all world-space debug functions for backward compatibility

func DrawCollisionDebug(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawCollisionDebug(w, screen, cam)
}

func DrawBodyDebug(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawBodyDebug(w, screen, cam)
}

func DrawPhysicsDebug(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawPhysicsDebug(w, screen, cam)
}

func DrawHitboxDebug(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawHitboxDebug(w, screen, cam)
}

func DrawAIDebug(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawAIDebug(w, screen, cam)
}

func DrawStatsDebug(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawStatsDebug(w, screen, cam)
}

func DrawAnimDebug(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawAnimDebug(w, screen, cam)
}

func DrawEntityIDDebug(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawEntityIDDebug(w, screen, cam)
}

func DrawBehaviorTreeDebug(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawBehaviorTreeDebug(w, screen, cam)
}

func DrawVisualPrimitives(w *ecs.World, screen *ebiten.Image, cam CameraProvider) {
	world.DrawVisualPrimitives(w, screen, cam)
}

func DrawTileDebugOverlay(w *ecs.World, tiledMap *tilemap.Map, cam CameraProvider, screen *ebiten.Image, vp *components.ViewPort) {
	world.DrawTileDebugOverlay(w, tiledMap, cam, screen, vp)
}

func DrawLadderDebugOverlay(w *ecs.World, cam CameraProvider, screen *ebiten.Image, vp *components.ViewPort) {
	world.DrawLadderDebugOverlay(w, cam, screen, vp)
}

func DrawStatsDebugText(w *ecs.World, screen *ebiten.Image, eid entities.EntityId, entityPos ebiten.GeoM) {
	world.DrawStatsDebugText(w, screen, eid, entityPos)
}

func DrawEntityIDText(screen *ebiten.Image, eid entities.EntityId, transform *components.Transform, entityPos ebiten.GeoM) {
	world.DrawEntityIDText(screen, eid, transform, entityPos)
}

func DrawPlayerFrameDebug(w *ecs.World, screen *ebiten.Image, eid entities.EntityId, entityPos ebiten.GeoM) {
	world.DrawPlayerFrameDebug(w, screen, eid, entityPos)
}
