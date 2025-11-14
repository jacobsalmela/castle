//go:build !release

// Package world provides world-space debug rendering systems.
// All debug visualizations that render in world coordinates (logical screen space).
package world

import (
	"game/components"
	"game/ecs"
	"game/pkg/tilemap"
	"game/systems/draw/debug/primitives"

	"github.com/hajimehoshi/ebiten/v2"
)

// Update renders all world-space debug visualizations.
// This is called from the main debug orchestrator.
func Update(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider, tiledMap *tilemap.Map, vp *components.ViewPort) {
	// Draw collision debug (Cmd+1)
	DrawCollisionDebug(world, screen, cam)

	// Draw physics debug (Cmd+2)
	DrawBodyDebug(world, screen, cam)
	DrawPhysicsDebug(world, screen, cam)

	// Draw hitbox debug (Cmd+3)
	DrawHitboxDebug(world, screen, cam)

	// Draw AI debug (Cmd+4)
	DrawAIDebug(world, screen, cam)

	// Draw behavior tree debug (Cmd+7)
	DrawBehaviorTreeDebug(world, screen, cam)

	// Draw stats debug (Cmd+5)
	DrawStatsDebug(world, screen, cam)

	// Draw animation debug (Cmd+6)
	DrawAnimDebug(world, screen, cam)

	// Draw entity ID debug (Shift+E)
	DrawEntityIDDebug(world, screen, cam)

	// Draw visual primitives (DebugVisual components)
	DrawVisualPrimitives(world, screen, cam)

	// Draw tile debug (Shift+T)
	DrawTileDebugOverlay(world, tiledMap, cam, screen, vp)

	// Draw ladder debug (Shift+D)
	DrawLadderDebugOverlay(world, cam, screen, vp)
}
