package bump

// Types, constants, and constructors for the bump collision detection library.

// Package-level configuration variables.
var (
	// CellSize is the size of each spatial hash grid cell.
	CellSize = 32.0
	// SlopePivot defines the pivot for slope collision (0 to 0.5, 0.5 is center of rect).
	SlopePivot = 0.5
)

// Constants for collision calculations.
const (
	// Epsilon is the floating-point margin of error for collision calculations.
	Epsilon = 1e-10
)

// Item is any value that can be tracked in the spatial hash.
type Item any

// Tag labels spatial cells for filtering collision partners.
type Tag string

// Rect defines an axis-aligned bounding box with position, size, and type.
type Rect struct {
	X, Y, W, H float64
	Type       RectType
}

// Vec2 represents a 2D vector or point.
type Vec2 struct{ X, Y float64 }

// Filter determines whether to consider a collision between two items and how.
type Filter func(item, other Item) (response ColType, selected bool)

// SelectFilter restricts items in a spatial query.
type SelectFilter func(item Item) bool

// Response applies collision resolution for a given collision event.
type Response func(goal Vec2, col *Collision, filter Filter, tags ...Tag) (newGoal Vec2, newCols []*Collision)

// Collision holds details about one detected collision between two items.
type Collision struct {
	Overlaps            bool
	Intersection        float64
	Move, Touch, Normal Vec2
	Item, Other         Item
	ItemRect, OtherRect Rect
	Type                ColType
	PreviousGoal        Vec2
}

// RectType defines the shape type of a rectangle for collision purposes.
type RectType uint

const (
	// Full is a standard axis-aligned bounding box.
	Full RectType = iota
	// TopRightSlope is a triangle slope where right angle is at the top right.
	TopRightSlope
	// TopLeftSlope is a triangle slope where right angle is at the top left.
	TopLeftSlope
	// BottomRightSlope is a triangle slope where right angle is at the bottom right.
	BottomRightSlope
	// BottomLeftSlope is a triangle slope where right angle is at the bottom left.
	BottomLeftSlope
)

// ColType defines the collision response type.
type ColType uint

const (
	// Touch stops at collision point.
	Touch ColType = iota
	// Cross passes through (ghost collision).
	Cross
	// RectSlide slides along rectangular surfaces only.
	RectSlide
	// Slide slides along surfaces (handles slopes).
	Slide
)

// Internal types for spatial hash.
type cell [2]int

type location struct {
	tag  Tag
	cell cell
}

// NewRect creates a new rectangle with the given position and size.
func NewRect(x, y, w, h float64) Rect { return Rect{X: x, Y: y, W: w, H: h} }

// DefaultResponseFilter returns Slide response for all collisions.
func DefaultResponseFilter(_, _ Item) (ColType, bool) { return Slide, true }

// NilFilter rejects all collisions.
func NilFilter(_, _ Item) (ColType, bool) { return 0, false }
