package visual

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Render component holds data for rendering an entity's sprite.
type Render struct {
	Image *ebiten.Image
	// NormalImage, if set, will be drawn to the normal map target instead of reusing Image.
	// Use this for tiles/sprites that have a dedicated normal texture.
	NormalImage  *ebiten.Image
	X, Y         float64
	R            float64
	FlipX, FlipY bool
	Layer        int
	ColorScale   color.Color
	Normal       bool
	// When true, draw using anim-style geometry (no center-rotate; offsets mirror anim OX/OY and flip deltas).
	FromAnim bool

	// Optional: auto-rotation at a fixed cadence, used by some effects.
	RollingTime  time.Duration
	rollingTimer *time.Timer
}
