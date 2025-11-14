package resources

import (
	"game/pkg/bump"
	"game/pkg/config"
	"game/pkg/tween"
	"image"
	"math"
	"math/rand"
)

const (
	defaultStiffness = 9
)

type Recter interface {
	Rect() (float64, float64, float64, float64)
}

type Camera struct {
	x, y, w, h                           float64
	following                            Recter
	shakeElapsed                         float64
	shakeDuration                        float64
	shakeMagnitude                       float64
	borders                              *bump.Rect
	rooms                                []bump.Rect
	transitionElapsed                    float64
	transitionDuration                   float64
	transitionStartX, transitionStartY   float64
	transitionTargetX, transitionTargetY float64
	stiffness                            int
	betweenRooms                         bool
}

func NewCamera(w, h float64, cfg *config.Config) *Camera {
	if cfg == nil {
		panic("NewCamera requires a non-nil config")
	}
	duration := float64(cfg.Camera.TransitionDuration)
	// **FIX FOR ISSUE #3**: Initialize transitionElapsed to duration so isTransitioning() returns false
	// This prevents an unwanted transition on first camera follow/update
	return &Camera{
		w:                  w,
		h:                  h,
		transitionDuration: duration,
		transitionElapsed:  duration, // Start with no transition active
		stiffness:          defaultStiffness,
	}
}

func (c *Camera) Position() (float64, float64) { return c.x, c.y }
func (c *Camera) SetPosition(x, y float64)     { c.x, c.y = x, y }
func (c *Camera) SetRooms(rooms []bump.Rect)   { c.rooms = rooms }
func (c *Camera) Follow(e Recter) {
	// Don't re-follow if already following this target
	if c.following == e {
		return
	}

	// Reset shake but DON'T reset transitionElapsed - that would trigger an unwanted transition
	// **FIX FOR ISSUE #3**: Only reset transitionElapsed if it was already < duration (active transition)
	// If transitionElapsed >= duration, leave it alone (no transition active)
	c.shakeElapsed = 0
	c.shakeDuration = 0
	if c.transitionElapsed < c.transitionDuration {
		// Active transition in progress - reset it
		c.transitionElapsed = 0
	}
	// Don't reset transitionDuration - it's set in constructor from config
	c.following = e
	c.SetRoomBorders(false)
}

func (c *Camera) InFrame(e Recter, widthAddedMult, heightAddedMult float64) bool {
	frame := bump.Rect{X: c.x, Y: c.y, W: c.w, H: c.h}
	frame.X -= c.w * widthAddedMult
	frame.W += c.w * widthAddedMult * 2
	frame.Y -= c.h * heightAddedMult
	frame.H += c.h * heightAddedMult * 2
	x, y, w, h := e.Rect()
	rect := bump.Rect{X: x, Y: y, W: w, H: h}

	return bump.Overlaps(rect, frame)
}

func (c *Camera) Translate(x, y float64) {
	c.x += x
	c.y += y
}

func (c *Camera) Bounds() image.Rectangle {
	return image.Rect(int(c.x), int(c.y), int(c.x+c.w), int(c.y+c.h))
}

func (c *Camera) BoundsWithOffsetAndParallax(offsetX, offsetY int, parallaxX, parallaxY float64) image.Rectangle {
	if offsetX == 0 && offsetY == 0 && parallaxX == 1 && parallaxY == 1 {
		return c.Bounds()
	}
	x, y := c.Position()
	x *= parallaxX
	y *= parallaxY

	return image.Rect(int(x), int(y), int(x+c.w), int(y+c.h)).Add(image.Point{-offsetX, -offsetY})
}

func (c *Camera) Update(dt float64) {
	c.updateShake(dt)

	if c == nil || c.following == nil {
		return
	}

	ex, ey, w, h := c.following.Rect()
	x, y := ex+w/2-c.w/2, ey+h/2-c.h/2
	dx, dy := x-c.x, y-c.y

	c.SetRoomBorders(true)

	// During room transition, skip damper/following and let transition control position
	if c.isTransitioning() {
		c.applyTransition(dt)
	} else {
		// Normal following behavior when not transitioning
		c.Translate(damper(dt, dx, dy, c.stiffness))
		c.clampToBorders()
	}

	c.updateShake(dt)
}

func (c *Camera) isTransitioning() bool {
	return c.transitionElapsed < c.transitionDuration
}

func (c *Camera) applyTransition(dt float64) {
	if c.isTransitioning() {
		c.transitionElapsed += dt

		// Calculate progress (0 to 1)
		progress := c.transitionElapsed / c.transitionDuration
		if progress > 1 {
			progress = 1
			c.transitionElapsed = c.transitionDuration
		}

		// Apply easing (cubic ease-out)
		easedProgress := tween.EaseOutCubic(progress)

		// Interpolate position
		newX := tween.Lerp(c.transitionStartX, c.transitionTargetX, easedProgress)
		newY := tween.Lerp(c.transitionStartY, c.transitionTargetY, easedProgress)
		c.SetPosition(newX, newY)
	}
}

func (c *Camera) clampToBorders() {
	// Only clamp if not transitioning between rooms
	if c.borders != nil && !c.isTransitioning() {
		x := math.Max(math.Min(c.x, c.borders.X+c.borders.W-c.w), c.borders.X)
		y := math.Max(math.Min(c.y, c.borders.Y+c.borders.H-c.h), c.borders.Y)
		c.SetPosition(x, y)
	}
}

func (c *Camera) updateShake(dt float64) {
	if c != nil && c.shakeElapsed < c.shakeDuration {
		c.shakeElapsed += dt

		// Calculate progress (1 to 0 for shake decay)
		progress := 1.0 - (c.shakeElapsed / c.shakeDuration)
		if progress < 0 {
			progress = 0
			c.shakeElapsed = c.shakeDuration
		}

		// Apply shake with linear easing (decay from full magnitude to zero)
		progress = tween.EaseLinear(progress)
		shakex := (rand.Float64()*2 - 1) * c.shakeMagnitude * progress
		shakey := (rand.Float64()*2 - 1) * c.shakeMagnitude * progress
		c.Translate(shakex, shakey)
	}
}

func (c *Camera) Shake(duration float32, magnitude float64) {
	// Allow new shakes to override existing ones
	// This ensures stronger shakes (like door opening) can replace weaker attack shakes
	c.shakeElapsed = 0
	c.shakeDuration = float64(duration)
	c.shakeMagnitude = magnitude
}

func (c *Camera) SetRoomBorders(transition bool) {
	if c.following == nil || c.rooms == nil {
		return
	}

	// Don't process anything while a transition is active - let it complete
	if c.isTransitioning() {
		return
	}

	x, y, w, h := c.following.Rect()
	follow := bump.Rect{X: x + w/4, Y: y, W: w / 2, H: h}

	prevRoom := c.borders
	roomCount := 0
	var newRoom *bump.Rect
	for i, room := range c.rooms {
		if bump.Overlaps(follow, room) {
			roomCount++
			if &c.rooms[i] != prevRoom {
				newRoom = &c.rooms[i]
			}
		}
	}

	// Don't change borders while straddling rooms, but remember which room we're entering
	if c.betweenRooms && newRoom != nil {
		// Keep old border but remember the new room for transition
		c.borders = prevRoom
	} else if newRoom != nil {
		c.borders = newRoom
	}

	c.betweenRooms = roomCount > 1
	if roomCount == 0 {
		c.borders = nil
	}

	// Trigger transition when we detect a new room (even if currently straddling)
	// Only create if no transition is already active
	if transition && prevRoom != nil && newRoom != nil && prevRoom != newRoom && !c.isTransitioning() {
		// Store current position as start
		c.transitionStartX = c.x
		c.transitionStartY = c.y

		// Calculate target position clamped to new room
		c.transitionTargetX = math.Max(math.Min(c.x, newRoom.X+newRoom.W-c.w), newRoom.X)
		c.transitionTargetY = math.Max(math.Min(c.y, newRoom.Y+newRoom.H-c.h), newRoom.Y)

		// Only create transition if there's actual movement needed
		if c.transitionStartX != c.transitionTargetX || c.transitionStartY != c.transitionTargetY {
			// Start transition (elapsed begins at 0, duration is set)
			c.transitionElapsed = 0
			// Duration already set in constructor from config
		}
	}
}

func damper(dt, dx, dy float64, stiffness int) (float64, float64) {
	dts := dt * float64(stiffness)

	return dx * dts, dy * dts
}
