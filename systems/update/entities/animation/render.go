package animation

import (
	"image/color"
	"log"

	"game/components"
	"game/ecs"

	"github.com/hajimehoshi/ebiten/v2"
)

// UpdateRenderFromAnimation updates or creates a components.Render for any entity
// that has an Animation component. This syncs Animation → Render for the
// sprite rendering system.
func UpdateRenderFromAnimation(world *ecs.World) {
	if world == nil {
		return
	}

	// Query all entities with Animation component
	for _, eid := range world.EntitiesWith((*components.Animation)(nil)) {
		anim := ecs.GetComponent[components.Animation](world, eid)
		if anim == nil || anim.Image == nil || anim.Data == nil {
			continue
		}

		// Get or create Render component
		render := ecs.GetComponent[components.Render](world, eid)
		if render == nil {
			render = &components.Render{}
			world.AddComponent(eid, render)
		}

		// Get Facing component for FlipX/FlipY state (Pure ECS pattern)
		facing := ecs.GetComponent[components.Facing](world, eid)

		// Sync Animation → Render
		toRenderFromAnimation(anim, render, facing)
	}
}

// toRenderFromAnimation updates r from the given Animation component using
// animation-style geometry and subimage. FlipX/FlipY state is read from the
// Facing component (Pure ECS pattern).
//
// This mirrors toRenderFromAnim() but works with Pure ECS Animation component.
func toRenderFromAnimation(anim *components.Animation, r *components.Render, facing *components.Facing) {
	if anim == nil || r == nil || anim.Image == nil || anim.Data == nil {
		return
	}

	// Check if current animation is valid before extracting frame
	// This can be nil during state transitions
	if anim.Data.CurrentAnimation != nil {
		// Update render image from current animation frame
		updateRenderImage(anim, r)
	}
	// Note: If CurrentAnimation is nil, we keep the previous frame's image
	// This prevents sprite disappearance during state transitions

	// ALWAYS update rendering properties, even if image update failed
	// This ensures the sprite can still be rendered with the previous frame

	// Set rendering properties
	r.Layer = anim.Layer

	// Read FlipX from Facing component
	if facing != nil {
		r.FlipX = facing.FlipX
	} else {
		r.FlipX = false
	}
	r.FlipY = false // FlipY not yet migrated to Facing component

	// Color scale
	r.ColorScale = anim.ColorScale
	if r.ColorScale == nil {
		r.ColorScale = color.White
	}

	// Apply sprite offsets (with flip offsets if facing opposite direction)
	r.X = anim.OX
	r.Y = anim.OY
	if facing != nil && facing.FlipX {
		r.X += anim.OXFlip
	}
	if false { // FlipY not migrated yet
		r.Y += anim.OYFlip
	}

	// Animation rendering flags
	r.FromAnim = true
	r.Normal = false
	r.R = 0
}

// updateRenderImage extracts the current animation frame and updates the Render component's image.
func updateRenderImage(anim *components.Animation, r *components.Render) {
	// Extract current frame subimage using Aseprite API
	frameBounds := anim.Data.FrameBoundaries().Rectangle()

	// Extract subimage for current frame
	sub, ok := anim.Image.SubImage(frameBounds).(*ebiten.Image)
	if !ok || sub == nil {
		log.Printf("ERROR: SubImage failed for %s frame %d (%s): frameBounds=%v",
			anim.FilesName, anim.Data.CurrentFrame, anim.State, frameBounds)

		// If this is the first frame (no image yet), use full sprite as emergency fallback
		if r.Image == nil {
			log.Printf("  → Using full sprite sheet as fallback for %s", anim.FilesName)
			r.Image = anim.Image
		}
		// Otherwise keep previous frame
		return
	}

	r.Image = sub

	// Debug: verify image was set
	if r.Image == nil {
		log.Printf("ERROR: r.Image is nil after SubImage for %s frame %d",
			anim.FilesName, anim.Data.CurrentFrame)
	}
}
