package bump

import "math"

// Geometric utilities for collision detection.

// Overlaps checks if two rects overlap.
func Overlaps(r1, r2 Rect) bool {
	return rectContainsPoint(rectDiff(r1, r2), Vec2{})
}

// slopeHeight calculates the height of a slope at a given x position.
func (r Rect) slopeHeight(x float64) float64 {
	if r.Type == Full {
		return r.Y
	}
	lerp := math.Min(math.Max((x-r.X)/r.W, 0), 1)
	if r.Type == TopRightSlope || r.Type == BottomLeftSlope {
		return r.Y + lerp*r.H
	}

	return r.Y + (1-lerp)*r.H
}

// lineSegmentIntersection implements the Liang-Barsky algorithm for line-rectangle intersection.
// Returns intersection parameters (i1, i2), collision normal, and whether intersection occurred.
func lineSegmentIntersection(rect Rect, p1, p2 Vec2) (float64, float64, Vec2, bool) {
	dx, dy := p2.X-p1.X, p2.Y-p1.Y
	p := [4]float64{-dx, dx, -dy, dy} // left, right, top, bottom
	q := [4]float64{p1.X - rect.X, rect.X + rect.W - p1.X, p1.Y - rect.Y, rect.Y + rect.H - p1.Y}
	normals := [4]Vec2{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	i1, i2 := math.Inf(-1), math.Inf(1)
	normal := Vec2{}

	for i := range 4 {
		if checkParallelRejection(p[i], q[i]) {
			return 0, 0, Vec2{}, false
		}
		if p[i] == 0 {
			continue
		}

		var success bool
		i1, i2, success = updateIntersectionBounds(p[i], q[i], i1, i2, normals[i], &normal)
		if !success {
			return 0, 0, Vec2{}, false
		}
	}

	return i1, i2, normal, true
}

// checkParallelRejection checks if a parallel line segment should be rejected.
func checkParallelRejection(p, q float64) bool {
	return p == 0 && q <= 0
}

// updateIntersectionBounds updates the intersection bounds during Liang-Barsky algorithm.
func updateIntersectionBounds(p, q, i1, i2 float64, n Vec2, normal *Vec2) (float64, float64, bool) {
	r := q / p
	if p < 0 {
		if r > i2 {
			return i1, i2, false
		}
		if r > i1 {
			*normal = n
			return r, i2, true
		}
	} else {
		if r < i1 {
			return i1, i2, false
		}
		if r < i2 {
			return i1, r, true
		}
	}
	return i1, i2, true
}

// rectDiff computes the Minkowski Difference between two rectangles.
func rectDiff(r1, r2 Rect) Rect {
	return Rect{r2.X - r1.X - r1.W, r2.Y - r1.Y - r1.H, r1.W + r2.W, r1.H + r2.H, Full}
}

// rectContainsPoint checks if a rectangle contains a point (with epsilon tolerance).
func rectContainsPoint(r Rect, p Vec2) bool {
	return p.X-r.X > Epsilon && p.Y-r.Y > Epsilon && r.X+r.W-p.X > Epsilon && r.Y+r.H-p.Y > Epsilon
}

// rectSquareDistance returns squared distance between centers of two rects.
func rectSquareDistance(r1, r2 Rect) float64 {
	dx := r1.X - r2.X + (r1.W-r2.W)/2
	dy := r1.Y - r2.Y + (r1.H-r2.H)/2

	return dx*dx + dy*dy
}

// rectNearestCorner finds the nearest corner of r to point p.
func rectNearestCorner(rect Rect, p Vec2) Vec2 {
	nearest := func(x, a, b float64) float64 {
		if math.Abs(a-x) < math.Abs(b-x) {
			return a
		}

		return b
	}

	return Vec2{nearest(p.X, rect.X, rect.X+rect.W), nearest(p.Y, rect.Y, rect.Y+rect.H)}
}
