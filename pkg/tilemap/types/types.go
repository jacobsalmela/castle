// Package types defines core data structures for the tilemap package.
package types

import (
	"image"
	"io/fs"

	"game/pkg/bump"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
)

// Camera represents the minimal interface needed by Map rendering.
// This avoids import cycles with resources package.
type Camera interface {
	BoundsWithOffsetAndParallax(offsetX, offsetY int, parallaxX, parallaxY float64) image.Rectangle
	Position() (float64, float64)
}

// MapRenderQueue interface for queue-based map rendering without import cycles.
// Implemented by resources.RenderQueue.
type MapRenderQueue interface {
	PushMapTile(image *ebiten.Image, targetType int, layer int, geom ebiten.GeoM)
}

// Properties holds parsed Tiled object properties.
type Properties struct {
	FlipX, FlipY bool
	View         *tiled.Object
	Custom       map[string]string
}

// Tile represents a single map tile.
type Tile struct {
	X, Y                float64
	FlipX, FlipY, FlipR bool
	Image               *ebiten.Image
	ImageTag            string
}

// EntityObject represents a Tiled entity for spawning.
type EntityObject struct {
	X, Y, W, H float64
	ID         uint
	Name       string
	TileID     uint32
	Props      *Properties
}

// AnimationFrame holds a single animation frame.
type AnimationFrame struct {
	Image    *ebiten.Image
	Duration float64
}

// AnimationPosition tracks where an animation plays on the map.
type AnimationPosition struct {
	X, Y                float64
	FlipX, FlipY, FlipR bool
}

// Animation manages a tile animation sequence.
type Animation struct {
	Frames    []*AnimationFrame
	Positions []AnimationPosition
	Timer     float64
	Current   int
}

// LayerData holds rendered layer state.
type LayerData struct {
	Image                *ebiten.Image
	Animations           map[uint32]*Animation
	OffsetX, OffsetY     int
	ParallaxX, ParallaxY float64
	FS                   fs.FS
}

// Map is the main tilemap structure.
type Map struct {
	Data                *tiled.Map
	Layers              map[string][]*LayerData
	ObjectLayers        []*LayerData
	Tileset             map[string][]*ebiten.Image
	FirstImageTag       string
	BackgroundLayersNum int
}

// LadderRegistry interface to avoid import cycles.
type LadderRegistry interface {
	AddLadder(rect bump.Rect)
	Clear()
}
