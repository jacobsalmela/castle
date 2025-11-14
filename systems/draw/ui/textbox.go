package ui

import (
	"game/assets"
	"game/assets/fonts"
	"game/components"
	"game/ecs"
	"game/pkg/config"
	"game/resources"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	// Textbox images - loaded lazily on first use
	textboxIndicatorImage *ebiten.Image
	textboxAdvanceImage   *ebiten.Image
)

// ensureTextboxImagesLoaded lazily loads textbox images from the unified asset system.
// This is called once on first RenderTextbox() call to avoid init-time loading issues.
func ensureTextboxImagesLoaded() {
	if textboxIndicatorImage != nil {
		return // Already loaded
	}

	// Try to get from unified asset system first
	textboxIndicatorImage = assets.GetSpriteImage("textboxindicator")
	textboxAdvanceImage = assets.GetSpriteImage("textboxadvance")

	// Fallback: try to load directly if not in asset registry
	if textboxIndicatorImage == nil {
		var err error
		textboxIndicatorImage, _, err = ebitenutil.NewImageFromFileSystem(assets.FS, "images/ui/textboxindicator.png")
		if err != nil {
			// Create minimal fallback
			textboxIndicatorImage = ebiten.NewImage(8, 8)
		}
	}

	if textboxAdvanceImage == nil {
		var err error
		textboxAdvanceImage, _, err = ebitenutil.NewImageFromFileSystem(assets.FS, "images/ui/textboxadvance.png")
		if err != nil {
			// Create minimal fallback
			textboxAdvanceImage = ebiten.NewImage(8, 8)
		}
	}
}

// ComputeTextboxLines wraps and paginates text for textbox display.
// Returns the processed text with line breaks, number of lines, and max advance pages.
// This is a pure function that can be called to prepare TextboxData for rendering.
// Uses NanoFont (DPR-scaled) for text measurements.
func ComputeTextboxLines(txt string, advanceW float64, cfg *config.Config) (processedText string, lineCount int, advanceMax int) {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Normalize line breaks and paragraph markers
	txt = strings.ReplaceAll(txt, "\n\n", " \r ")
	txt = strings.ReplaceAll(txt, "\n", " \n ")

	var lines []string
	currentLine := ""
	for word := range strings.SplitSeq(txt, " ") {
		if count, ok := controlBreakCount(word, len(lines), cfg); ok {
			appendBreaks(&lines, &currentLine, count)
			continue
		}

		maxLineWidth := cfg.Textbox.LineWidth
		if len(lines)%int(cfg.Textbox.MaxLines) == int(cfg.Textbox.MaxLines)-1 {
			maxLineWidth -= advanceW
		}
		// Use NanoFont for text measurements
		testText := currentLine + " " + word
		if len(currentLine) == 0 {
			testText = word
		}
		w, _ := text.Measure(testText, fonts.NanoFont, 0)
		if w > maxLineWidth {
			lines = append(lines, currentLine)
			currentLine = ""
		}
		if len(currentLine) > 0 {
			currentLine += " "
		}
		currentLine += word
	}
	lines = append(lines, currentLine)

	joined := strings.Join(lines, "\n")
	lineCount = len(lines)
	advanceMax = int(math.Ceil(float64(lineCount)/cfg.Textbox.MaxLines)) - 1
	return joined, lineCount, advanceMax
}

// controlBreakCount returns the number of "virtual" newline additions based on a control word.
// "\n" inserts a single newline; "\r" pads until the end of the current page.
func controlBreakCount(word string, linesSoFar int, cfg *config.Config) (int, bool) {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	switch word {
	case "\n":
		return 1, true
	case "\r":
		return int(cfg.Textbox.MaxLines) - (linesSoFar % int(cfg.Textbox.MaxLines)) - 1, true
	default:
		return 0, false
	}
}

// appendBreaks appends new lines and clears the current composing line.
func appendBreaks(lines *[]string, currentLine *string, count int) {
	if count <= 0 {
		return
	}
	for range count {
		*lines = append(*lines, *currentLine)
		*currentLine = ""
	}
}

// PrepareTextboxData initializes a TextboxData component with processed text.
// Call this after creating a TextboxData to prepare it for rendering.
func PrepareTextboxData(data *components.TextboxData) {
	if data == nil {
		return
	}

	// Use default config for textbox preparation (no world available at this point)
	cfg := config.NewDefaultConfig()

	// Ensure textbox images are loaded before using them
	ensureTextboxImagesLoaded()

	advanceW := float64(textboxAdvanceImage.Bounds().Size().X)
	data.ProcessedText, data.Lines, data.AdvanceMax = ComputeTextboxLines(data.Text, advanceW, cfg)
	data.BoxHeight = min(int(cfg.Textbox.MaxLines), data.Lines) * int(cfg.Textbox.LineHeight)
}

// RenderTextbox draws a single textbox using the ECS render queue in logical pixel space.
// The textbox renders to the logical screen buffer (BEFORE viewport scaling), so it uses
// logical coordinates and will be scaled to device pixels automatically by the viewport.
// Size adapts to text content, and the box is centered above the entity when using indicators.
func RenderTextbox(world *ecs.World, queue *resources.RenderQueue, data *components.TextboxData, cam *resources.Camera) {
	if world == nil || queue == nil || data == nil || !data.Active || data.ProcessedText == "" {
		return
	}

	// Get config from ECS world
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Ensure textbox images are loaded (lazy initialization)
	ensureTextboxImagesLoaded()

	// Get viewport to check DPR (for font selection only)
	vp := ecs.Resource[components.ViewPort](world)
	if vp == nil {
		return
	}

	// === LOGICAL PIXEL SIZING ===
	// Work in logical pixels - viewport will scale to device pixels
	paddingX := cfg.Textbox.PaddingX
	paddingY := cfg.Textbox.PaddingY
	lineHeight := cfg.Textbox.LineHeight

	// Calculate text-dependent box dimensions (in logical pixels)
	visibleLines := min(int(cfg.Textbox.MaxLines), data.Lines)
	boxWidth := calculateDynamicTextboxWidth(data, vp, cfg)
	boxHeight := float64(visibleLines)*lineHeight + paddingY*2

	// === LOGICAL PIXEL POSITIONING ===
	var boxX, boxY float64
	if data.Indicator {
		// Center above entity sprite (in logical space)
		boxX, boxY = calculateCenteredPosition(data, cam, boxWidth, boxHeight, 1.0, cfg) // DPR=1.0 for logical space
	} else {
		// Fixed position (logical pixels)
		boxX = cfg.Textbox.BoxX
		boxY = cfg.Textbox.BoxY
	}

	// Create background image (in logical pixels)
	bgImage := ebiten.NewImage(int(boxWidth), int(boxHeight))
	bgImage.Fill(cfg.Textbox.BackgroundColor.ToColor())

	// Draw background box
	boxGeoM := ebiten.GeoM{}
	boxGeoM.Translate(boxX, boxY) // Normal map punch-out
	queue.Push(resources.RenderCommand{
		Image:      bgImage,
		TargetType: resources.TargetNormal,
		Layer:      config.PipelineUILayer,
		GeoM:       boxGeoM,
		ColorScale: components.NormalMaskColor,
	})

	// Background box
	queue.Push(resources.RenderCommand{
		Image:      bgImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       boxGeoM,
	})

	// Render text lines in logical space
	layout := textboxLayout{
		boxX:       boxX,
		boxY:       boxY,
		boxWidth:   boxWidth,
		boxHeight:  boxHeight,
		paddingX:   paddingX,
		paddingY:   paddingY,
		lineHeight: lineHeight,
	}
	renderTextLines(queue, data, layout, cfg)

	// Draw advance indicator if needed
	if !data.AdvanceFlicker && data.AdvanceState < data.AdvanceMax {
		renderAdvanceIndicator(queue, layout, data)
	}

	// Draw position indicator if needed
	if data.Indicator {
		renderPositionIndicator(queue, data, cam, layout)
	}
}

// calculateCenteredPosition determines the logical pixel position for a textbox centered above an entity.
// Returns boxX and boxY in logical pixels, clamped to screen bounds.
// Works in logical space since textbox renders before viewport scaling.
func calculateCenteredPosition(data *components.TextboxData, cam *resources.Camera, boxWidth, boxHeight, dpr float64, cfg *config.Config) (boxX, boxY float64) {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	camX, camY := cam.Position()

	// Calculate entity center in screen space (logical pixels)
	entityCenterX := data.EntityX + data.EntityW/2 - camX
	entityTopY := data.EntityY - camY
	entityBottomY := data.EntityY + data.EntityH - camY

	// Center the box horizontally on the entity
	boxX = entityCenterX - boxWidth/2

	// Position box above entity with margin (logical pixels)
	marginY := cfg.Textbox.BoxMarginY
	boxY = entityTopY - boxHeight - marginY

	// Clamp to screen bounds (in logical pixels)
	screenWidth := cfg.Screen.Width
	screenHeight := cfg.Screen.Height
	minX := 10.0
	minY := 10.0
	maxX := screenWidth - 10.0 - boxWidth
	maxY := screenHeight - 10.0 - boxHeight

	boxX = max(minX, min(boxX, maxX))

	// If box doesn't fit above, try below
	if boxY < minY {
		boxYBelow := entityBottomY + marginY
		if boxYBelow+boxHeight <= maxY {
			boxY = boxYBelow
		} else {
			boxY = minY // Clamp to top if neither works
		}
	}

	return boxX, boxY
}

// textboxLayout holds DPR-scaled layout values for textbox rendering.
type textboxLayout struct {
	boxX, boxY          float64
	boxWidth, boxHeight float64
	paddingX, paddingY  float64
	lineHeight          float64
}

// calculateDynamicTextboxWidth calculates the box width based on actual text content.
// Returns the width in logical pixels needed to fit all visible lines plus padding.
func calculateDynamicTextboxWidth(data *components.TextboxData, vp *components.ViewPort, cfg *config.Config) float64 {
	if data == nil || data.ProcessedText == "" {
		return cfg.Textbox.BoxW // Fallback to default
	}

	// Get visible lines for current page
	var lines []string
	allLines := strings.Split(data.ProcessedText, "\n")
	linesPerPage := int(cfg.Textbox.MaxLines)
	startLine := data.AdvanceState * linesPerPage
	if startLine < 0 {
		startLine = 0
	}
	endLine := startLine + linesPerPage
	if endLine > len(allLines) {
		endLine = len(allLines)
	}
	if startLine < len(allLines) {
		lines = allLines[startLine:endLine]
	}

	// Get device-space font and measure each line
	face := fonts.DeviceFace(cfg.Textbox.FontSize, vp.DPR)
	maxWidth := 0.0
	for _, line := range lines {
		if line == "" {
			continue
		}
		w, _ := text.Measure(line, face, 0)
		if w > maxWidth {
			maxWidth = w
		}
	}

	// Convert device pixels to logical pixels
	// Device pixels = logical * scale * DPR, so logical = device / (scale * DPR)
	logicalWidth := maxWidth / (vp.Scale * vp.DPR)

	// Add padding (8px on each side)
	paddingX := cfg.Textbox.PaddingX
	boxWidth := logicalWidth + paddingX*2

	// Ensure minimum width for indicators
	minWidth := 60.0
	if boxWidth < minWidth {
		boxWidth = minWidth
	}

	return boxWidth
}

// renderTextLines is a no-op placeholder.
// Text rendering has been moved to RenderTextboxesDeviceSpace for crisp device-pixel fonts.
// This function is kept for API compatibility but does nothing.
func renderTextLines(queue *resources.RenderQueue, data *components.TextboxData, layout textboxLayout, cfg *config.Config) {
	// Text is now rendered in device space via RenderTextboxesDeviceSpace
	// Called after viewport scaling for pixel-perfect text
	_ = queue
	_ = data
	_ = layout
	_ = cfg
}

// renderAdvanceIndicator draws the page advance arrow in logical pixel space.
func renderAdvanceIndicator(queue *resources.RenderQueue, layout textboxLayout, data *components.TextboxData) {
	advanceSize := textboxAdvanceImage.Bounds().Size()

	// Position in bottom-right corner of box (logical pixels)
	advanceX := layout.boxX + layout.boxWidth - float64(advanceSize.X) - 8.0
	advanceY := layout.boxY + layout.boxHeight - float64(advanceSize.Y) - 4.0

	advanceGeoM := ebiten.GeoM{}
	advanceGeoM.Translate(advanceX, advanceY)

	queue.Push(resources.RenderCommand{
		Image:      textboxAdvanceImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       advanceGeoM,
	})
}

// renderPositionIndicator draws the arrow pointing to the entity in logical pixel space.
func renderPositionIndicator(queue *resources.RenderQueue, data *components.TextboxData, cam *resources.Camera, layout textboxLayout) {
	camX, _ := cam.Position()
	indicatorSize := textboxIndicatorImage.Bounds().Size()
	iw := float64(indicatorSize.X)

	// Calculate entity center in screen space (logical pixels)
	entityCenterX := data.EntityX + data.EntityW/2 - camX

	// Position indicator at bottom of box, horizontally aligned with entity
	ix := entityCenterX - iw/2

	// Clamp to box bounds
	ix = max(layout.boxX, min(ix, layout.boxX+layout.boxWidth-iw))

	// Position at bottom edge of box
	iy := layout.boxY + layout.boxHeight

	indicatorGeoM := ebiten.GeoM{}
	indicatorGeoM.Translate(ix, iy)

	// Draw indicator on screen
	queue.Push(resources.RenderCommand{
		Image:      textboxIndicatorImage,
		TargetType: resources.TargetScreen,
		Layer:      config.PipelineUILayer,
		GeoM:       indicatorGeoM,
	})

	// Normal map punch-out
	queue.Push(resources.RenderCommand{
		Image:      textboxIndicatorImage,
		TargetType: resources.TargetNormal,
		Layer:      config.PipelineUILayer,
		GeoM:       indicatorGeoM,
		ColorScale: components.NormalMaskColor,
	})
}

// RenderDismissedIndicator is a no-op placeholder.
// The "i" indicator is now rendered in device space via RenderDismissedIndicatorDeviceSpace.
// This function is kept for API compatibility but does nothing.
func RenderDismissedIndicator(world *ecs.World, queue *resources.RenderQueue, data *components.TextboxData, cam *resources.Camera) {
	// Text is now rendered in device space via RenderDismissedIndicatorDeviceSpace
	// Called after viewport scaling for pixel-perfect text
	_ = world
	_ = queue
	_ = data
	_ = cam
}

// RenderTextboxesDeviceSpace renders textbox text in device-space for crisp fonts.
// This is called AFTER viewport scaling to ensure pixel-perfect text rendering.
// The textbox backgrounds and indicators are rendered in logical space via RenderTextbox.
//
// Parameters:
//   - world: ECS world instance
//   - screen: Device screen (after viewport scaling)
//   - vp: Viewport for coordinate conversion
func RenderTextboxesDeviceSpace(world *ecs.World, screen *ebiten.Image, vp *components.ViewPort) {
	if world == nil || screen == nil || vp == nil {
		return
	}

	camera := ecs.Resource[resources.Camera](world)
	if camera == nil {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Get device-space font
	face := fonts.DeviceFace(cfg.Textbox.FontSize, vp.DPR)

	// Process all active textboxes
	for _, eid := range world.EntitiesWith((*components.TextboxData)(nil)) {
		textboxData := ecs.GetComponent[components.TextboxData](world, eid)
		if textboxData == nil || !textboxData.Active || textboxData.ProcessedText == "" {
			continue
		}

		// Calculate logical pixel box position (same as RenderTextbox)
		paddingX := cfg.Textbox.PaddingX
		paddingY := cfg.Textbox.PaddingY
		lineHeight := cfg.Textbox.LineHeight

		visibleLines := min(int(cfg.Textbox.MaxLines), textboxData.Lines)
		boxWidth := calculateDynamicTextboxWidth(textboxData, vp, cfg)
		boxHeight := float64(visibleLines)*lineHeight + paddingY*2

		var boxX, boxY float64
		if textboxData.Indicator {
			boxX, boxY = calculateCenteredPosition(textboxData, camera, boxWidth, boxHeight, 1.0, cfg)
		} else {
			boxX = cfg.Textbox.BoxX
			boxY = cfg.Textbox.BoxY
		}

		// Get visible lines for current page
		var lines []string
		if textboxData.ProcessedText != "" {
			allLines := strings.Split(textboxData.ProcessedText, "\n")
			linesPerPage := int(cfg.Textbox.MaxLines)
			startLine := textboxData.AdvanceState * linesPerPage
			if startLine < 0 {
				startLine = 0
			}
			endLine := startLine + linesPerPage
			if endLine > len(allLines) {
				endLine = len(allLines)
			}
			if startLine < len(allLines) {
				lines = allLines[startLine:endLine]
			}
		}

		// Calculate device-space line height based on font metrics
		// Use the font size scaled by DPR for consistent spacing
		deviceLineHeight := float64(cfg.Textbox.FontSize) * vp.DPR * 1.2

		// Render each line in device space
		for i, line := range lines {
			if line == "" {
				continue
			}

			// Calculate logical position for this line
			logicalX := boxX + paddingX
			logicalY := boxY + paddingY + float64(i)*lineHeight

			// Convert to device space
			dx, dy, _ := vp.LogicalToDevice(logicalX, logicalY)

			// Adjust Y for device-space line spacing
			// The logical lineHeight is scaled, but we want consistent device-pixel spacing
			dy = dy - float64(i)*lineHeight*vp.Scale*vp.DPR + float64(i)*deviceLineHeight

			// Draw text directly to device screen
			opts := &text.DrawOptions{}
			opts.GeoM.Translate(dx, dy)
			opts.ColorScale.ScaleWithColor(cfg.Textbox.TextColor.ToColor())
			text.Draw(screen, line, face, opts)
		}
	}
}

// RenderDismissedIndicatorDeviceSpace draws indicators in device space.
// Called after viewport scaling for crisp text.
func RenderDismissedIndicatorDeviceSpace(world *ecs.World, screen *ebiten.Image, vp *components.ViewPort) {
	if world == nil || screen == nil || vp == nil {
		return
	}

	camera := ecs.Resource[resources.Camera](world)
	if camera == nil {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	// Get indicator font from config (larger than text font)
	face := fonts.DeviceFace(cfg.Textbox.IndicatorFontSize, vp.DPR)
	camX, camY := camera.Position()

	for _, eid := range world.EntitiesWith((*components.TextboxData)(nil)) {
		textboxData := ecs.GetComponent[components.TextboxData](world, eid)
		if textboxData == nil {
			continue
		}

		// Only render for dismissed, inactive textboxes
		if !textboxData.Dismissed || textboxData.Active {
			continue
		}

		// Calculate logical position (above entity)
		centerX := textboxData.EntityX + textboxData.EntityW/2 - camX
		indicatorY := textboxData.EntityY - camY - 12

		// Convert to device space
		dx, dy, _ := vp.LogicalToDevice(centerX, indicatorY)

		// Measure text width for centering
		indicatorText := cfg.Textbox.IndicatorText
		textWidth, _ := text.Measure(indicatorText, face, 0)

		// Draw centered
		opts := &text.DrawOptions{}
		opts.GeoM.Translate(dx-textWidth/2, dy)
		opts.ColorScale.ScaleWithColor(cfg.Textbox.IndicatorColor.ToColor())
		text.Draw(screen, indicatorText, face, opts)
	}
}
