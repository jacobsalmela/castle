package assets

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"game/pkg/bump"
)

// ═══════════════════════════════════════════════════════════════════════════════
// ASEPRITE SLICE DATA STRUCTURES
// ═══════════════════════════════════════════════════════════════════════════════

// AsepriteSliceData represents the complete slice export from an Aseprite file.
// This matches the structure of the "slices" array in the JSON export.
type AsepriteSliceData struct {
	Name  string             `json:"name"`  // Slice name (e.g., "hurtbox", "hitbox", "blockbox")
	Color string             `json:"color"` // Color hex (e.g., "#fe5b59ff")
	Keys  []AsepriteSliceKey `json:"keys"`  // Frame-specific bounds
}

// AsepriteSliceKey represents a single frame's slice bounds.
type AsepriteSliceKey struct {
	Frame  int                 `json:"frame"`  // Frame number (0-indexed)
	Bounds AsepriteSliceBounds `json:"bounds"` // Rectangle bounds
	Pivot  *AsepriteSlicePivot `json:"pivot"`  // Optional pivot point (for rotation)
}

// AsepriteSliceBounds represents the rectangle bounds of a slice.
type AsepriteSliceBounds struct {
	X int `json:"x"` // X position (sprite-relative)
	Y int `json:"y"` // Y position (sprite-relative)
	W int `json:"w"` // Width
	H int `json:"h"` // Height
}

// AsepriteSlicePivot represents an optional pivot point (currently unused).
type AsepriteSlicePivot struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// AsepriteJSON represents the root structure of an Aseprite JSON export.
// We only parse the "slices" field for hitbox data.
type AsepriteJSON struct {
	Meta AsepriteMetaData `json:"meta"`
}

// AsepriteMetaData contains metadata including slice definitions.
type AsepriteMetaData struct {
	Slices []AsepriteSliceData `json:"slices"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// SLICE PARSING
// ═══════════════════════════════════════════════════════════════════════════════

// LoadAllSlices parses all Aseprite JSON files in the assets directory.
// This should be called during InitSync() before entities are created.
//
// Parsing strategy:
//  1. Find all .json files in the embedded FS (organized by subdirectories)
//  2. Parse each file's "slices" field
//  3. Convert slice bounds to bump.Rect format
//  4. Store in SliceRegistry by file name (without path/extension)
func LoadAllSlices(fs embed.FS) error {
	// Animation file paths organized by category
	// Each entry: [base name, relative path to JSON]
	animFiles := map[string]string{
		// Enemies (in images/enemies/)
		"bat":      "images/enemies/bat.json",
		"rat":      "images/enemies/rat.json",
		"crawler":  "images/enemies/crawler.json",
		"ent":      "images/enemies/ent.json",
		"ghoul":    "images/enemies/ghoul.json",
		"skeleman": "images/enemies/skeleman.json",

		// NPCs (in images/npcs/)
		"ferragus": "images/npcs/ferragus.json",
		"gram":     "images/npcs/gram.json",
		"oscar":    "images/npcs/oscar.json",

		// Player (in images/player/)
		"knight": "images/player/knight.json",
	}

	loaded := 0
	for baseName, jsonPath := range animFiles {
		if err := parseSliceFile(fs, baseName, jsonPath); err != nil {
			// Non-critical: some files may not have JSON exports
			log.Printf("  ⚠ Slice parsing skipped for %s (%s): %v", baseName, jsonPath, err)
			continue
		}
		loaded++
	}

	if loaded > 0 {
		log.Printf("  ✓ Loaded slice data from %d animation files", loaded)

		// Validate knight slices (player sprite - critical)
		validateKnightSlices()
	} else {
		log.Println("  ⚠ No slice data found in animation files")
	}

	return nil
}

// validateKnightSlices validates that the knight (player) sprite has valid slice data.
// This is critical for debugging sprite rendering issues.
func validateKnightSlices() {
	asset, ok := AssetRegistry["knight"]
	if !ok || asset.Slices == nil {
		log.Println("  ⚠ WARNING: knight has NO slice data!")
		return
	}

	log.Println("  → Knight slice validation:")
	for sliceName, frameMap := range asset.Slices {
		log.Printf("    • %s: %d frames", sliceName, len(frameMap))

		// Show first few frames as examples
		frameCount := 0
		for frame, rect := range frameMap {
			if frameCount < 3 {
				log.Printf("      - Frame %d: x=%.0f y=%.0f w=%.0f h=%.0f",
					frame, rect.X, rect.Y, rect.W, rect.H)
			}
			frameCount++
		}
	}
}

// parseSliceFile parses a single Aseprite JSON file and stores its slices.
func parseSliceFile(fs embed.FS, baseName, jsonPath string) error {
	// Read JSON file from embedded FS
	data, err := fs.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON structure
	var aseJSON AsepriteJSON
	if err := json.Unmarshal(data, &aseJSON); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if slices exist
	if len(aseJSON.Meta.Slices) == 0 {
		return fmt.Errorf("no slices defined")
	}

	// Convert to SliceMap structure
	sliceMap := make(map[string]map[int]bump.Rect)
	totalSlices := 0

	for _, slice := range aseJSON.Meta.Slices {
		sliceName := slice.Name
		frameMap := make(map[int]bump.Rect)

		for _, key := range slice.Keys {
			frameNum := key.Frame
			rect := bump.Rect{
				X: float64(key.Bounds.X),
				Y: float64(key.Bounds.Y),
				W: float64(key.Bounds.W),
				H: float64(key.Bounds.H),
			}
			frameMap[frameNum] = rect
			totalSlices++
		}

		sliceMap[sliceName] = frameMap
	}

	// Note: This function is deprecated. Slices are now loaded via sprite_loader.go
	// which stores them directly in AssetRegistry during LoadAllAssets().
	// This parseSliceFile is only kept for reference.

	log.Printf("    • %s: %d slices across %d types", baseName, totalSlices, len(sliceMap))
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════════
// SLICE ACCESS API
// ═══════════════════════════════════════════════════════════════════════════════

// GetSliceMap returns the full slice map for an animation file.
// Returns nil if no slices exist.
//
// Example:
//
//	slices := assets.GetSliceMap("knight")
//	if slices != nil {
//	    hurtbox := slices["hurtbox"][0] // Frame 0 hurtbox
//	}
func GetSliceMap(animFileName string) map[string]map[int]bump.Rect {
	if asset, ok := AssetRegistry[animFileName]; ok {
		return asset.Slices
	}
	return nil
}

// GetSlice returns a specific slice rect for a given frame.
// Returns (rect, true) if found, (zero rect, false) if not found.
//
// Example:
//
//	if rect, ok := assets.GetSlice("knight", "hurtbox", 0); ok {
//	    // Use rect for hitbox
//	}
func GetSlice(animFileName, sliceName string, frame int) (bump.Rect, bool) {
	asset, ok := AssetRegistry[animFileName]
	if !ok || asset.Slices == nil {
		return bump.Rect{}, false
	}

	frameMap := asset.Slices[sliceName]
	if frameMap == nil {
		return bump.Rect{}, false
	}

	rect, ok := frameMap[frame]
	return rect, ok
}

// HasSlices reports whether an animation file has any slice data.
func HasSlices(animFileName string) bool {
	if asset, ok := AssetRegistry[animFileName]; ok {
		return asset.Slices != nil && len(asset.Slices) > 0
	}
	return false
}

// GetSliceNames returns all slice names defined for an animation file.
// Useful for debugging or validation.
//
// Example:
//
//	names := assets.GetSliceNames("knight")
//	// ["hurtbox", "hitbox", "blockbox"]
func GetSliceNames(animFileName string) []string {
	asset, ok := AssetRegistry[animFileName]
	if !ok || asset.Slices == nil {
		return nil
	}

	names := make([]string, 0, len(asset.Slices))
	for name := range asset.Slices {
		names = append(names, name)
	}
	return names
}

// GetSliceFrameCount returns the number of frames with slice data.
// Returns 0 if the slice doesn't exist.
//
// Example:
//
//	count := assets.GetSliceFrameCount("knight", "hitbox")
//	// 4 (frames 14, 15, 20, 21 for knight attack)
func GetSliceFrameCount(animFileName, sliceName string) int {
	asset, ok := AssetRegistry[animFileName]
	if !ok || asset.Slices == nil {
		return 0
	}

	frameMap := asset.Slices[sliceName]
	return len(frameMap)
}

// ═══════════════════════════════════════════════════════════════════════════════
// DEBUG UTILITIES
// ═══════════════════════════════════════════════════════════════════════════════

// PrintSliceRegistry dumps the entire slice registry to the log.
// Useful for debugging slice loading issues.
func PrintSliceRegistry() {
	log.Println("=== Slice Registry (AssetRegistry) ===")
	for animFile, asset := range AssetRegistry {
		if asset.Slices == nil || len(asset.Slices) == 0 {
			continue
		}
		log.Printf("  %s:", animFile)
		for sliceName, frameMap := range asset.Slices {
			frames := make([]int, 0, len(frameMap))
			for frame := range frameMap {
				frames = append(frames, frame)
			}
			log.Printf("    • %s: %d frames", sliceName, len(frames))
		}
	}
}

// ValidateSliceRegistry checks for common issues in the slice registry.
// Returns warnings (non-critical issues) and errors (critical issues).
func ValidateSliceRegistry() (warnings []string, errors []string) {
	// Count assets with slice data
	sliceCount := 0
	for _, asset := range AssetRegistry {
		if asset.Slices != nil && len(asset.Slices) > 0 {
			sliceCount++
		}
	}

	if sliceCount == 0 {
		errors = append(errors, "No assets with slice data loaded")
		return
	}

	// Check for missing standard slices
	standardSlices := []string{"hurtbox", "hitbox"}
	for animFile, asset := range AssetRegistry {
		if asset.Slices == nil {
			continue
		}
		// Skip validation for non-combat entities (chests, doors, etc.)
		if isCombatEntity(animFile) {
			for _, sliceName := range standardSlices {
				if _, ok := asset.Slices[sliceName]; !ok {
					warnings = append(warnings, fmt.Sprintf("%s missing '%s' slice", animFile, sliceName))
				}
			}
		}
	}

	return warnings, errors
}

// isCombatEntity reports whether an animation file represents a combat-capable entity.
// Non-combat entities (chests, doors, projectiles) don't need hurtbox/hitbox slices.
func isCombatEntity(animFile string) bool {
	nonCombat := []string{"chest", "door", "rock", "flake", "grave", "smoke", "spike"}
	for _, nc := range nonCombat {
		if strings.Contains(strings.ToLower(animFile), nc) {
			return false
		}
	}
	return true
}

// GetSliceColor returns the color associated with a slice type.
// This is metadata from Aseprite and can be used for debug rendering.
//
// Note: Colors are stored in the JSON but not currently parsed.
// Future enhancement: Parse slice.Color field from JSON and store in registry.
func GetSliceColor(animFileName, sliceName string) string {
	// Return standard colors based on slice name (matches Aseprite defaults)
	switch sliceName {
	case "hurtbox":
		return "#fe5b59ff" // Red
	case "hitbox":
		return "#6acd5bff" // Green
	case "blockbox", "block":
		return "#0000ffff" // Blue
	case "healbox":
		return "#ffff00ff" // Yellow
	default:
		return "#ffffffff" // White
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// MIGRATION HELPERS
// ═══════════════════════════════════════════════════════════════════════════════

