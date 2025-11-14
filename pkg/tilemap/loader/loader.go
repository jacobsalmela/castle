// Package loader handles map loading and initialization.
package loader

import (
	"fmt"
	"image"
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"strings"

	"game/pkg/tilemap/types"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
)

const secondToMillisecond = 1000

// extVariationFS wraps a filesystem to load texture variants.
type extVariationFS struct {
	ext, target string
	fs          fs.FS
}

// Open intercepts file opens to substitute variant filenames.
func (vfs *extVariationFS) Open(name string) (fs.File, error) {
	dir, file := path.Split(name)
	if fileExt := strings.Split(file, "."); len(fileExt) > 1 && fileExt[1] == vfs.target {
		name = path.Join(dir, fmt.Sprintf("%s_%s.%s", fileExt[0], vfs.ext, fileExt[1]))
	}

	return vfs.fs.Open(name)
}

// NewMap creates and initializes a Map from a Tiled map file.
func NewMap(mapPath string, backLayersNum int, filesystem fs.FS, drawImagesTags ...string) *types.Map {
	data := loadTiledMap(mapPath, filesystem)
	if data == nil {
		return nil
	}

	tilesets := buildAllTilesets(data, filesystem, drawImagesTags)
	layers := buildAllLayers(data, mapPath, filesystem, tilesets, drawImagesTags)
	objectLayers, err := buildObjectLayers(data, mapPath, filesystem)
	if err != nil {
		log.Println("Error building object layers from Tiled map:", err)
	}

	m := &types.Map{
		Data:                data,
		Layers:              layers,
		ObjectLayers:        objectLayers,
		Tileset:             tilesets,
		FirstImageTag:       drawImagesTags[0],
		BackgroundLayersNum: backLayersNum,
	}

	if err := Render(m); err != nil {
		log.Println("Error rendering Tiled map:", err)
	}

	return m
}

// loadTiledMap loads the .tmx file.
func loadTiledMap(mapPath string, filesystem fs.FS) *tiled.Map {
	data, err := tiled.LoadFile(mapPath, tiled.WithFileSystem(filesystem))
	if err != nil {
		log.Println("Error parsing Tiled map:", err)
		return nil
	}
	return data
}

// buildAllTilesets creates tilesets for all image tags.
func buildAllTilesets(data *tiled.Map, filesystem fs.FS, drawImagesTags []string) map[string][]*ebiten.Image {
	tilesets := map[string][]*ebiten.Image{}

	for i, tag := range drawImagesTags {
		vfs := createVariationFS(filesystem, tag, i)

		tileset, err := buildTileset(data, vfs)
		if err != nil {
			log.Printf("Error building tileset for tag %s: %v", tag, err)
			continue
		}

		tilesets[tag] = tileset
	}

	return tilesets
}

// createVariationFS creates a filesystem variant for texture tags.
func createVariationFS(filesystem fs.FS, tag string, index int) fs.FS {
	if index == 0 {
		return filesystem
	}
	return &extVariationFS{ext: tag, target: "png", fs: filesystem}
}

// buildAllLayers creates layer data for all image tags.
func buildAllLayers(data *tiled.Map, mapPath string, filesystem fs.FS, tilesets map[string][]*ebiten.Image, drawImagesTags []string) map[string][]*types.LayerData {
	layers := map[string][]*types.LayerData{}

	for i, tag := range drawImagesTags {
		vfs := createVariationFS(filesystem, tag, i)
		lastImageTag := i == len(drawImagesTags)-1

		layerList, err := buildLayers(data, mapPath, vfs, tilesets[tag], lastImageTag)
		if err != nil {
			log.Printf("Error building layers for tag %s: %v", tag, err)
			continue
		}

		layers[tag] = layerList
	}

	return layers
}

// buildTileset loads and slices tileset images into individual tiles.
func buildTileset(data *tiled.Map, filesystem fs.FS) ([]*ebiten.Image, error) {
	tileImages := []*ebiten.Image{nil} // Index 0 is reserved

	for _, tileset := range data.Tilesets {
		if tileset.Image == nil {
			continue
		}

		img, err := loadTilesetImage(filesystem, tileset)
		if err != nil {
			return nil, err
		}

		tileImages = append(tileImages, sliceTilesetImage(img, tileset)...)
	}

	return tileImages, nil
}

// loadTilesetImage loads the tileset source image.
func loadTilesetImage(filesystem fs.FS, tileset *tiled.Tileset) (image.Image, error) {
	path := filepath.ToSlash(tileset.GetFileFullPath(tileset.Image.Source))
	file, err := filesystem.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

// sliceTilesetImage slices a tileset image into individual tile images.
func sliceTilesetImage(img image.Image, tileset *tiled.Tileset) []*ebiten.Image {
	tiles := make([]*ebiten.Image, 0, tileset.TileCount)

	for tileID := range uint32(tileset.TileCount) { //nolint: gosec
		tileRect := tileset.GetTileRect(tileID)
		tileImg := cropImage(img, tileRect)
		tiles = append(tiles, ebiten.NewImageFromImage(tileImg))
	}

	return tiles
}

// cropImage extracts a sub-image from an image.
func cropImage(img image.Image, crop image.Rectangle) image.Image {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	simg, ok := img.(subImager)
	if !ok {
		panic("image does not support cropping")
	}

	return simg.SubImage(crop)
}

// buildLayers creates layer data for visible tile layers.
func buildLayers(data *tiled.Map, mapPath string, filesystem fs.FS, tileImages []*ebiten.Image, removeAnimatedTiles bool) ([]*types.LayerData, error) {
	layersData := []*types.LayerData{}

	for i, layer := range data.Layers {
		if !layer.Visible {
			continue
		}

		anims, err := extractLayerAnimations(data, tileImages, i, removeAnimatedTiles)
		if err != nil {
			return nil, fmt.Errorf("map: error extracting %s animations: %w", layer.Name, err)
		}

		parallax := extractParallax(layer.ParallaxX, layer.ParallaxY)

		layersData = append(layersData, &types.LayerData{
			Image:      nil,
			Animations: anims,
			OffsetX:    layer.OffsetX,
			OffsetY:    layer.OffsetY,
			ParallaxX:  parallax.X,
			ParallaxY:  parallax.Y,
			FS:         filesystem,
		})
	}

	return layersData, nil
}

// buildObjectLayers creates layer data for object and image layers.
func buildObjectLayers(data *tiled.Map, mapPath string, filesystem fs.FS) ([]*types.LayerData, error) {
	layersData := []*types.LayerData{}

	// Process drawable object groups
	layersData = append(layersData, processObjectGroups(data, filesystem)...)

	// Process image layers (parallax backgrounds)
	layersData = append(layersData, processImageLayers(data, mapPath, filesystem)...)

	return layersData, nil
}

type parallax struct {
	X, Y float64
}

// extractParallax converts Tiled parallax values (0 = default to 1.0).
func extractParallax(x, y float32) parallax {
	px, py := 1.0, 1.0
	if x != 0 {
		px = float64(x)
	}
	if y != 0 {
		py = float64(y)
	}
	return parallax{X: px, Y: py}
}

// processObjectGroups builds layer data from object groups with "draw" property.
func processObjectGroups(data *tiled.Map, filesystem fs.FS) []*types.LayerData {
	layersData := []*types.LayerData{}

	for _, layer := range data.ObjectGroups {
		if layer.Visible && layer.Properties.GetBool("draw") {
			parallax := extractParallax(layer.ParallaxX, layer.ParallaxY)

			layersData = append(layersData, &types.LayerData{
				Image:      nil,
				Animations: nil,
				OffsetX:    layer.OffsetX,
				OffsetY:    layer.OffsetY,
				ParallaxX:  parallax.X,
				ParallaxY:  parallax.Y,
				FS:         filesystem,
			})
		}
	}

	return layersData
}

// processImageLayers builds layer data from image layers.
func processImageLayers(data *tiled.Map, mapPath string, filesystem fs.FS) []*types.LayerData {
	layersData := []*types.LayerData{}

	for _, layer := range data.ImageLayers {
		if !layer.Visible || layer.Image == nil {
			continue
		}

		img, err := loadImageLayerImage(filesystem, mapPath, layer.Image.Source)
		if err != nil {
			log.Printf("Warning: Failed to load image layer '%s': %v", layer.Name, err)
			continue
		}

		parallax := extractParallax(layer.ParallaxX, layer.ParallaxY)
		layerImage := ebiten.NewImageFromImage(img)

		layersData = append(layersData, &types.LayerData{
			Image:      layerImage,
			Animations: nil,
			OffsetX:    layer.OffsetX,
			OffsetY:    layer.OffsetY,
			ParallaxX:  parallax.X,
			ParallaxY:  parallax.Y,
			FS:         filesystem,
		})
	}

	return layersData
}

// loadImageLayerImage loads an image for an image layer.
func loadImageLayerImage(filesystem fs.FS, mapPath, imgSource string) (image.Image, error) {
	imgPath := imgSource
	if !filepath.IsAbs(imgPath) {
		mapDir := filepath.Dir(mapPath)
		imgPath = filepath.Join(mapDir, imgPath)
	}
	imgPath = filepath.ToSlash(imgPath)

	imgFile, err := filesystem.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	return img, err
}

// extractLayerAnimations builds animation data from a layer's animated tiles.
func extractLayerAnimations(data *tiled.Map, tileImages []*ebiten.Image, layerIndex int, removeAnimatedTiles bool) (map[uint32]*types.Animation, error) {
	animationDefs := buildAnimationDefinitions(data, tileImages)
	populateAnimationPositions(data, layerIndex, animationDefs, removeAnimatedTiles)

	return animationDefs, nil
}

// buildAnimationDefinitions creates animation frame data from tilesets.
func buildAnimationDefinitions(data *tiled.Map, tileImages []*ebiten.Image) map[uint32]*types.Animation {
	animationFrames := map[uint32]*types.Animation{}

	for _, tileset := range data.Tilesets {
		for _, tile := range tileset.Tiles {
			if len(tile.Animation) == 0 {
				continue
			}

			// Build frames inline
			frames := make([]*types.AnimationFrame, len(tile.Animation))
			for i, frame := range tile.Animation {
				frames[i] = &types.AnimationFrame{
					Image:    tileImages[tileset.FirstGID+frame.TileID],
					Duration: float64(frame.Duration) / secondToMillisecond,
				}
			}
			animationFrames[tileset.FirstGID+tile.ID] = &types.Animation{Frames: frames}
		}
	}

	return animationFrames
}

// populateAnimationPositions finds where animated tiles appear on the map.
func populateAnimationPositions(data *tiled.Map, layerIndex int, animationDefs map[uint32]*types.Animation, removeAnimatedTiles bool) {
	layer := data.Layers[layerIndex]
	tileID := 0

	for y := range data.Height {
		for x := range data.Width {
			tile := layer.Tiles[tileID]

			if !tile.IsNil() {
				processAnimatedTile(data, layer, x, y, tileID, *tile, animationDefs, removeAnimatedTiles)
			}

			tileID++
		}
	}
}

// processAnimatedTile checks if a tile is animated and records its position.
func processAnimatedTile(data *tiled.Map, layer *tiled.Layer, x, y, tileID int, tile tiled.LayerTile, animationDefs map[uint32]*types.Animation, removeAnimatedTiles bool) {
	gid := tile.Tileset.FirstGID + tile.ID

	if animation, ok := animationDefs[gid]; ok {
		position := types.AnimationPosition{
			X:     float64(x * data.TileWidth),
			Y:     float64(y * data.TileHeight),
			FlipX: tile.HorizontalFlip,
			FlipY: tile.VerticalFlip,
			FlipR: tile.DiagonalFlip,
		}
		animation.Positions = append(animation.Positions, position)

		if removeAnimatedTiles {
			layer.Tiles[tileID] = tiled.NilLayerTile
		}
	}
}

// Render renders all layers to their internal images.
func Render(m *types.Map) error {
	if err := renderTileLayers(m); err != nil {
		return err
	}

	if err := renderObjectLayers(m); err != nil {
		return err
	}

	return nil
}

// renderTileLayers renders all visible tile layers.
func renderTileLayers(m *types.Map) error {
	skipped := 0

	for i, layer := range m.Data.Layers {
		if !layer.Visible {
			skipped++
			continue
		}

		if err := renderTileLayer(m, i, skipped); err != nil {
			return fmt.Errorf("map: tiled layer %s unsupported for rendering: %w", layer.Name, err)
		}
	}

	return nil
}

// renderTileLayer renders a single tile layer across all image tags.
func renderTileLayer(m *types.Map, layerIndex, skipped int) error {
	for _, layers := range m.Layers {
		adjustedIndex := layerIndex - skipped
		renderer, err := render.NewRendererWithFileSystem(m.Data, layers[adjustedIndex].FS)
		if err != nil {
			return fmt.Errorf("map: tiled map unsupported for rendering: %w", err)
		}

		if err := renderer.RenderLayer(layerIndex); err != nil {
			return err
		}

		updateLayerImage(layers[adjustedIndex], renderer.Result)
	}

	return nil
}

// renderObjectLayers renders all drawable object layers.
func renderObjectLayers(m *types.Map) error {
	skipped := 0

	for i, objectGroup := range m.Data.ObjectGroups {
		if !objectGroup.Visible || !objectGroup.Properties.GetBool("draw") {
			skipped++
			continue
		}

		if err := renderObjectLayer(m, i, skipped, objectGroup.Name); err != nil {
			return fmt.Errorf("map: tiled object group %s unsupported for rendering: %w", objectGroup.Name, err)
		}
	}

	return nil
}

// renderObjectLayer renders a single object layer.
func renderObjectLayer(m *types.Map, groupIndex, skipped int, groupName string) error {
	adjustedIndex := groupIndex - skipped
	objectLayer := m.ObjectLayers[adjustedIndex]

	renderer, err := render.NewRendererWithFileSystem(m.Data, objectLayer.FS)
	if err != nil {
		return fmt.Errorf("map: tiled map unsupported for rendering: %w", err)
	}

	if err := renderer.RenderObjectGroup(groupIndex); err != nil {
		return err
	}

	updateLayerImage(objectLayer, renderer.Result)
	return nil
}

// updateLayerImage replaces a layer's image with a new rendered result.
func updateLayerImage(layer *types.LayerData, newImage image.Image) {
	if layer.Image != nil {
		layer.Image.Deallocate()
	}
	layer.Image = ebiten.NewImageFromImage(newImage)
}
