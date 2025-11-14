//go:build !release

package world

import (
	"fmt"
	"game/assets/fonts"
	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/resources"
	"game/systems/draw/debug/primitives"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawBehaviorTreeDebug renders behavior tree visualization for AI entities.
// Enable with Cmd+7. Shows:
//   - Green text: Running nodes
//   - Blue text: Success nodes
//   - Red text: Failure nodes
//   - Gray text: Not executed yet
func DrawBehaviorTreeDebug(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryBehaviorTree) {
		return
	}

	// Get DPR
	dpr := 1.0
	if vp := ecs.Resource[components.ViewPort](world); vp != nil {
		dpr = vp.DPR
	}

	// Collect debug info for the first AI with a tree
	debugInfos := collectFirstBehaviorDebugInfos(world)
	if len(debugInfos) == 0 {
		return
	}

	face := fonts.FiveByFiveFont
	logicalFontHeight := 8.0
	logicalLineHeight := logicalFontHeight + 2.0
	lineHeightDevice := logicalLineHeight * dpr
	indentLogical := 6.0
	indentDevice := indentLogical * dpr

	// Build node texts and measure max width
	nodeTexts, maxWidth := buildNodeTextsAndMaxWidth(debugInfos, face, indentDevice)

	screenW := float64(screen.Bounds().Dx())
	screenH := float64(screen.Bounds().Dy())
	maxAllowed := math.Min(screenW*0.4, 400.0*dpr)

	paddingX := 8.0 * dpr
	paddingY := 6.0 * dpr
	innerMax := maxAllowed - paddingX*2

	boxWidth := maxWidth + paddingX*2
	if boxWidth > maxAllowed {
		boxWidth = maxAllowed
	}
	boxHeight := float64(len(debugInfos))*lineHeightDevice + paddingY*2

	// Fixed top-left placement
	boxX := 10.0 * dpr
	boxY := 10.0 * dpr
	if boxX+boxWidth > screenW-10.0*dpr {
		boxX = screenW - boxWidth - 10.0*dpr
	}
	if boxY+boxHeight > screenH-10.0*dpr {
		boxY = screenH - boxHeight - 10.0*dpr
	}

	// Draw background box
	bgImg := ebiten.NewImage(int(boxWidth), int(boxHeight))
	bgImg.Fill(primitives.BackgroundDark)
	bgOp := &ebiten.DrawImageOptions{}
	bgOp.GeoM.Translate(boxX, boxY)
	screen.DrawImage(bgImg, bgOp)

	// Draw lines
	ctx := behaviorDrawCtx{
		boxX:         boxX,
		boxY:         boxY,
		paddingX:     paddingX,
		paddingY:     paddingY,
		lineHeight:   lineHeightDevice,
		indentDevice: indentDevice,
		innerMax:     innerMax,
		dpr:          dpr,
	}
	drawBehaviorLines(screen, debugInfos, nodeTexts, face, ctx)
}

// collectFirstBehaviorDebugInfos returns the debug info slice for the first AI
// entity that has a behavior tree with debug info.
func collectFirstBehaviorDebugInfos(world *ecs.World) []*ai.DebugNodeInfo {
	for _, eid := range world.EntitiesWith((*components.AI)(nil)) {
		aiComp := ecs.GetComponent[components.AI](world, eid)
		if aiComp == nil || aiComp.BehaviorTree == nil {
			continue
		}
		infos := ai.CollectDebugInfo(aiComp.BehaviorTree)
		if len(infos) > 0 {
			return infos
		}
	}
	return nil
}

// buildNodeTextsAndMaxWidth prepares the display strings and returns the
// maximum measured width (device pixels) accounting for indentation.
func buildNodeTextsAndMaxWidth(infos []*ai.DebugNodeInfo, face *text.GoTextFace, indentDevice float64) ([]string, float64) {
	nodeTexts := make([]string, 0, len(infos))
	maxWidth := 0.0
	for _, info := range infos {
		nodeText := fmt.Sprintf("%s (%s)", info.Name, info.LastStatus.String())
		nodeTexts = append(nodeTexts, nodeText)
		w, _ := text.Measure(nodeText, face, 0)
		totalW := w + float64(info.Depth)*indentDevice
		if totalW > maxWidth {
			maxWidth = totalW
		}
	}
	return nodeTexts, maxWidth
}

// truncateToFit shortens s to fit within available width when rendered with
// face. Leaves space for an ellipsis if truncated.
func truncateToFit(s string, face *text.GoTextFace, available float64) string {
	w, _ := text.Measure(s, face, 0)
	if w <= available {
		return s
	}
	ell := "..."
	for len(s) > 0 {
		s = s[:len(s)-1]
		tw, _ := text.Measure(s+ell, face, 0)
		if tw <= available {
			return s + ell
		}
	}
	return ""
}

// behaviorDrawCtx bundles draw parameters used by drawBehaviorLines.
type behaviorDrawCtx struct {
	boxX         float64
	boxY         float64
	paddingX     float64
	paddingY     float64
	lineHeight   float64
	indentDevice float64
	innerMax     float64
	dpr          float64
}

// drawBehaviorLines draws each behavior node line inside the boxed overlay.
func drawBehaviorLines(screen *ebiten.Image, infos []*ai.DebugNodeInfo, nodeTexts []string, face *text.GoTextFace, ctx behaviorDrawCtx) {
	for i, info := range infos {
		nodeText := nodeTexts[i]
		available := ctx.innerMax - float64(info.Depth)*ctx.indentDevice
		if available < 20*ctx.dpr {
			available = 20 * ctx.dpr
		}
		nodeText = truncateToFit(nodeText, face, available)

		textX := ctx.boxX + ctx.paddingX + float64(info.Depth)*ctx.indentDevice
		textY := ctx.boxY + ctx.paddingY + float64(i)*ctx.lineHeight

		var textColor color.RGBA
		switch info.LastStatus {
		case ai.Running:
			textColor = primitives.BTRunning
		case ai.Success:
			textColor = primitives.BTSuccess
		case ai.Failure:
			textColor = primitives.BTFailure
		default:
			textColor = primitives.BTNotExecuted
		}

		if info.Depth > 0 {
			vx1 := float32(textX - 8.0*ctx.dpr)
			vy := float32(textY + 3.0*ctx.dpr)
			vx2 := float32(textX - 2.0*ctx.dpr)
			vector.StrokeLine(screen, vx1, vy, vx2, vy, 1, textColor, false)
		}

		opts := &text.DrawOptions{}
		opts.GeoM.Translate(textX, textY)
		opts.ColorScale.ScaleWithColor(textColor)
		text.Draw(screen, nodeText, face, opts)
	}
}
