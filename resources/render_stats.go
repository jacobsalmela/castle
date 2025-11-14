package resources

// RenderStats is an ECS resource that tracks rendering statistics.
// This replaces vars.ECSSpritesDrawn with proper ECS resource management.
type RenderStats struct {
	SpritesDrawn int
	DrawCalls    int
}

// Reset clears all statistics for a new frame.
func (r *RenderStats) Reset() {
	if r != nil {
		r.SpritesDrawn = 0
		r.DrawCalls = 0
	}
}

// IncrementSprites increments the sprites drawn counter.
func (r *RenderStats) IncrementSprites(count int) {
	if r != nil {
		r.SpritesDrawn += count
	}
}

// IncrementDrawCalls increments the draw calls counter.
func (r *RenderStats) IncrementDrawCalls() {
	if r != nil {
		r.DrawCalls++
	}
}
