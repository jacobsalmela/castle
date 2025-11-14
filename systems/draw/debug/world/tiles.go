//go:build !release

package world

import (
	"fmt"
	"game/assets/fonts"
	"game/components"
	"game/ecs"
	"game/pkg/tilemap"
	"game/resources"
	"game/systems/draw/debug/primitives"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawTileDebugOverlay draws tile coordinates on each map tile.
// Enable with Shift+T. Shows:
// - Faint white outlines: Tile boundaries
// - White text: Tile coordinates (x,y)
func DrawTileDebugOverlay(world *ecs.World, tiledMap *tilemap.Map, cam primitives.CameraProvider, screen *ebiten.Image, vp *components.ViewPort) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryTile) {
		return
	}
	if tiledMap == nil || cam == nil || screen == nil || vp == nil {
		return
	}

	tileWidth := tiledMap.Data.TileWidth
	tileHeight := tiledMap.Data.TileHeight
	mapWidth := tiledMap.Data.Width
	mapHeight := tiledMap.Data.Height

	if tileWidth == 0 || tileHeight == 0 || mapWidth == 0 || mapHeight == 0 {
		return
	}

	// Get camera position to only draw visible tiles
	cx, cy := cam.Position()
	screenW, screenH := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())

	// Calculate visible tile range (with small margin)
	startX := int(math.Max(0, math.Floor(cx/float64(tileWidth))-1))
	endX := int(math.Min(float64(mapWidth), math.Ceil((cx+screenW)/float64(tileWidth))+1))
	startY := int(math.Max(0, math.Floor(cy/float64(tileHeight))-1))
	endY := int(math.Min(float64(mapHeight), math.Ceil((cy+screenH)/float64(tileHeight))+1))

	// Draw tile grid and coordinates for visible tiles
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			// Calculate tile screen position (convert to device pixels using DPR)
			tileScreenX := (float64(x*tileWidth) - cx) * vp.DPR
			tileScreenY := (float64(y*tileHeight) - cy) * vp.DPR
			tileW := float32(float64(tileWidth) * vp.DPR)
			tileH := float32(float64(tileHeight) * vp.DPR)

			// Draw faint white outline around tile (convert to float32 for vector calls)
			sx := float32(tileScreenX)
			sy := float32(tileScreenY)
			vector.StrokeLine(screen, sx, sy, sx+tileW, sy, 1, primitives.TileOutline, false)
			vector.StrokeLine(screen, sx+tileW, sy, sx+tileW, sy+tileH, 1, primitives.TileOutline, false)
			vector.StrokeLine(screen, sx+tileW, sy+tileH, sx, sy+tileH, 1, primitives.TileOutline, false)
			vector.StrokeLine(screen, sx, sy+tileH, sx, sy, 1, primitives.TileOutline, false)

			// Draw tile coordinates in center of tile
			tileCoord := fmt.Sprintf("%d,%d", x, y)

			// Estimate text width to center it (rough approximation for NanoFont)
			textWidth := float64(len(tileCoord)) * 3.5

			// Center text in tile (text positions are device pixels)
			textX := tileScreenX + float64(tileW)/2 - textWidth/2
			textY := tileScreenY + float64(tileH)/2 - 3*vp.DPR

			// Draw text using NanoFont (created with DPR-scaled size). Positions are
			// in device pixels so the text will align with the tile correctly.
			opts := &text.DrawOptions{}
			opts.GeoM.Translate(textX, textY)
			opts.ColorScale.ScaleWithColor(primitives.TileText)
			text.Draw(screen, tileCoord, fonts.NanoFont, opts)
		}
	}
}
