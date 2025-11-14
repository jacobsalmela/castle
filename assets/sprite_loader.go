package assets

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"game/pkg/bump"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// ═══════════════════════════════════════════════════════════════════════════════
// UNIFIED ASSET STRUCTURE
// ═══════════════════════════════════════════════════════════════════════════════

// SpriteAsset represents a complete sprite asset (image + optional slice data).
// This unifies character sprites (with Aseprite JSON) and non-character sprites
// (like VFX and environment objects) into a single consistent structure.
type SpriteAsset struct {
	BaseName string                       // "bat", "knight", "flake" (filename without extension)
	Category string                       // "enemies", "player", "vfx", "environment", "ui"
	PNGPath  string                       // "images/enemies/bat.png"
	JSONPath string                       // "images/enemies/bat.json" (empty if no JSON)
	Image    *ebiten.Image                // Loaded sprite sheet
	Slices   map[string]map[int]bump.Rect // Slice data (nil if no JSON)
}

// AssetRegistry stores all discovered sprite assets.
// Key: basename (e.g., "bat", "knight", "flake")
// Value: Complete asset metadata + loaded data
var AssetRegistry = make(map[string]*SpriteAsset)

// ═══════════════════════════════════════════════════════════════════════════════
// AUTO-DISCOVERY LOADER
// ═══════════════════════════════════════════════════════════════════════════════

// debugConsoleEnabled is set during LoadAllAssets and used by helper functions.
// This avoids passing the flag through all internal functions.
var debugConsoleEnabled bool

// LoadAllAssets scans the embedded FS and loads all sprite assets.
// This replaces both InitImages() and LoadAllSlices().
//
// Process:
//  1. Scan images/ directory tree for .png files
//  2. For each PNG: load image, check for matching JSON, parse slices if present
//  3. Store in AssetRegistry by basename
//
// Benefits:
//   - No hardcoded file lists to maintain
//   - VFX/environment objects can have Aseprite animations
//   - Single unified API for all sprite access
//
// Parameters:
//   - embedFS: The embedded filesystem containing assets
//   - debugConsole: Whether to log verbose debug messages
func LoadAllAssets(embedFS embed.FS, debugConsole bool) error {
	debugConsoleEnabled = debugConsole
	log.Println("Auto-discovering sprite assets...")

	// Scan for all PNG files in images/ directory tree
	pngFiles, err := discoverPNGFiles(embedFS, "images")
	if err != nil {
		return fmt.Errorf("failed to discover PNG files: %w", err)
	}

	if debugConsoleEnabled {
		log.Printf("  Found %d PNG files to load", len(pngFiles))
	}

	// Load each sprite asset (PNG + optional JSON)
	loadedCount := 0
	for _, pngPath := range pngFiles {
		if err := loadSpriteAsset(embedFS, pngPath); err != nil {
			log.Printf("  ⚠ Failed to load %s: %v", pngPath, err)
			continue
		}
		loadedCount++
	}

	if debugConsoleEnabled {
		log.Printf("  ✓ Loaded %d sprite assets", loadedCount)
	}

	// Validate critical assets (player sprite must exist)
	if err := validateCriticalAssets(); err != nil {
		return err
	}

	return nil
}

// discoverPNGFiles recursively finds all .png files in a directory.
// Returns relative paths suitable for use with embed.FS.
func discoverPNGFiles(embedFS embed.FS, dir string) ([]string, error) {
	var pngFiles []string

	err := fs.WalkDir(embedFS, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check for .png extension
		if strings.HasSuffix(strings.ToLower(path), ".png") {
			pngFiles = append(pngFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return pngFiles, nil
}

// loadSpriteAsset loads a single sprite (PNG + optional JSON).
// This is the core loader that unifies character and non-character sprites.
func loadSpriteAsset(embedFS embed.FS, pngPath string) error {
	// Parse path to extract basename and category
	// Example: "images/enemies/bat.png" → basename="bat", category="enemies"
	baseName, category := parseSpritePathname(pngPath)

	// Load PNG image
	image, _, err := ebitenutil.NewImageFromFileSystem(embedFS, pngPath)
	if err != nil {
		return fmt.Errorf("failed to load PNG: %w", err)
	}

	// Check if matching JSON exists
	jsonPath := strings.Replace(pngPath, ".png", ".json", 1)
	slices, jsonExists := tryLoadSlices(embedFS, jsonPath)

	// Store in registry
	AssetRegistry[baseName] = &SpriteAsset{
		BaseName: baseName,
		Category: category,
		PNGPath:  pngPath,
		JSONPath: jsonPath,
		Image:    image,
		Slices:   slices,
	}

	// Log with slice indicator
	sliceIndicator := ""
	if jsonExists {
		sliceCount := 0
		for _, frameMap := range slices {
			sliceCount += len(frameMap)
		}
		sliceIndicator = fmt.Sprintf(" [%d slices]", sliceCount)
	}
	if debugConsoleEnabled {
		log.Printf("  ✓ %s/%s%s", category, baseName, sliceIndicator)
	}

	return nil
}

// parseSpritePathname extracts basename and category from a sprite path.
// Example: "images/enemies/bat.png" → ("bat", "enemies")
func parseSpritePathname(pngPath string) (basename, category string) {
	// Remove file extension
	withoutExt := strings.TrimSuffix(pngPath, filepath.Ext(pngPath))

	// Get basename (last component)
	basename = filepath.Base(withoutExt)

	// Extract category from path
	// "images/enemies/bat.png" → "enemies"
	// "images/vfx/flake.png" → "vfx"
	dir := filepath.Dir(withoutExt)
	if dir != "." && dir != "images" {
		category = filepath.Base(dir)
	} else {
		category = "root" // For files directly in images/ or root
	}

	return basename, category
}

// tryLoadSlices attempts to load and parse Aseprite JSON slice data.
// Returns (slices, true) if successful, (nil, false) if file doesn't exist or fails to parse.
// Non-fatal - sprites without JSON are valid (static images).
func tryLoadSlices(embedFS embed.FS, jsonPath string) (map[string]map[int]bump.Rect, bool) {
	// Check if JSON file exists
	jsonData, err := embedFS.ReadFile(jsonPath)
	if err != nil {
		// File doesn't exist or can't be read - not an error
		return nil, false
	}

	// Parse Aseprite JSON structure
	var asepriteData AsepriteJSON
	if err := json.Unmarshal(jsonData, &asepriteData); err != nil {
		log.Printf("  ⚠ Failed to parse JSON %s: %v", jsonPath, err)
		return nil, false
	}

	// Convert Aseprite slice format to our internal format
	sliceMap := make(map[string]map[int]bump.Rect)
	for _, slice := range asepriteData.Meta.Slices {
		frameMap := make(map[int]bump.Rect)
		for _, key := range slice.Keys {
			frameMap[key.Frame] = bump.Rect{
				X: float64(key.Bounds.X),
				Y: float64(key.Bounds.Y),
				W: float64(key.Bounds.W),
				H: float64(key.Bounds.H),
			}
		}
		sliceMap[slice.Name] = frameMap
	}

	return sliceMap, true
}

// validateCriticalAssets checks that essential assets are loaded.
// Currently validates that the player sprite (knight) exists.
func validateCriticalAssets() error {
	// Player sprite is critical - game cannot function without it
	if _, ok := AssetRegistry["knight"]; !ok {
		return fmt.Errorf("critical asset missing: knight.png (player sprite)")
	}

	// Validate knight has required slices
	knightAsset := AssetRegistry["knight"]
	if knightAsset.Slices == nil {
		return fmt.Errorf("critical asset incomplete: knight.json missing (required for hitbox data)")
	}

	requiredSlices := []string{"hurtbox", "hitbox", "blockbox"}
	for _, sliceName := range requiredSlices {
		if _, ok := knightAsset.Slices[sliceName]; !ok {
			log.Printf("  ⚠ Warning: knight.json missing '%s' slice", sliceName)
		}
	}

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// UNIFIED ACCESS API
// ═══════════════════════════════════════════════════════════════════════════════

// GetSpriteAsset returns the complete asset descriptor.
// Use this when you need both image and slice data.
func GetSpriteAsset(basename string) *SpriteAsset {
	return AssetRegistry[basename]
}

// GetSpriteImage returns just the sprite sheet image.
// This is the primary accessor for rendering.
//
// Returns nil if asset doesn't exist (caller should handle fallback).
func GetSpriteImage(basename string) *ebiten.Image {
	if asset := AssetRegistry[basename]; asset != nil {
		return asset.Image
	}
	return nil
}

// GetSpriteSlices returns just the slice data.
// Returns nil if asset doesn't exist or has no JSON.
func GetSpriteSlices(basename string) map[string]map[int]bump.Rect {
	if asset := AssetRegistry[basename]; asset != nil {
		return asset.Slices
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// DEBUG & INSPECTION
// ═══════════════════════════════════════════════════════════════════════════════

// ListAllAssets logs all loaded assets for debugging.
// Useful for verifying auto-discovery worked correctly.
func ListAllAssets() {
	log.Println("Loaded sprite assets:")
	for name, asset := range AssetRegistry {
		sliceInfo := ""
		if asset.Slices != nil {
			sliceNames := make([]string, 0, len(asset.Slices))
			for sliceName := range asset.Slices {
				sliceNames = append(sliceNames, sliceName)
			}
			sliceInfo = fmt.Sprintf(" [slices: %v]", sliceNames)
		}
		if debugConsoleEnabled {
			log.Printf("  %s [%s]: %s%s", name, asset.Category, asset.PNGPath, sliceInfo)
		}
	}
}
