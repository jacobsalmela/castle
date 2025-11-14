package ui

import "math"

// ViewPort describes the various coordinate spaces we use so that code can
// be explicit about what units it is working in.
//
// ┌────────────── logical (design pixels) ────────────────┐
// │ LW × LH:  the fixed-resolution virtual back‑buffer    │
// └────────────────────────────────────────────────────────┘
//
//	⬇  *Scale  (fit to window, keep AR)
//
// ┌────────────── point space (window points) ────────────┐
// │ PW × PH:  the rectangle inside the window that gets   │
// │           the scaled logical buffer.  UI hit‑testing  │
// │           usually happens here.                       │
// └────────────────────────────────────────────────────────┘
//
//	⬇  *DPR  (monitor/device pixel ratio)
//
// ┌────────────── device pixel space ─────────────────────┐
// │ PX × PY:  actual framebuffer pixels the GPU draws     │
// └────────────────────────────────────────────────────────┘
//
//	OffsetX/OffsetY are the letterbox bars in point space.
type ViewPort struct {
	// Logical (design pixels)
	// World and physics math
	LW, LH float64 // logical (inside) size (the design resolution for sprites and game elements)

	// Point (window points)
	// UI hit-testing, mouse/touch coords
	PW, PH  float64 // Point‑space rectangle inside the window
	OffsetX float64 // offset to center the viewport in the physical screen
	OffsetY float64 // offset to center the viewport in the physical screen
	Scale   float64 // Scale factor from logical to point space (points-per-logical-pixel)

	// Device pixel (framebuffer)
	// Low-level draws and screenshots
	DPR    float64 // Device‑pixel ratio reported by the monitor / OS( device-pixels-per-point)
	PX, PY float64 // Framebuffer size in device pixels
	// SW, SH float64 // screen/pixel/point/physical size (locgical * scale)

	// Gameplay Helpers
	OW, OH    float64 // outside width/height
	HorizonY  float64 // horizon line is pre-converted to pixels so no *DPR is needed in draw/update
	FarthestZ float64 // viewport
}

// logicalX/logicalY are in your logical pixel space (same space as vp.LW, etc).
func (vp *ViewPort) LogicalToDevice(logicalX, logicalY float64) (dx, dy, ds float64) {
	// Convert logical -> device using a single combined scale (Scale * DPR).
	ds = vp.Scale * vp.DPR
	// apply logical scaling and offset, then multiply by DPR and round
	dx = math.Round((logicalX*vp.Scale + vp.OffsetX) * vp.DPR)
	dy = math.Round((logicalY*vp.Scale + vp.OffsetY) * vp.DPR)
	return dx, dy, ds
}
