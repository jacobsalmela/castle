// Package collision handles collision detection and physics integration.
package collision

import (
	"fmt"
	"game/pkg/bump"
	"game/pkg/tilemap/types"
	"math"

	"github.com/lafriks/go-tiled"
)

// LoadTilesetCollisionObjects loads collision data from tileset collision objects.
func LoadTilesetCollisionObjects(m *types.Map, space *bump.Space) {
	for _, tileset := range m.Data.Tilesets {
		for _, tile := range tileset.Tiles {
			if len(tile.ObjectGroups) == 0 {
				continue
			}
			obj := tile.ObjectGroups[0].Objects[0]
			rect := createCollisionRect(obj)
			tags := buildCollisionTags(obj)
			placeTileColliders(m, space, tileset, tile.ID, rect, tags)
		}
	}
}

// buildCollisionTags creates collision tags based on object properties.
func buildCollisionTags(obj *tiled.Object) []bump.Tag {
	tags := []bump.Tag{"map"}

	if isLadder(obj) {
		tags = append(tags, "passthrough", "ladder")
	} else if isPassthrough(obj) {
		tags = append(tags, "passthrough")
	}

	if obj.Polygons != nil {
		tags = append(tags, "slope")
	}

	return tags
}

// isLadder checks if object is a ladder.
func isLadder(obj *tiled.Object) bool {
	return obj.Class == "ladder" || obj.Type == "ladder"
}

// isPassthrough checks if object is passthrough.
func isPassthrough(obj *tiled.Object) bool {
	return obj.Class == "passthrough" || obj.Type == "passthrough"
}

// placeTileColliders places collision objects in the physics space.
func placeTileColliders(m *types.Map, space *bump.Space, tileset *tiled.Tileset, tileID uint32, rect bump.Rect, tags []bump.Tag) {
	for _, layer := range m.Data.Layers {
		for y := range m.Data.Height {
			for x := range m.Data.Width {
				layerTile := layer.Tiles[y*m.Data.Width+x]
				if !layerTile.IsNil() && layerTile.Tileset == tileset && layerTile.ID == tileID {
					placeCollider(space, layerTile, x, y, m.Data.TileWidth, m.Data.TileHeight, rect, tags)
				}
			}
		}
	}
}

// placeCollider adds a single collision rect to the space.
func placeCollider(space *bump.Space, layerTile *tiled.LayerTile, x, y, tileW, tileH int, rect bump.Rect, tags []bump.Tag) {
	tileRect := bump.Rect{
		X:    rect.X + float64(x*tileW),
		Y:    rect.Y + float64(y*tileH),
		W:    rect.W,
		H:    rect.H,
		Type: rect.Type,
	}
	space.Set(layerTile, tileRect, tags...)
}

// createCollisionRect creates a collision rectangle from an object.
func createCollisionRect(obj *tiled.Object) bump.Rect {
	rect := bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height}
	if obj.Polygons != nil {
		rect = PolygonRect(obj)
	}
	return rect
}

// LoadBumpObjects loads collision objects from a Tiled object layer.
func LoadBumpObjects(m *types.Map, space *bump.Space, objectGroupName string) {
	objects := findObjectGroup(m.Data, objectGroupName)
	for _, obj := range objects {
		loadObjectCollision(space, obj)
	}
}

// findObjectGroup retrieves objects from a named group.
func findObjectGroup(data *tiled.Map, name string) []*tiled.Object {
	for _, group := range data.ObjectGroups {
		if name == group.Name {
			return group.Objects
		}
	}
	return nil
}

// loadObjectCollision loads a single object's collision data.
func loadObjectCollision(space *bump.Space, obj *tiled.Object) {
	// CRITICAL FIX: Tiled uses bottom-left origin, game uses top-left origin
	// For rectangles, we need to convert: Y_topLeft = Y_bottomLeft - Height
	rect := bump.Rect{X: obj.X, Y: obj.Y - obj.Height, W: obj.Width, H: obj.Height, Type: bump.Full}
	tags := []bump.Tag{"map"}

	if obj.Polygons != nil {
		rect = PolygonRect(obj)
		tags = append(tags, "slope")
	}
	if isLadder(obj) {
		tags = append(tags, "passthrough", "ladder")
	} else if isPassthrough(obj) {
		tags = append(tags, "passthrough")
	}

	space.Set(obj, rect, tags...)
}

// PolygonRect calculates a bounding rectangle and slope type from a polygon.
func PolygonRect(object *tiled.Object) bump.Rect {
	bounds := calculatePolygonBounds(object)
	slope := detectSlopeType(object.Polygons[0].Points, bounds)

	return bump.Rect{
		X:    object.X + bounds.left,
		Y:    object.Y + bounds.top,
		W:    bounds.right - bounds.left,
		H:    bounds.bottom - bounds.top,
		Type: slope,
	}
}

type polygonBounds struct {
	left, right, top, bottom float64
}

// calculatePolygonBounds finds the bounding box of a polygon.
func calculatePolygonBounds(object *tiled.Object) polygonBounds {
	points := *object.Polygons[0].Points
	bounds := polygonBounds{}

	for _, p := range points {
		bounds.left = math.Min(bounds.left, p.X)
		bounds.right = math.Max(bounds.right, p.X)
		bounds.top = math.Min(bounds.top, p.Y)
		bounds.bottom = math.Max(bounds.bottom, p.Y)
	}

	return bounds
}

// detectSlopeType determines slope direction from polygon corners.
func detectSlopeType(points *tiled.Points, bounds polygonBounds) bump.RectType {
	corners := findCorners(points, bounds)
	return cornerToSlope(corners)
}

// findCorners identifies which corners of the bounding box are in the polygon.
func findCorners(points *tiled.Points, bounds polygonBounds) [4]bool {
	corners := [4]bool{} // topLeft, topRight, bottomLeft, bottomRight

	for _, p := range *points {
		switch {
		case p.X == bounds.left && p.Y == bounds.top:
			corners[0] = true
		case p.X == bounds.right && p.Y == bounds.top:
			corners[1] = true
		case p.X == bounds.left && p.Y == bounds.bottom:
			corners[2] = true
		case p.X == bounds.right && p.Y == bounds.bottom:
			corners[3] = true
		}
	}

	return corners
}

// cornerToSlope converts missing corners to slope type.
func cornerToSlope(corners [4]bool) bump.RectType {
	for i, hasCorner := range corners {
		if !hasCorner {
			switch i {
			case 0:
				return bump.BottomRightSlope
			case 1:
				return bump.BottomLeftSlope
			case 2:
				return bump.TopRightSlope
			case 3:
				return bump.TopLeftSlope
			}
		}
	}
	return bump.Full
}

// LoadLadderTiles scans all map layers for tiles with type="ladder" or class="ladder"
// and registers their positions in the LadderRegistry AND collision space.
func LoadLadderTiles(m *types.Map, registry types.LadderRegistry, space *bump.Space) error {
	if registry == nil {
		return nil
	}

	registry.Clear()

	// Step 1: Collect all ladder positions first
	type ladderPos struct {
		x, y int
		rect bump.Rect
		key  string
	}
	var ladderPositions []ladderPos

	// Scan all tilesets for tiles marked as "ladder"
	for _, tileset := range m.Data.Tilesets {
		for _, tile := range tileset.Tiles {
			// Check if this tile is a ladder (by type or class)
			isLadder := tile.Type == "ladder" || tile.Class == "ladder"
			if !isLadder {
				continue
			}

			// Find all instances of this ladder tile in the map
			for _, layer := range m.Data.Layers {
				for y := 0; y < m.Data.Height; y++ {
					for x := 0; x < m.Data.Width; x++ {
						layerTile := layer.Tiles[y*m.Data.Width+x]

						// Check if this layer tile matches our ladder tile
						if !layerTile.IsNil() && layerTile.Tileset == tileset && layerTile.ID == tile.ID {
							// Calculate world position (top-left corner)
							worldX := float64(x * m.Data.TileWidth)
							worldY := float64(y * m.Data.TileHeight)

							// Create ladder rectangle (tile dimensions)
							ladderRect := bump.Rect{
								X: worldX,
								Y: worldY,
								W: float64(m.Data.TileWidth),
								H: float64(m.Data.TileHeight),
							}

							// Add to registry for climbing detection
							registry.AddLadder(ladderRect)

							// Store position for top-tile detection
							ladderKey := fmt.Sprintf("ladder_tile_%d_%d_%d", tileset.FirstGID+tile.ID, x, y)
							ladderPositions = append(ladderPositions, ladderPos{
								x: x, y: y, rect: ladderRect, key: ladderKey,
							})
						}
					}
				}
			}
		}
	}

	// Step 2: Create a set of ladder positions for quick lookup
	ladderSet := make(map[[2]int]bool)
	for _, pos := range ladderPositions {
		ladderSet[[2]int{pos.x, pos.y}] = true
	}

	// Step 3: Add to collision space with appropriate tags
	for _, pos := range ladderPositions {
		if space == nil {
			continue
		}

		// Check if there's a ladder tile directly above this one
		hasLadderAbove := ladderSet[[2]int{pos.x, pos.y - 1}]

		if hasLadderAbove {
			// Non-top tile: fully passthrough (no platform behavior)
			space.Set(pos.key, pos.rect, bump.Tag("ladder"))
		} else {
			// Top tile: acts as one-way platform
			space.Set(pos.key, pos.rect, bump.Tag("passthrough"), bump.Tag("ladder"), bump.Tag("ladder_top"))
		}
	}

	return nil
}
