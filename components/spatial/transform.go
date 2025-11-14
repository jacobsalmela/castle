package spatial

// Transform represents position and size data.
type Transform struct {
	X, Y float64
	W, H float64
}

// Position returns the current X, Y coordinates.
func (t *Transform) Position() (float64, float64) {
	if t == nil {
		return 0, 0
	}
	return t.X, t.Y
}

// Rect returns the full bounding rectangle.
func (t *Transform) Rect() (float64, float64, float64, float64) {
	if t == nil {
		return 0, 0, 0, 0
	}
	return t.X, t.Y, t.W, t.H
}

// SetPosition updates the X, Y coordinates.
func (t *Transform) SetPosition(x, y float64) {
	if t != nil {
		t.X, t.Y = x, y
	}
}

// SetSize updates the width and height.
func (t *Transform) SetSize(w, h float64) {
	if t != nil {
		t.W, t.H = w, h
	}
}
