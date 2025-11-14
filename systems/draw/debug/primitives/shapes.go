//go:build !release

package primitives

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// StrokeRect draws a rectangle outline.
func StrokeRect(screen *ebiten.Image, x, y, w, h float32, strokeWidth float32, col color.RGBA) {
	vector.StrokeLine(screen, x, y, x+w, y, strokeWidth, col, false)
	vector.StrokeLine(screen, x+w, y, x+w, y+h, strokeWidth, col, false)
	vector.StrokeLine(screen, x+w, y+h, x, y+h, strokeWidth, col, false)
	vector.StrokeLine(screen, x, y+h, x, y, strokeWidth, col, false)
}

// FillRect draws a filled rectangle.
func FillRect(screen *ebiten.Image, x, y, w, h float64, col color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}

	width := int(math.Ceil(w))
	height := int(math.Ceil(h))
	if width <= 0 || height <= 0 {
		return
	}

	img := ebiten.NewImage(width, height)
	img.Fill(col)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

// FillCircle draws a filled circle.
func FillCircle(screen *ebiten.Image, x, y, radius float32, col color.RGBA) {
	size := int(radius*2) + 2
	if size < 2 {
		size = 2
	}

	img := ebiten.NewImage(size, size)
	centerOffset := float32(size) / 2

	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			dx := float32(px) - centerOffset
			dy := float32(py) - centerOffset
			if dx*dx+dy*dy <= radius*radius {
				img.Set(px, py, col)
			}
		}
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x-centerOffset), float64(y-centerOffset))
	screen.DrawImage(img, op)
}

// StrokeCircle draws a circle outline.
func StrokeCircle(screen *ebiten.Image, centerX, centerY, radius float32, col color.RGBA) {
	segments := 20
	if radius > 10 {
		segments = 28
	}
	if radius > 20 {
		segments = 36
	}

	angleStep := 2 * math.Pi / float64(segments)

	for i := 0; i < segments; i++ {
		angle1 := float64(i) * angleStep
		angle2 := float64(i+1) * angleStep

		x1 := centerX + radius*float32(math.Cos(angle1))
		y1 := centerY + radius*float32(math.Sin(angle1))
		x2 := centerX + radius*float32(math.Cos(angle2))
		y2 := centerY + radius*float32(math.Sin(angle2))

		vector.StrokeLine(screen, x1, y1, x2, y2, 1, col, false)
	}
}
