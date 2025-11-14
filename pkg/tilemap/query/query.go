// Package query provides entity and tile lookup functions.
package query

import (
	"fmt"
	"log"
	"strconv"

	"game/pkg/bump"
	"game/pkg/tilemap/types"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

const viewPropName = "view"

// FindObjectID finds an object by its unique ID.
func FindObjectID(m *types.Map, id int) (*tiled.Object, error) {
	for _, group := range m.Data.ObjectGroups {
		for _, obj := range group.Objects {
			if int(obj.ID) == id {
				return obj, nil
			}
		}
	}
	return nil, tiled.ErrInvalidObjectPoint
}

// FindObjectFromTileID finds the first object with a matching tile ID.
func FindObjectFromTileID(m *types.Map, id uint32, objectGroupName string) (*tiled.Object, error) {
	objects := getObjectsFromGroup(m.Data, objectGroupName)

	for _, obj := range objects {
		if matchesTileID(m.Data, obj, id) {
			return obj, nil
		}
	}

	return nil, tiled.ErrInvalidTileGID
}

// getObjectsFromGroup retrieves objects from a named group (or all groups).
func getObjectsFromGroup(data *tiled.Map, groupName string) []*tiled.Object {
	var objects []*tiled.Object

	for _, group := range data.ObjectGroups {
		if groupName == "" {
			objects = append(objects, group.Objects...)
		} else if groupName == group.Name {
			return group.Objects
		}
	}

	return objects
}

// matchesTileID checks if an object's tile matches the given ID.
func matchesTileID(data *tiled.Map, obj *tiled.Object, id uint32) bool {
	tile, err := data.TileGIDToTile(obj.GID)
	if err != nil || tile.IsNil() {
		return false
	}
	return tile.ID == id
}

// VisitEntityObjects walks entity objects in a named group.
func VisitEntityObjects(m *types.Map, objectGroupName string, visit func(obj types.EntityObject)) {
	objects := getObjectsFromGroup(m.Data, objectGroupName)
	if objects == nil || visit == nil {
		return
	}

	for _, obj := range objects {
		entityObj := parseEntityObject(m, obj)
		visit(entityObj)
	}
}

// parseEntityObject converts a Tiled object to an EntityObject.
func parseEntityObject(m *types.Map, obj *tiled.Object) types.EntityObject {
	props := &types.Properties{Custom: map[string]string{}}
	var tileID uint32

	// Extract tile data
	if obj.GID != 0 {
		if tile, err := m.Data.TileGIDToTile(obj.GID); err == nil && !tile.IsNil() {
			props.FlipX = tile.HorizontalFlip
			props.FlipY = tile.VerticalFlip
			tileID = tile.ID
		}
	}

	// Parse properties
	for _, prop := range obj.Properties {
		parseProperty(m, props, *prop, obj)
	}

	return types.EntityObject{
		X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height,
		ID: uint(obj.ID), Name: obj.Name, TileID: tileID,
		Props: props,
	}
}

// parseProperty handles a single object property.
func parseProperty(m *types.Map, props *types.Properties, prop tiled.Property, obj *tiled.Object) {
	if prop.Name == viewPropName {
		id, _ := strconv.Atoi(prop.Value)
		vobj, err := FindObjectID(m, id)
		if err != nil {
			log.Printf("WARNING: cannot find view object with id %s for entity '%s' (ID: %d): %v - skipping view property", prop.Value, obj.Name, obj.ID, err)
			return
		}
		props.View = vobj
	} else {
		props.Custom[prop.Name] = prop.Value
	}
}

// FindTilePosition finds all positions where a tile GID appears.
func FindTilePosition(m *types.Map, gid uint32) [][2]float64 {
	var positions [][2]float64

	// Check animated tiles first
	positions = append(positions, findInAnimations(m, gid)...)

	// Check static tiles
	positions = append(positions, findInLayers(m, gid)...)

	return positions
}

// findInAnimations searches for a GID in layer animations.
func findInAnimations(m *types.Map, gid uint32) [][2]float64 {
	var positions [][2]float64

loop:
	for _, layers := range m.Layers {
		for _, layer := range layers {
			if anim, ok := layer.Animations[gid]; ok {
				for _, pos := range anim.Positions {
					positions = append(positions, [2]float64{pos.X, pos.Y})
				}
				break loop
			}
		}
	}

	return positions
}

// findInLayers searches for a GID in static layer tiles.
func findInLayers(m *types.Map, gid uint32) [][2]float64 {
	var positions [][2]float64

	for _, layer := range m.Data.Layers {
		for y := range m.Data.Height {
			for x := range m.Data.Width {
				tile := layer.Tiles[y*m.Data.Width+x]
				if !tile.IsNil() && tile.Tileset.FirstGID+tile.ID == gid {
					positions = append(positions, [2]float64{
						float64(x * m.Data.TileWidth),
						float64(y * m.Data.TileHeight),
					})
				}
			}
		}
	}

	return positions
}

// TilesFromPosition retrieves tiles at a world position.
func TilesFromPosition(m *types.Map, x, y float64, removeTiles bool, space *bump.Space) (map[string]*types.Tile, error) {
	mapX, mapY := int(x)/m.Data.TileWidth, int(y)/m.Data.TileHeight

	if err := validatePosition(m, mapX, mapY); err != nil {
		return nil, err
	}

	position := mapY*m.Data.Width + mapX
	tiles, layerIndex, skipped := findTileAtPosition(m, position)

	if tiles == nil {
		return nil, fmt.Errorf("map: no tile found at position: %f, %f", x, y)
	}

	if removeTiles {
		removeTileAtPosition(m, mapX, mapY, layerIndex, skipped, space, tiles)
	}

	return tiles, nil
}

// validatePosition checks if map coordinates are in bounds.
func validatePosition(m *types.Map, mapX, mapY int) error {
	if mapX < 0 || mapY < 0 || mapX >= m.Data.Width || mapY >= m.Data.Height {
		return fmt.Errorf("map: position out of bounds: %d, %d", mapX, mapY)
	}
	return nil
}

// findTileAtPosition searches for a non-nil tile at the position.
func findTileAtPosition(m *types.Map, position int) (map[string]*types.Tile, int, int) {
	skipped := 0

	for layerIndex := len(m.Data.Layers) - 1; layerIndex >= 0; layerIndex-- {
		layer := m.Data.Layers[layerIndex]

		if !layer.Visible {
			skipped++
			continue
		}

		tile := layer.Tiles[position]
		if tile.IsNil() {
			continue
		}

		// Found a tile, create tile map for all image tags
		tiles := createTileMap(m, *tile, position)
		return tiles, layerIndex, skipped
	}

	return nil, -1, skipped
}

// createTileMap builds a map of Tile objects for each image tag.
func createTileMap(m *types.Map, tile tiled.LayerTile, position int) map[string]*types.Tile {
	mapX := position % m.Data.Width
	mapY := position / m.Data.Width

	tiles := map[string]*types.Tile{}
	for imageTag := range m.Layers {
		tiles[imageTag] = &types.Tile{
			X:        float64(mapX * m.Data.TileWidth),
			Y:        float64(mapY * m.Data.TileHeight),
			FlipX:    tile.HorizontalFlip,
			FlipY:    tile.VerticalFlip,
			FlipR:    tile.DiagonalFlip,
			Image:    m.Tileset[imageTag][tile.Tileset.FirstGID+tile.ID],
			ImageTag: imageTag,
		}
	}

	return tiles
}

// removeTileAtPosition removes a tile from all layers and collision space.
func removeTileAtPosition(m *types.Map, mapX, mapY, layerIndex, skipped int, space *bump.Space, tiles map[string]*types.Tile) {
	for _, layers := range m.Layers {
		imageLayerIndex := len(m.Layers) - 1 - (len(m.Data.Layers) - 1 - layerIndex + skipped)
		clearTileFromLayer(layers[imageLayerIndex].Image, mapX, mapY, m.Data.TileWidth, m.Data.TileHeight)
	}

	if space != nil {
		layer := m.Data.Layers[layerIndex]
		position := mapY*m.Data.Width + mapX
		space.Remove(layer.Tiles[position])
	}
}

// clearTileFromLayer draws an empty tile to clear the position.
func clearTileFromLayer(layerImage *ebiten.Image, mapX, mapY, tileW, tileH int) {
	emptyTile := ebiten.NewImage(tileW, tileH)
	op := &ebiten.DrawImageOptions{Blend: ebiten.BlendCopy}
	op.GeoM.Translate(float64(mapX*tileW), float64(mapY*tileH))
	layerImage.DrawImage(emptyTile, op)
}

// GetObjects retrieves all objects from a named object group.
func GetObjects(m *types.Map, objectGroupName string) []*tiled.Object {
	for _, group := range m.Data.ObjectGroups {
		if objectGroupName == group.Name {
			return group.Objects
		}
	}
	return nil
}

// GetObjectsRects converts objects to bump rectangles.
func GetObjectsRects(m *types.Map, objectGroupName string) []bump.Rect {
	objects := GetObjects(m, objectGroupName)
	if objects == nil {
		return nil
	}

	rects := make([]bump.Rect, len(objects))
	for i, obj := range objects {
		rects[i] = bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height}
	}

	return rects
}

// TileWidth returns the width of a single tile in pixels.
func TileWidth(m *types.Map) int {
	if m.Data == nil {
		return 0
	}
	return m.Data.TileWidth
}

// TileHeight returns the height of a single tile in pixels.
func TileHeight(m *types.Map) int {
	if m.Data == nil {
		return 0
	}
	return m.Data.TileHeight
}

// Width returns the width of the map in tiles.
func Width(m *types.Map) int {
	if m.Data == nil {
		return 0
	}
	return m.Data.Width
}

// Height returns the height of the map in tiles.
func Height(m *types.Map) int {
	if m.Data == nil {
		return 0
	}
	return m.Data.Height
}
