//go:build !release

package debug

import (
	"game/ecs"
	"game/pkg/perfmon"
	"game/resources"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// drawProfiler renders the performance profiler overlay in the bottom-right corner.
// This is the Pure ECS replacement for the profiler section in game_draw.go.
//
// Shows (in bottom-right corner):
// - Performance stats (FPS, update/draw times, memory)
// - Top 3 slowest systems
//
// Enable/disable with Cmd+Shift+0 debug toggle.
func drawProfiler(screen *ebiten.Image) {
	if screen == nil {
		return
	}

	// Draw performance stats in bottom-right corner
	perfStats := perfmon.GlobalPerfMonitor.GetStats()
	screenBounds := screen.Bounds()

	// Position text in bottom-right, accounting for text size
	perfX := screenBounds.Dx() - 450
	perfY := screenBounds.Dy() - 85 // Move up to make room for profiler
	ebitenutil.DebugPrintAt(screen, perfStats, perfX, perfY)

	// Draw top 3 slowest systems
	slowest := perfmon.GlobalProfiler.GetTopSlowest(3)
	if len(slowest) > 0 {
		profX := screenBounds.Dx() - 450
		profY := screenBounds.Dy() - 50
		for i, sys := range slowest {
			ebitenutil.DebugPrintAt(screen, sys, profX, profY+i*15)
		}
	}
}

// drawProfilerIfEnabled checks the flag and renders the profiler if enabled.
func drawProfilerIfEnabled(world *ecs.World, screen *ebiten.Image) {
	debugState := ecs.Resource[resources.DebugState](world)
	if debugState != nil && debugState.IsEnabled(resources.DebugCategoryProfiler) {
		drawProfiler(screen)
	}
}
