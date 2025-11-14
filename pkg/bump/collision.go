package bump

import "math"

// Collision detection functions.

// detectCollision performs collision detection between two rectangles.
// Returns a Collision struct and whether a collision was detected.
func detectCollision(rect1, rect2 Rect, goal Vec2) (*Collision, bool) {
	col := &Collision{}
	col.Move = Vec2{goal.X - rect1.X, goal.Y - rect1.Y}
	col.ItemRect, col.OtherRect = rect1, rect2
	interRect := rectDiff(rect1, rect2)

	if !detectCollisionFirstPhase(interRect, rect1, col) {
		return col, false
	}
	if !col.Overlaps {
		return col, true
	}

	if (col.Move == Vec2{}) {
		p := rectNearestCorner(interRect, Vec2{})
		col.Normal = Vec2{math.Copysign(1, p.X), math.Copysign(1, p.Y)}
		if math.Abs(p.X) < math.Abs(p.Y) {
			p.Y = 0
			col.Normal.Y = 0
		} else {
			p.X = 0
			col.Normal.X = 0
		}
		col.Touch = Vec2{rect1.X + p.X, rect1.Y + p.Y}
	} else {
		i1, _, normal, found := lineSegmentIntersection(interRect, Vec2{}, col.Move)
		if !found {
			return col, false
		}
		col.Normal = normal
		col.Touch = Vec2{rect1.X + col.Move.X*i1, rect1.Y + col.Move.Y*i1}
	}

	return col, true
}

// detectCollisionFirstPhase performs the first phase of collision detection.
// Determines if there's an overlap or intersection and populates collision data.
func detectCollisionFirstPhase(interRect, rect1 Rect, col *Collision) bool {
	collided := false
	if rectContainsPoint(interRect, Vec2{}) {
		collided = true
		p := rectNearestCorner(interRect, Vec2{})
		wi, hi := math.Min(rect1.W, math.Abs(p.X)), math.Min(rect1.H, math.Abs(p.Y))
		col.Intersection = -wi * hi
		col.Overlaps = true
	} else {
		i1, i2, normal, ok := lineSegmentIntersection(interRect, Vec2{}, col.Move)
		if ok && i1 < 1 && math.Abs(i1-i2) >= Epsilon && (i1 > -Epsilon || i1 == 0 && i2 > 0) {
			collided = true
			col.Normal = normal
			col.Intersection = i1
			col.Overlaps = false
			col.Touch = Vec2{rect1.X + col.Move.X*i1, rect1.Y + col.Move.Y*i1}
		}
	}

	return collided
}
