//go:build !release

package primitives

// CameraProvider interface for world-to-screen coordinate conversion
type CameraProvider interface {
	Position() (float64, float64)
}

// WorldToScreen handles coordinate transformation from world to screen space.
type WorldToScreen struct {
	CameraX, CameraY float64
}

// NewWorldToScreen creates a WorldToScreen transformer from a camera.
func NewWorldToScreen(cam CameraProvider) WorldToScreen {
	cx, cy := cam.Position()
	return WorldToScreen{CameraX: cx, CameraY: cy}
}

// Transform converts world coordinates to screen coordinates.
func (w WorldToScreen) Transform(worldX, worldY float64) (float32, float32) {
	return float32(worldX - w.CameraX), float32(worldY - w.CameraY)
}

// TransformRect converts a world rectangle to screen coordinates.
func (w WorldToScreen) TransformRect(x, y, width, height float64) (float32, float32, float32, float32) {
	sx, sy := w.Transform(x, y)
	return sx, sy, float32(width), float32(height)
}

// TransformF64 converts world coordinates to screen coordinates as float64.
func (w WorldToScreen) TransformF64(worldX, worldY float64) (float64, float64) {
	return worldX - w.CameraX, worldY - w.CameraY
}
