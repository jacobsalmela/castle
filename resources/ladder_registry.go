package resources

import (
	"game/pkg/bump"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// LadderRegistry stores all ladder positions from the tilemap.
// Ladders are vertical climbable objects that:
// - Allow player to walk through horizontally
// - Can be stood on top of (when player is above)
// - Can be climbed with UP/DOWN keys (when overlapping)
// - Can be descended through with DOWN key (when standing on top)
type LadderRegistry struct {
	Ladders []bump.Rect // All ladder rectangles in world coordinates
	Debug   bool        // Toggle with Shift+E to visualize ladders
}

// NewLadderRegistry creates an empty ladder registry.
func NewLadderRegistry() *LadderRegistry {
	return &LadderRegistry{
		Ladders: make([]bump.Rect, 0),
		Debug:   false,
	}
}

// AddLadder registers a ladder rectangle.
func (lr *LadderRegistry) AddLadder(rect bump.Rect) {
	lr.Ladders = append(lr.Ladders, rect)
}

// Clear removes all ladders.
func (lr *LadderRegistry) Clear() {
	lr.Ladders = lr.Ladders[:0]
}

// FindOverlapping returns all ladders that overlap with the given rectangle.
func (lr *LadderRegistry) FindOverlapping(rect bump.Rect) []bump.Rect {
	var overlapping []bump.Rect
	for _, ladder := range lr.Ladders {
		if bump.Overlaps(rect, ladder) {
			overlapping = append(overlapping, ladder)
		}
	}
	return overlapping
}

// IsOnTopOf checks if rect is standing on top of any ladder.
// "On top" means rect's bottom edge is within tolerance of ladder's top edge.
func (lr *LadderRegistry) IsOnTopOf(rect bump.Rect, tolerance float64) (bool, bump.Rect) {
	rectBottom := rect.Y + rect.H
	for _, ladder := range lr.Ladders {
		ladderTop := ladder.Y
		// Check if player bottom is near ladder top (within tolerance)
		// and horizontally overlaps with ladder
		if rectBottom >= ladderTop-tolerance && rectBottom <= ladderTop+tolerance {
			// Check horizontal overlap
			if rect.X+rect.W > ladder.X && rect.X < ladder.X+ladder.W {
				return true, ladder
			}
		}
	}
	return false, bump.Rect{}
}

// DrawDebug renders ladder rectangles for debugging (called with Shift+E).
func (lr *LadderRegistry) DrawDebug(screen *ebiten.Image, camera Camera) {
	if !lr.Debug {
		return
	}

	camX, camY := camera.Position()

	for _, ladder := range lr.Ladders {
		// Convert world coordinates to screen coordinates
		screenX := float32(ladder.X - camX)
		screenY := float32(ladder.Y - camY)
		w := float32(ladder.W)
		h := float32(ladder.H)

		// Draw filled rectangle with transparency
		vector.DrawFilledRect(screen, screenX, screenY, w, h, color.RGBA{0, 255, 0, 50}, false)

		// Draw outline
		vector.StrokeRect(screen, screenX, screenY, w, h, 1, color.RGBA{0, 255, 0, 255}, false)
	}
}
