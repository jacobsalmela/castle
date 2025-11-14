package prefabs

import (
	"game/assets"
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
)

const (
	// Visual properties
	graveSpriteW = 24 // Sprite width in pixels
	graveSpriteH = 19 // Sprite height in pixels

	// Render settings
	graveRenderLayer = -1 // Render behind player (negative layer index)

	// Tiled alignment
	graveBaselineOffset = -3 // Vertical offset to align with Tiled baseline (pixels)

	// Interaction text
	graveDefaultText = "Here lies a hero that saved the world from the darkness that consumed him. Rest in peace."
	gravePromptText  = " \n[Press Up to rest at the grave]"
)

// NewGravePrefab constructs a grave entity.
//
// Graves are interactive save/rest points with a single-phase lifecycle:
//  1. Idle: Display textbox when player is nearby
//  2. Activation: When player presses Up, trigger save and reset
//
// The grave uses the textbox system for proximity detection and displays
// an interaction prompt. When activated, it sets global save/reset flags.
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Tiled object position (top-left corner in world coordinates)
//   - tiledW: Tiled object width (unused, grave uses fixed sprite size)
//   - objH: Tiled object height (for bottom-alignment calculation)
//   - text: Custom interaction text (empty string uses default epitaph)
//
// Returns: EntityId of the created grave, or 0 if world is nil
func NewGravePrefab(world *ecs.World, x, y, tiledW, objH float64, text string) entities.EntityId {
	if world == nil {
		return 0
	}

	// Prepare interaction text
	interactionText := buildGraveText(text)

	// Create entity
	entityID := world.NewEntity()

	// === SPATIAL COMPONENT ===
	// Bottom-align to ground with baseline offset for Tiled compatibility
	alignedY := calculateGraveAlignedY(y, objH)
	transform := &components.Transform{
		X: x,
		Y: alignedY,
		W: graveSpriteW,
		H: graveSpriteH,
	}
	world.AddComponent(entityID, transform)

	// === VISUAL COMPONENT ===
	// Render with the grave sprite
	render := &components.Render{
		Image: assets.GetSpriteImage("grave"),
		Layer: graveRenderLayer,
	}
	world.AddComponent(entityID, render)

	// === ANIMATION COMPONENT ===
	// Animation FSM for consistency with other environment objects
	// grave.json has all tags (idle/activate/open) pointing to frame 0 (static sprite)
	anim := &components.Animation{
		FilesName:  "grave",
		State:      "idle",
		FSMInitial: "idle",
		FSMTransitions: map[string]string{
			"activate": "open",
			"open":     "open", // Self-loop to stay open
		},
		OX: 0, OY: 0, // No offsets needed for grave
		OXFlip: 0, OYFlip: 0,
		Layer: graveRenderLayer,
	}
	world.AddComponent(entityID, anim)
	InitializeAnimation(anim)

	// === COLLISION COMPONENT ===
	// Ghost collider - graves don't block movement, just need proximity detection
	collider := &components.Collider{
		Tags:      []string{"ghost"},
		QueryTags: []string{},
		Solid:     false, // Ghost collider doesn't block movement
		Immovable: false,
		OffsetX:   0,
		OffsetY:   0,
		Width:     0, // Use Transform size
		Height:    0, // Use Transform size
		FilterOut: []entities.EntityId{},
	}
	world.AddComponent(entityID, collider)

	// === INTERACTION COMPONENT ===
	// Textbox for proximity-based interaction prompt
	textboxData := &components.TextboxData{
		Text:      interactionText,
		Indicator: true, // Show position indicator
		Area: func() bump.Rect {
			return bump.NewRect(transform.X, transform.Y, transform.W, transform.H)
		},
		AdvanceFlicker:      false,
		AdvanceFlickerTimer: 0,
		Active:              false,
	}
	// Store entity bounds for indicator positioning
	textboxData.EntityX = transform.X
	textboxData.EntityY = transform.Y
	textboxData.EntityW = transform.W
	textboxData.EntityH = transform.H
	world.AddComponent(entityID, textboxData)

	// === BEHAVIOR COMPONENT ===
	// Grave marker for save/rest functionality
	grave := &components.Grave{}
	world.AddComponent(entityID, grave)

	return entityID
}

// buildGraveText prepares the interaction text with prompt appended.
// If custom text is empty, uses the default epitaph.
func buildGraveText(customText string) string {
	if customText == "" {
		return graveDefaultText + gravePromptText
	}
	return customText + gravePromptText
}

// calculateGraveAlignedY computes the Y position for bottom-alignment.
func calculateGraveAlignedY(tiledY, tiledH float64) float64 {
	return tiledY + tiledH - graveSpriteH + graveBaselineOffset
}
