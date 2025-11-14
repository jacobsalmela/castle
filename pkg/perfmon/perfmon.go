package perfmon

import (
	"fmt"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// PerfMonitor tracks performance metrics for debugging
type PerfMonitor struct {
	frameCount    int
	lastTime      time.Time
	currentFPS    float64
	currentTPS    float64
	memStats      runtime.MemStats
	lastMemUpdate time.Time
}

var GlobalPerfMonitor = &PerfMonitor{
	lastTime:      time.Now(),
	lastMemUpdate: time.Now(),
}

// Update should be called once per frame
func (pm *PerfMonitor) Update() {
	pm.frameCount++
	now := time.Now()
	elapsed := now.Sub(pm.lastTime).Seconds()

	// Update FPS every second
	if elapsed >= 1.0 {
		pm.currentFPS = float64(pm.frameCount) / elapsed
		pm.currentTPS = ebiten.ActualTPS()
		pm.frameCount = 0
		pm.lastTime = now

		// Update memory stats ONLY once per second (not every 2 seconds)
		// and only when we're already updating FPS to avoid extra overhead
		runtime.ReadMemStats(&pm.memStats)
		pm.lastMemUpdate = now
	}
}

// GetStats returns formatted performance statistics
func (pm *PerfMonitor) GetStats() string {
	return fmt.Sprintf(
		"FPS: %.1f | TPS: %.1f | Alloc: %.1f MB | Sys: %.1f MB | GC: %d",
		pm.currentFPS,
		pm.currentTPS,
		float64(pm.memStats.Alloc)/1024/1024,
		float64(pm.memStats.Sys)/1024/1024,
		pm.memStats.NumGC,
	)
}

// FPS returns the current frames per second
func (pm *PerfMonitor) FPS() float64 {
	return pm.currentFPS
}

// TPS returns the current ticks per second
func (pm *PerfMonitor) TPS() float64 {
	return pm.currentTPS
}

// AllocMB returns allocated memory in megabytes
func (pm *PerfMonitor) AllocMB() float64 {
	return float64(pm.memStats.Alloc) / 1024 / 1024
}
