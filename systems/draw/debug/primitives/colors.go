//go:build !release

package primitives

import "image/color"

// Combat colors
var (
	HitboxOutline   = color.RGBA{0, 255, 0, 255}      // Green
	HurtboxOutline  = color.RGBA{255, 0, 0, 255}      // Red
	HurtboxFill     = color.RGBA{255, 0, 0, 120}      // Red (translucent)
	BlockboxOutline = color.RGBA{64, 160, 255, 255}   // Blue
	ParryboxOutline = color.RGBA{255, 255, 0, 255}    // Yellow
)

// Physics colors
var (
	CollisionBox   = color.RGBA{0, 255, 255, 255}    // Cyan
	PhysicsBox     = color.RGBA{255, 255, 0, 255}    // Yellow
	VelocityVector = color.RGBA{0, 255, 0, 255}      // Green
	GroundedLine   = color.RGBA{0, 255, 0, 255}      // Green
	AirborneLine   = color.RGBA{255, 0, 0, 255}      // Red
	PhysicsWhite   = color.RGBA{255, 255, 255, 255}  // White
	PhysicsCyan    = color.RGBA{0, 255, 255, 255}    // Cyan
)

// AI colors
var (
	TargetLine     = color.RGBA{255, 255, 0, 255}    // Yellow
	DetectionRange = color.RGBA{255, 255, 0, 255}   // Yellow
)

// UI colors
var (
	Background     = color.RGBA{64, 64, 64, 180}     // Dark gray
	BackgroundDark = color.RGBA{32, 32, 32, 200}     // Darker gray
	TextWhite      = color.RGBA{255, 255, 255, 255}  // White
	TextGray       = color.RGBA{200, 200, 200, 255}  // Light gray
)

// Behavior tree colors
var (
	BTRunning    = color.RGBA{0, 255, 0, 255}      // Green
	BTSuccess    = color.RGBA{0, 150, 255, 255}    // Blue
	BTFailure    = color.RGBA{255, 0, 0, 255}      // Red
	BTNotExecuted = color.RGBA{128, 128, 128, 255}  // Gray
)

// Tile debug colors
var (
	TileOutline = color.RGBA{255, 255, 255, 60}   // Faint white
	TileText    = color.RGBA{255, 255, 255, 200}  // White
)

// WithAlpha returns a copy of the color with a new alpha value.
func WithAlpha(c color.RGBA, alpha uint8) color.RGBA {
	return color.RGBA{c.R, c.G, c.B, alpha}
}
