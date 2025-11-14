//go:build !release

package debug

import (
	"fmt"
	"game/assets/fonts"
	"game/components"
	"game/ecs"
	"game/pkg/config"
	"game/resources"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// ActiveCategory represents an active debug category with its name and color
type ActiveCategory struct {
	Name  string
	Color color.RGBA
}

// getActiveDebugCategories returns a list of currently enabled debug categories.
// This uses the Pure ECS DebugState resource and key bindings registry.
func getActiveDebugCategories(world *ecs.World) []ActiveCategory {
	debugState := ecs.Resource[resources.DebugState](world)
	debugCategories := ecs.Resource[resources.DebugCategories](world)
	if debugState == nil || debugCategories == nil {
		return nil
	}

	var active []ActiveCategory

	// Get all active debug flags
	activeNames := debugState.GetActive()
	for _, name := range activeNames {
		// Get color from debug categories resource
		if color, ok := debugCategories.GetColor(name); ok {
			active = append(active, ActiveCategory{
				Name:  name,
				Color: color,
			})
		}
	}

	return active
}

// drawECSOverlay renders the ECS status overlay showing sprite count and active debug flags.
// This is the Pure ECS replacement for the drawECSFlagsDevice function.
//
// Shows:
// - Bottom-left: ECS sprite count and active debug categories (colored by category)
// - Top-right: Collision debug legend (when collision debug is enabled)
//
// Enable/disable with Cmd+Shift+1 debug toggle.
func drawECSOverlay(world *ecs.World, screen *ebiten.Image, vp *components.ViewPort) {
	if world == nil || screen == nil || vp == nil {
		return
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Get active categories by checking DebugState resource
	activeCategories := getActiveDebugCategories(world)

	// Only show if there are active categories OR ECS overlay is explicitly enabled
	debugState := ecs.Resource[resources.DebugState](world)
	if len(activeCategories) == 0 && (debugState == nil || !debugState.IsEnabled(resources.DebugCategoryECSOverlay)) {
		return
	}

	// Use FiveByFiveFont which is DPR-scaled
	face := fonts.FiveByFiveFont

	// Calculate position in logical pixels (screen is device pixels, need to divide by DPR scale)
	screenBounds := screen.Bounds()
	logicalFontHeight := 24.0                    // FiveByFiveFont base size (reduced from 16 to 8)
	logicalLineHeight := logicalFontHeight + 2.0 // Count lines needed for bottom-left display
	lineCount := 0

	// Show sprite count only if ECS overlay explicitly enabled
	if debugState != nil && debugState.IsEnabled(resources.DebugCategoryECSOverlay) {
		lineCount++ // Sprite count line
		lineCount++ // DPR info line
	}

	if len(activeCategories) > 0 {
		lineCount++ // Debug status line
	}

	if lineCount == 0 && (debugState == nil || !debugState.IsEnabled(resources.DebugCategoryCollision)) {
		return // Nothing to draw
	}

	// Draw unified debug legends in top-right corner for all active overlays
	drawUnifiedDebugLegends(world, screen, screenBounds, logicalLineHeight, face, vp.DPR)

	// Draw bottom-left status if there's anything to show
	if lineCount > 0 {
		// Calculate starting Y position in device pixels (bottom - total height)
		lineHeightDevice := logicalLineHeight * vp.DPR
		totalHeightDevice := float64(lineCount) * lineHeightDevice
		startYDevice := float64(screenBounds.Dy()) - totalHeightDevice - 10.0*vp.DPR

		// Starting X position in device pixels
		xDevice := 10.0 * vp.DPR

		// Measure the width needed for the background box (measurements are already in device pixels)
		maxWidth := 0.0

		// Measure sprite count and DPR lines if shown
		if debugState != nil && debugState.IsEnabled(resources.DebugCategoryECSOverlay) {
			stats := ecs.Resource[resources.RenderStats](world)
			spritesDrawn := 0
			if stats != nil {
				spritesDrawn = stats.SpritesDrawn
			}
			w := measureText(fmt.Sprintf("ECS Sprites: %d", spritesDrawn), face)
			if w > maxWidth {
				maxWidth = w
			}

			// Measure DPR line
			dprText := fmt.Sprintf("DPR: %.1f (Config HighDPI: %v)", vp.DPR, cfg.Screen.HighDpi)
			w = measureText(dprText, face)
			if w > maxWidth {
				maxWidth = w
			}
		}

		// Measure debug status line if shown
		if len(activeCategories) > 0 {
			// Measure "Debug:" prefix
			debugWidth := measureText("Debug:", face) + 5*vp.DPR // Add gap

			// Measure each category
			for _, cat := range activeCategories {
				debugWidth += measureText(cat.Name, face) + 8*vp.DPR // Add spacing
			}

			if debugWidth > maxWidth {
				maxWidth = debugWidth
			}
		}

		// Box dimensions with padding (in device pixels)
		paddingX := 8.0 * vp.DPR
		paddingY := 4.0 * vp.DPR
		boxWidth := maxWidth + paddingX*2
		boxHeight := totalHeightDevice + paddingY*2

		// Draw background box
		bgImage := ebiten.NewImage(int(boxWidth), int(boxHeight))
		bgColor := color.RGBA{64, 64, 64, 180} // Dark gray, translucent
		bgImage.Fill(bgColor)

		bgOpts := &ebiten.DrawImageOptions{}
		bgOpts.GeoM.Translate(xDevice-paddingX, startYDevice-paddingY)
		screen.DrawImage(bgImage, bgOpts)

		// Current Y position (in device pixels)
		yDevice := startYDevice

		// Draw sprite count and DPR only if ECS overlay explicitly enabled
		if debugState != nil && debugState.IsEnabled(resources.DebugCategoryECSOverlay) {
			stats := ecs.Resource[resources.RenderStats](world)
			spritesDrawn := 0
			if stats != nil {
				spritesDrawn = stats.SpritesDrawn
			}
			drawColoredText(screen, fmt.Sprintf("ECS Sprites: %d", spritesDrawn), xDevice, yDevice, face, color.RGBA{255, 255, 255, 255})
			yDevice += lineHeightDevice

			// Draw DPR info with color indicating HighDPI status
			// Use config value, not DPR value, to determine color
			dprColor := color.RGBA{255, 165, 0, 255} // Orange for disabled
			if cfg.Screen.HighDpi {
				dprColor = color.RGBA{0, 255, 0, 255} // Green for enabled
			}
			dprText := fmt.Sprintf("DPR: %.1f (Config HighDPI: %v)", vp.DPR, cfg.Screen.HighDpi)
			drawColoredText(screen, dprText, xDevice, yDevice, face, dprColor)
			yDevice += lineHeightDevice
		}

		// Draw debug status line with colored category names
		if len(activeCategories) > 0 {
			drawDebugStatusLine(screen, xDevice, yDevice, lineHeightDevice, face, activeCategories, vp.DPR)
		}
	}
}

// drawDebugStatusLine renders "Debug:" followed by colored category names
func drawDebugStatusLine(screen *ebiten.Image, x, y, lineHeight float64, face text.Face, categories []ActiveCategory, dpr float64) float64 {
	currentX := x

	// Draw "Debug:" prefix in white
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(currentX, y)
	opts.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
	text.Draw(screen, "Debug:", face, opts)

	// Measure "Debug:" width
	advance := measureText("Debug:", face)
	currentX += advance + 5*dpr // Small gap after prefix

	// Draw each category name with its color
	for _, cat := range categories {
		if cat.Name == "Body" {
			// Special case: Body uses mixed colors (yellow, green, red)
			currentX = drawMixedColorText(screen, "Body", currentX, y, face, []color.RGBA{
				{255, 255, 0, 255}, // B = yellow
				{0, 255, 0, 255},   // o = green
				{255, 0, 0, 255},   // d = red
				{255, 255, 0, 255}, // y = yellow
			})
		} else {
			// Draw category name with its color
			drawColoredText(screen, cat.Name, currentX, y, face, cat.Color)
			advance = measureText(cat.Name, face)
			currentX += advance
		}

		// Add space after category (scaled by DPR)
		currentX += 8 * dpr
	}

	return y + lineHeight
}

// drawMixedColorText draws text with each character in a different color
func drawMixedColorText(screen *ebiten.Image, txt string, x, y float64, face text.Face, colors []color.RGBA) float64 {
	currentX := x
	for i, ch := range txt {
		charColor := colors[i%len(colors)]
		charStr := string(ch)
		drawColoredText(screen, charStr, currentX, y, face, charColor)
		advance := measureText(charStr, face)
		currentX += advance
	}
	return currentX
}

// drawColoredText draws text with the specified color
func drawColoredText(screen *ebiten.Image, txt string, x, y float64, face text.Face, col color.RGBA) {
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(x, y)
	opts.ColorScale.ScaleWithColor(col)
	text.Draw(screen, txt, face, opts)
}

// measureText returns the width of the text in pixels
func measureText(txt string, face text.Face) float64 {
	advance, _ := text.Measure(txt, face, 0)
	return advance
}

// legendLine represents a single line in a debug legend with text and color
type legendLine struct {
	text  string
	color color.RGBA
}

// debugLegendSection represents a section of the debug legend (e.g., "Collision", "Physics")
type debugLegendSection struct {
	title string
	lines []legendLine
}

// drawUnifiedDebugLegends draws all active debug legends in top-right corner with DPR-scaled positioning.
// This replaces the old DrawDebugLegend system from legend.go with proper DPR scaling.
func drawUnifiedDebugLegends(world *ecs.World, screen *ebiten.Image, screenBounds image.Rectangle, logicalLineHeight float64, face text.Face, dpr float64) {
	debugState := ecs.Resource[resources.DebugState](world)
	if debugState == nil {
		return
	}

	// Collect all active debug legend sections
	var sections []debugLegendSection

	if debugState.IsEnabled(resources.DebugCategoryCollision) {
		sections = append(sections, debugLegendSection{
			title: "Collision (Cmd+1)",
			lines: []legendLine{
				{"Impact point", color.RGBA{255, 255, 0, 255}},
				{"Normal (direction)", color.RGBA{255, 0, 0, 255}},
				{"Connection", color.RGBA{0, 255, 0, 255}},
				{"Collision box", color.RGBA{0, 255, 255, 255}},
			},
		})
	}

	if debugState.IsEnabled(resources.DebugCategoryPhysics) {
		sections = append(sections, debugLegendSection{
			title: "Physics (Cmd+2)",
			lines: []legendLine{
				{"Yellow boxes: Entity boundaries", color.RGBA{255, 255, 0, 255}},
				{"Green arrows: Velocity vectors", color.RGBA{0, 255, 0, 255}},
				{"Green line above: Grounded", color.RGBA{0, 255, 0, 255}},
				{"Red line above: Airborne", color.RGBA{255, 0, 0, 255}},
			},
		})
	}

	if debugState.IsEnabled(resources.DebugCategoryHitbox) {
		sections = append(sections, debugLegendSection{
			title: "Hitbox (Cmd+3)",
			lines: []legendLine{
				{"Red fill: Hurtbox slices (Aseprite)", color.RGBA{255, 0, 0, 160}},
				{"Green outline: Hitbox slices (attack)", color.RGBA{0, 255, 0, 255}},
				{"Blue outline: Block slices", color.RGBA{64, 160, 255, 255}},
				{"Yellow outline: Parry window", color.RGBA{255, 255, 0, 255}},
			},
		})
	}

	if debugState.IsEnabled(resources.DebugCategoryAI) {
		sections = append(sections, debugLegendSection{
			title: "AI (Cmd+4)",
			lines: []legendLine{
				{"Yellow boxes: Detection ranges", color.RGBA{255, 255, 0, 255}},
				{"Yellow lines: Target connections", color.RGBA{255, 255, 0, 255}},
				{"White text: AI state machine", color.RGBA{255, 255, 255, 255}},
			},
		})
	}

	if debugState.IsEnabled(resources.DebugCategoryBehaviorTree) {
		sections = append(sections, debugLegendSection{
			title: "Behavior Tree (Cmd+7)",
			lines: []legendLine{
				{"Green: Running nodes", color.RGBA{0, 255, 0, 255}},
				{"Blue: Success nodes", color.RGBA{0, 150, 255, 255}},
				{"Red: Failure nodes", color.RGBA{255, 0, 0, 255}},
				{"Gray: Not executed", color.RGBA{128, 128, 128, 255}},
			},
		})
	}

	if debugState.IsEnabled(resources.DebugCategoryStats) {
		sections = append(sections, debugLegendSection{
			title: "Stats (Cmd+5)",
			lines: []legendLine{
				{"H:health S:stamina P:poise", color.RGBA{255, 255, 255, 255}},
			},
		})
	}

	if debugState.IsEnabled(resources.DebugCategoryAnim) {
		sections = append(sections, debugLegendSection{
			title: "Animation (Cmd+6)",
			lines: []legendLine{
				{"White text: Animation state and frame", color.RGBA{255, 255, 255, 255}},
			},
		})
	}

	if debugState.IsEnabled(resources.DebugCategoryTile) {
		sections = append(sections, debugLegendSection{
			title: "Tile Grid (Shift+T)",
			lines: []legendLine{
				{"Faint white outlines: Tile boundaries", color.RGBA{255, 255, 255, 60}},
				{"White text: Tile coordinates", color.RGBA{255, 255, 255, 200}},
			},
		})
	}

	if len(sections) == 0 {
		return // No active debug overlays
	}

	// Calculate total lines needed (title + lines for each section + spacing)
	totalLines := 0
	for _, section := range sections {
		totalLines += 1 + len(section.lines) // Title + lines
	}
	totalLines += len(sections) - 1 // Spacing between sections

	// Measure the widest line to determine box width
	maxWidth := 0.0
	for _, section := range sections {
		// Measure title
		w, _ := text.Measure(section.title, face, 0)
		if w > maxWidth {
			maxWidth = w
		}
		// Measure each line
		for _, line := range section.lines {
			w, _ = text.Measure(line.text, face, 0)
			if w > maxWidth {
				maxWidth = w
			}
		}
	}

	// Box dimensions with padding (in device pixels)
	paddingX := 8.0 * dpr
	paddingY := 4.0 * dpr
	lineHeightDevice := logicalLineHeight * dpr
	spacingDevice := lineHeightDevice * 0.5 // Half line height for section spacing
	boxWidth := maxWidth + paddingX*2
	boxHeight := float64(totalLines)*lineHeightDevice + float64(len(sections)-1)*spacingDevice + paddingY*2

	// Position: top-right with margin (in device pixels)
	rightMargin := 10.0 * dpr
	topMargin := 10.0 * dpr
	boxX := float64(screenBounds.Dx()) - boxWidth - rightMargin
	boxY := topMargin

	// Draw background box (translucent gray, matching textbox style)
	bgImage := ebiten.NewImage(int(boxWidth), int(boxHeight))
	bgColor := color.RGBA{64, 64, 64, 220} // Darker gray, more translucent
	bgImage.Fill(bgColor)

	bgOpts := &ebiten.DrawImageOptions{}
	bgOpts.GeoM.Translate(boxX, boxY)
	screen.DrawImage(bgImage, bgOpts)

	// Draw each section with its lines, right-aligned
	textX := boxX + boxWidth - paddingX // Right edge minus padding
	textY := boxY + paddingY

	for i, section := range sections {
		// Draw section title in white
		opts := &text.DrawOptions{}
		opts.GeoM.Translate(textX, textY)
		opts.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
		opts.PrimaryAlign = text.AlignEnd
		text.Draw(screen, section.title, face, opts)
		textY += lineHeightDevice

		// Draw each line with its color
		for _, line := range section.lines {
			opts := &text.DrawOptions{}
			opts.GeoM.Translate(textX, textY)
			opts.ColorScale.ScaleWithColor(line.color)
			opts.PrimaryAlign = text.AlignEnd
			text.Draw(screen, line.text, face, opts)
			textY += lineHeightDevice
		}

		// Add spacing between sections (except after last section)
		if i < len(sections)-1 {
			textY += spacingDevice
		}
	}
}

// drawECSOverlayIfEnabled always calls drawECSOverlay, which internally
// determines what to show based on active debug categories and flags.
func drawECSOverlayIfEnabled(world *ecs.World, screen *ebiten.Image, vp *components.ViewPort) {
	// Always call - function checks internally what to render
	drawECSOverlay(world, screen, vp)
}

// drawDebugKeyboardShortcuts renders keyboard shortcuts help in bottom-left corner.
// This replaces the old DrawDebugHelp from legend.go with proper DPR scaling.
func drawDebugKeyboardShortcuts(screen *ebiten.Image, vp *components.ViewPort) {
	if vp == nil {
		return
	}

	face := fonts.FiveByFiveFont
	dpr := vp.DPR

	helpText := []string{
		"Debug Shortcuts:",
		"Cmd+1: Collision",
		"Cmd+2: Physics",
		"Cmd+3: Hitbox",
		"Cmd+4: AI",
		"Cmd+5: Stats",
		"Cmd+6: Animation",
		"Cmd+0: Toggle All",
		"",
		"Shift+T: Tile Grid",
		"Cmd+Shift+0: Profiler",
		"Cmd+Shift+1: ECS Overlay",
	}

	// Calculate dimensions
	paddingDevice := 10.0 * dpr
	logicalLineHeight := 14.0
	lineHeightDevice := logicalLineHeight * dpr

	// Measure max width
	maxWidth := 0.0
	for _, line := range helpText {
		w, _ := text.Measure(line, face, 0)
		if w > maxWidth {
			maxWidth = w
		}
	}

	bgWidth := maxWidth + paddingDevice*2
	bgHeight := float64(len(helpText))*lineHeightDevice + paddingDevice*2

	// Position: bottom-left with margin
	bgX := paddingDevice
	bgY := float64(screen.Bounds().Dy()) - bgHeight - paddingDevice

	// Draw background (matching grey style)
	bgImg := ebiten.NewImage(int(bgWidth), int(bgHeight))
	bgImg.Fill(color.RGBA{64, 64, 64, 200})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(bgX, bgY)
	screen.DrawImage(bgImg, op)

	// Draw help text
	textX := bgX + paddingDevice
	textY := bgY + paddingDevice

	for i, line := range helpText {
		textColor := color.RGBA{200, 200, 200, 255}
		if i == 0 {
			textColor = color.RGBA{255, 255, 255, 255} // Title in white
		}

		opts := &text.DrawOptions{}
		opts.GeoM.Translate(textX, textY)
		opts.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, line, face, opts)
		textY += lineHeightDevice
	}
}
