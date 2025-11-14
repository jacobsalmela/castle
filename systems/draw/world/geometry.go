package world

import (
	"math"

	"game/components"

	"github.com/hajimehoshi/ebiten/v2"
)

// ═══════════════════════════════════════════════════════════════════════════════
// GEOMETRY HELPERS
// Extracted from sprites.go for better code organization and reduced complexity.
// ═══════════════════════════════════════════════════════════════════════════════

// applyFlip applies flip transformations to geometry and returns the translation offsets.
func applyFlip(geom *ebiten.GeoM, flipX, flipY bool, w, h float64) (dx, dy float64) {
	sx, sy := 1.0, 1.0

	if flipX {
		sx = -1.0
		dx = math.Floor(w)
	}
	if flipY {
		sy = -1.0
		dy = math.Floor(h)
	}

	geom.Scale(sx, sy)
	return dx, dy
}

// animGeom builds transformation matrix for animated sprites (FromAnim=true).
// Uses center-based flip offset for sprite sheet frames.
func animGeom(c *components.Render, w, h float64, opGeom ebiten.GeoM) ebiten.GeoM {
	var sx, sy, dx, dy float64 = 1, 1, 0, 0
	if c.FlipX {
		sx, dx = -1, math.Floor(w/2)+dx
	}
	if c.FlipY {
		sy, dy = -1, math.Floor(h/2)+dy
	}
	var geom ebiten.GeoM
	geom.Scale(sx, sy)
	geom.Translate(c.X+dx, c.Y+dy)
	geom.Concat(opGeom)
	return geom
}

// defaultGeom builds transformation matrix for static sprites (FromAnim=false).
// Includes rotation around center point.
func defaultGeom(c *components.Render, w, h float64, opGeom ebiten.GeoM) ebiten.GeoM {
	var geom ebiten.GeoM

	dx, dy := applyFlip(&geom, c.FlipX, c.FlipY, w, h)

	geom.Translate(-w/2, -h/2)
	geom.Rotate(c.R)
	geom.Translate(w/2, h/2)
	geom.Translate(c.X+dx, c.Y+dy)
	geom.Concat(opGeom)

	return geom
}
