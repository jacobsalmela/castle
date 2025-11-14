//go:build !release

package world

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/resources"
	"game/systems/draw/debug/primitives"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// DrawHitboxDebug renders combat collision overlays.
// Enable with Cmd+3. Shows:
// - Red outline: hurtbox slices (vulnerable area)
// - Green outline: hitbox slices (attack window)
// - Blue outline: blockbox slices (blocking window)
// - Yellow outline: parry block slices (parry timing)
// - Red fading fills: recent hitbox activations (damage taken)
func DrawHitboxDebug(world *ecs.World, screen *ebiten.Image, cam primitives.CameraProvider) {
	if !primitives.IsDebugEnabled(world, resources.DebugCategoryHitbox) {
		return
	}

	coords := primitives.NewWorldToScreen(cam)

	drawAnimationSliceOverlays(world, screen, coords)

	if events := ecs.Resource[resources.HitboxEventQueue](world); events != nil {
		drawHitboxEvents(screen, events.Recent(), coords)
	}
}

func drawHitboxOutline(screen *ebiten.Image, x, y, w, h float64, col color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}

	screenX := float32(x)
	screenY := float32(y)
	width := float32(w)
	height := float32(h)

	vector.StrokeLine(screen, screenX, screenY, screenX+width, screenY, 1, col, false)
	vector.StrokeLine(screen, screenX+width, screenY, screenX+width, screenY+height, 1, col, false)
	vector.StrokeLine(screen, screenX+width, screenY+height, screenX, screenY+height, 1, col, false)
	vector.StrokeLine(screen, screenX, screenY+height, screenX, screenY, 1, col, false)
}

type sliceRenderContext struct {
	screen      *ebiten.Image
	transform   *components.Transform
	anim        *components.Animation
	frame       int
	flipX       bool
	flipY       bool
	cameraX     float64
	cameraY     float64
	frameWidth  float64
	frameHeight float64
	geom        ebiten.GeoM
	hasGeom     bool
}

func renderSlice(ctx sliceRenderContext, sliceName string, outline color.RGBA, fill *color.RGBA) {
	frames := ctx.anim.SliceMap[sliceName]
	if len(frames) == 0 {
		return
	}

	rect, ok := frames[ctx.frame]
	if !ok || rect.W <= 0 || rect.H <= 0 {
		return
	}

	screenX, screenY, width, height, ok := projectSliceRect(ctx, rect)
	if !ok || width <= 0 || height <= 0 {
		return
	}

	if fill != nil {
		primitives.FillRect(ctx.screen, screenX, screenY, width, height, *fill)
	}
	drawHitboxOutline(ctx.screen, screenX, screenY, width, height, outline)
}

func projectSliceRect(ctx sliceRenderContext, rect bump.Rect) (float64, float64, float64, float64, bool) {
	if ctx.hasGeom {
		return projectSliceWithGeom(ctx.geom, rect)
	}

	offsetX := rect.X
	if ctx.flipX {
		if ctx.frameWidth > 0 {
			offsetX = ctx.frameWidth - rect.X - rect.W
		}
		offsetX += ctx.anim.OX + ctx.anim.OXFlip
	} else {
		offsetX += ctx.anim.OX
	}

	offsetY := rect.Y
	if ctx.flipY {
		if ctx.frameHeight > 0 {
			offsetY = ctx.frameHeight - rect.Y - rect.H
		}
		offsetY += ctx.anim.OY + ctx.anim.OYFlip
	} else {
		offsetY += ctx.anim.OY
	}

	screenX := ctx.transform.X + offsetX - ctx.cameraX
	screenY := ctx.transform.Y + offsetY - ctx.cameraY
	return screenX, screenY, rect.W, rect.H, true
}

func projectSliceWithGeom(geom ebiten.GeoM, rect bump.Rect) (float64, float64, float64, float64, bool) {
	corners := [][2]float64{
		{rect.X, rect.Y},
		{rect.X + rect.W, rect.Y},
		{rect.X, rect.Y + rect.H},
		{rect.X + rect.W, rect.Y + rect.H},
	}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, pt := range corners {
		x, y := geom.Apply(pt[0], pt[1])
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}
	width := maxX - minX
	height := maxY - minY
	if width <= 0 || height <= 0 {
		return 0, 0, 0, 0, false
	}
	return minX, minY, width, height, true
}

func animGeom(c *components.Render, w, h float64, opGeom ebiten.GeoM) ebiten.GeoM {
	var sx, sy, dx, dy float64 = 1, 1, 0, 0
	if c.FlipX {
		sx, dx = -1, math.Floor(w/2)+dx
	}
	if c.FlipY {
		sy, dy = -1, math.Floor(h/2)+dy
	}
	var geom ebiten.GeoM
	geom.Scale(sx, sy)
	geom.Translate(c.X+dx, c.Y+dy)
	geom.Concat(opGeom)
	return geom
}

func drawAnimationSliceOverlays(world *ecs.World, screen *ebiten.Image, coords primitives.WorldToScreen) {
	entityIDs := world.EntitiesWith((*components.Animation)(nil), (*components.Transform)(nil))
	for _, eid := range entityIDs {
		anim := ecs.GetComponent[components.Animation](world, eid)
		transform := ecs.GetComponent[components.Transform](world, eid)
		if anim == nil || transform == nil || anim.Data == nil {
			continue
		}
		if anim.SliceMap == nil || len(anim.SliceMap) == 0 {
			continue
		}

		frame := currentAnimationFrame(anim)
		if frame < 0 {
			continue
		}

		flipX, flipY := animationFlipState(world, entities.EntityId(eid))
		frameBounds := anim.Data.FrameBoundaries().Rectangle()
		ctx := sliceRenderContext{
			screen:      screen,
			transform:   transform,
			anim:        anim,
			frame:       frame,
			flipX:       flipX,
			flipY:       flipY,
			cameraX:     coords.CameraX,
			cameraY:     coords.CameraY,
			frameWidth:  float64(frameBounds.Dx()),
			frameHeight: float64(frameBounds.Dy()),
		}

		if render := ecs.GetComponent[components.Render](world, eid); render != nil && render.Image != nil {
			entityPos := ebiten.GeoM{}
			entityPos.Translate(math.Ceil(transform.X-coords.CameraX), math.Ceil(transform.Y-coords.CameraY))
			imgSize := render.Image.Bounds().Size()
			ctx.geom = animGeom(render, float64(imgSize.X), float64(imgSize.Y), entityPos)
			ctx.hasGeom = true
		}

		// Hurtbox: outline only (fill shows on hit via drawHitboxEvents)
		renderSlice(ctx, components.HurtboxSliceName, primitives.HurtboxOutline, nil)
		renderSlice(ctx, components.BlockSliceName, primitives.BlockboxOutline, nil)
		renderSlice(ctx, components.HitboxSliceName, primitives.HitboxOutline, nil)
	}
}

func currentAnimationFrame(anim *components.Animation) int {
	if anim == nil {
		return -1
	}
	if anim.Data != nil {
		if frame := anim.Data.CurrentFrame; frame >= 0 {
			return frame
		}
	}
	if anim.Frame >= 0 {
		return anim.Frame
	}
	return -1
}

func animationFlipState(world *ecs.World, eid entities.EntityId) (bool, bool) {
	if render := ecs.GetComponent[components.Render](world, eid); render != nil {
		return render.FlipX, render.FlipY
	}
	if facing := ecs.GetComponent[components.Facing](world, eid); facing != nil {
		return facing.FlipX, false
	}
	return false, false
}

func drawHitboxEvents(screen *ebiten.Image, events []resources.HitboxEvent, coords primitives.WorldToScreen) {
	now := time.Now()
	const lifetime = 1000 * time.Millisecond

	for _, event := range events {
		age := now.Sub(event.Timestamp)
		ageRatio := float64(age) / float64(lifetime)
		alpha := 1.0 - ageRatio

		if alpha <= 0 {
			continue
		}

		screenX, screenY := coords.TransformF64(event.X, event.Y)

		brightness := uint8(255 * alpha)
		highlightColor := color.RGBA{brightness, 0, 0, brightness}

		primitives.FillRect(screen, screenX, screenY, event.W, event.H, highlightColor)
	}
}
