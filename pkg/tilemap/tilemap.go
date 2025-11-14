// Package tilemap provides loading, rendering, and querying for Tiled maps.
//
// The package is organized into several subpackages:
//   - types: Core data structures (Map, Tile, Properties)
//   - loader: Map loading from Tiled .tmx files
//   - renderer: Rendering maps to Ebiten images or render queues
//   - query: Searching for objects, tiles, and entity data
//   - collision: Integration with bump physics engine
//
// Example usage:
//
//	m := tilemap.NewMap("maps/level1.tmx", 2, embedFS, "screen", "normal")
//	tilemap.Update(m, dt)
//	tilemap.DrawToQueue(m, renderQueue, camera)
package tilemap

import (
	"io/fs"

	"game/pkg/bump"
	"game/pkg/tilemap/collision"
	"game/pkg/tilemap/loader"
	"game/pkg/tilemap/query"
	"game/pkg/tilemap/renderer"
	"game/pkg/tilemap/types"

	"github.com/lafriks/go-tiled"
)

// Re-export types for backwards compatibility
type (
	Map            = types.Map
	Tile           = types.Tile
	Properties     = types.Properties
	EntityObject   = types.EntityObject
	Camera         = types.Camera
	MapRenderQueue = types.MapRenderQueue
	LadderRegistry = types.LadderRegistry
)

// LayerIndex is the base z-depth for map layers.
var LayerIndex = &renderer.LayerIndex

// NewMap creates and initializes a Map from a Tiled map file.
func NewMap(mapPath string, backLayersNum int, filesystem fs.FS, drawImagesTags ...string) *Map {
	return loader.NewMap(mapPath, backLayersNum, filesystem, drawImagesTags...)
}

// Update advances all map animations.
func Update(m *Map, dt float64) {
	renderer.Update(m, dt)
}

// DrawToQueue renders the map to a render queue (Pure ECS).
func DrawToQueue(m *Map, queue MapRenderQueue, camera Camera) {
	renderer.DrawToQueue(m, queue, camera)
}

// FindObjectID finds an object by its unique ID.
func FindObjectID(m *Map, id int) (*tiled.Object, error) {
	return query.FindObjectID(m, id)
}

// FindObjectFromTileID finds the first object with a matching tile ID.
func FindObjectFromTileID(m *Map, id uint32, objectGroupName string) (*tiled.Object, error) {
	return query.FindObjectFromTileID(m, id, objectGroupName)
}

// FindTilePosition finds all positions where a tile GID appears.
func FindTilePosition(m *Map, gid uint32) [][2]float64 {
	return query.FindTilePosition(m, gid)
}

// TilesFromPosition retrieves tiles at a world position.
func TilesFromPosition(m *Map, x, y float64, removeTiles bool, space *bump.Space) (map[string]*Tile, error) {
	return query.TilesFromPosition(m, x, y, removeTiles, space)
}

// LoadTilesetCollisionObjects loads collision data from tileset objects.
func LoadTilesetCollisionObjects(m *Map, space *bump.Space) {
	collision.LoadTilesetCollisionObjects(m, space)
}

// LoadBumpObjects loads collision objects from a named object group.
func LoadBumpObjects(m *Map, space *bump.Space, objectGroupName string) {
	collision.LoadBumpObjects(m, space, objectGroupName)
}

// VisitEntityObjects walks entity objects in a named group.
func VisitEntityObjects(m *Map, objectGroupName string, visit func(obj EntityObject)) {
	query.VisitEntityObjects(m, objectGroupName, visit)
}

// GetObjects retrieves all objects from a named object group.
func GetObjects(m *Map, objectGroupName string) []*tiled.Object {
	return query.GetObjects(m, objectGroupName)
}

// GetObjectsRects converts objects to bump rectangles.
func GetObjectsRects(m *Map, objectGroupName string) []bump.Rect {
	return query.GetObjectsRects(m, objectGroupName)
}

// TileWidth returns the width of a single tile in pixels.
func TileWidth(m *Map) int {
	return query.TileWidth(m)
}

// TileHeight returns the height of a single tile in pixels.
func TileHeight(m *Map) int {
	return query.TileHeight(m)
}

// Width returns the width of the map in tiles.
func Width(m *Map) int {
	return query.Width(m)
}

// Height returns the height of the map in tiles.
func Height(m *Map) int {
	return query.Height(m)
}

// LoadLadderTiles scans all map layers for tiles with type="ladder" or class="ladder"
// and registers their positions in the LadderRegistry AND collision space.
func LoadLadderTiles(m *Map, registry LadderRegistry, space *bump.Space) error {
	return collision.LoadLadderTiles(m, registry, space)
}
