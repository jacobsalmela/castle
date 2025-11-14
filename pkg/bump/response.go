package bump

// Collision response handlers.

// defaultResponses returns the default response functions for a new Space.
func defaultResponses(space *Space) map[ColType]Response {
	return map[ColType]Response{
		Touch:     touchResponse,
		Cross:     crossResponse(space),
		RectSlide: rectSlideResponse(space),
		Slide:     slideResponse(space),
	}
}

// touchResponse stops at the collision point.
func touchResponse(_ Vec2, col *Collision, _ Filter, _ ...Tag) (Vec2, []*Collision) {
	return col.Touch, nil
}

// crossResponse passes through the collision (ghost collision).
func crossResponse(space *Space) Response {
	return func(goal Vec2, col *Collision, filter Filter, tags ...Tag) (Vec2, []*Collision) {
		return goal, space.Project(col.Item, col.ItemRect, goal, filter, tags...)
	}
}

// rectSlideResponse slides along rectangular surfaces.
func rectSlideResponse(space *Space) Response {
	return func(goal Vec2, col *Collision, filter Filter, tags ...Tag) (Vec2, []*Collision) {
		col.PreviousGoal = goal
		if col.Move != (Vec2{}) {
			if col.Normal.X != 0 {
				goal.X = col.Touch.X
			} else {
				goal.Y = col.Touch.Y
			}
		}
		moved := Rect{col.Touch.X, col.Touch.Y, col.ItemRect.W, col.ItemRect.H, col.ItemRect.Type}
		return goal, space.Project(col.Item, moved, goal, filter, tags...)
	}
}

// slideResponse slides along surfaces including slopes.
func slideResponse(space *Space) Response {
	rectSlide := rectSlideResponse(space)
	return func(goal Vec2, col *Collision, filter Filter, tags ...Tag) (Vec2, []*Collision) {
		if col.OtherRect.Type == Full {
			return rectSlide(goal, col, filter, tags...)
		}
		goal = handleSlopeSlide(col, goal)
		return goal, nil
	}
}

// handleSlopeSlide handles sliding along sloped surfaces.
func handleSlopeSlide(col *Collision, goal Vec2) Vec2 {
	col.PreviousGoal = goal
	col.Normal = Vec2{}
	col.Touch.Y = goal.Y

	pivotL := goal.X + col.ItemRect.W*SlopePivot
	pivotR := goal.X + col.ItemRect.W*(1-SlopePivot)

	switch col.OtherRect.Type {
	case TopRightSlope, TopLeftSlope:
		h := max(col.OtherRect.slopeHeight(pivotL), col.OtherRect.slopeHeight(pivotR))
		if goal.Y < h {
			goal.Y = h
			col.Normal = Vec2{0, 1}
		}
	case BottomRightSlope, BottomLeftSlope:
		h := min(col.OtherRect.slopeHeight(pivotL), col.OtherRect.slopeHeight(pivotR))
		if goal.Y > h-col.ItemRect.H {
			goal.Y = h - col.ItemRect.H
			col.Normal = Vec2{0, -1}
		}
	}

	return goal
}
