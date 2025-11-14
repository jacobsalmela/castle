// Package helpers provides utility functions for behavior tree node implementations.
package helpers

import (
	"game/components"
	"math"
)

// GetCenter returns the center point of a transform.
func GetCenter(t *components.Transform) (x, y float64) {
	return t.X + t.W/2, t.Y + t.H/2
}

// CalculateDistance returns the Euclidean distance between two transforms' centers.
func CalculateDistance(t1, t2 *components.Transform) float64 {
	x1, y1 := GetCenter(t1)
	x2, y2 := GetCenter(t2)
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// CalculateDirection returns the direction vector from t1 to t2.
func CalculateDirection(t1, t2 *components.Transform) (dx, dy float64) {
	x1, y1 := GetCenter(t1)
	x2, y2 := GetCenter(t2)
	return x2 - x1, y2 - y1
}

// IsTargetOnRight returns true if t2 is to the right of t1.
func IsTargetOnRight(t1, t2 *components.Transform) bool {
	x1, _ := GetCenter(t1)
	x2, _ := GetCenter(t2)
	return x2 > x1
}
