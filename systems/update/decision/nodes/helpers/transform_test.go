package helpers

import (
	"game/components"
	"math"
	"testing"
)

func TestGetCenter(t *testing.T) {
	tests := []struct {
		name     string
		t1       *components.Transform
		expectedX float64
		expectedY float64
	}{
		{
			name:     "Simple square",
			t1:       &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			expectedX: 5,
			expectedY: 5,
		},
		{
			name:     "Offset rectangle",
			t1:       &components.Transform{X: 10, Y: 20, W: 30, H: 40},
			expectedX: 25,
			expectedY: 40,
		},
		{
			name:     "Negative position",
			t1:       &components.Transform{X: -10, Y: -20, W: 10, H: 10},
			expectedX: -5,
			expectedY: -15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := GetCenter(tt.t1)
			if x != tt.expectedX || y != tt.expectedY {
				t.Errorf("expected (%f, %f), got (%f, %f)", tt.expectedX, tt.expectedY, x, y)
			}
		})
	}
}

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name     string
		t1       *components.Transform
		t2       *components.Transform
		expected float64
	}{
		{
			name:     "Same position",
			t1:       &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			t2:       &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			expected: 0,
		},
		{
			name:     "Horizontal distance",
			t1:       &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			t2:       &components.Transform{X: 30, Y: 0, W: 10, H: 10},
			expected: 30,
		},
		{
			name:     "Vertical distance",
			t1:       &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			t2:       &components.Transform{X: 0, Y: 40, W: 10, H: 10},
			expected: 40,
		},
		{
			name:     "Diagonal distance (3-4-5 triangle)",
			t1:       &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			t2:       &components.Transform{X: 30, Y: 40, W: 10, H: 10},
			expected: 50, // 3-4-5 triangle: sqrt(30^2 + 40^2) = 50
		},
		{
			name:     "Negative coordinates",
			t1:       &components.Transform{X: -10, Y: -10, W: 10, H: 10},
			t2:       &components.Transform{X: 20, Y: 30, W: 10, H: 10},
			expected: 50, // sqrt(30^2 + 40^2) = 50
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDistance(tt.t1, tt.t2)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateDirection(t *testing.T) {
	tests := []struct {
		name      string
		t1        *components.Transform
		t2        *components.Transform
		expectedDX float64
		expectedDY float64
	}{
		{
			name:      "Same position",
			t1:        &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			t2:        &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			expectedDX: 0,
			expectedDY: 0,
		},
		{
			name:      "Target to the right",
			t1:        &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			t2:        &components.Transform{X: 30, Y: 0, W: 10, H: 10},
			expectedDX: 30,
			expectedDY: 0,
		},
		{
			name:      "Target to the left",
			t1:        &components.Transform{X: 30, Y: 0, W: 10, H: 10},
			t2:        &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			expectedDX: -30,
			expectedDY: 0,
		},
		{
			name:      "Target above",
			t1:        &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			t2:        &components.Transform{X: 0, Y: 40, W: 10, H: 10},
			expectedDX: 0,
			expectedDY: 40,
		},
		{
			name:      "Target diagonal",
			t1:        &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			t2:        &components.Transform{X: 30, Y: 40, W: 10, H: 10},
			expectedDX: 30,
			expectedDY: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dx, dy := CalculateDirection(tt.t1, tt.t2)
			if math.Abs(dx-tt.expectedDX) > 0.001 || math.Abs(dy-tt.expectedDY) > 0.001 {
				t.Errorf("expected (%f, %f), got (%f, %f)", tt.expectedDX, tt.expectedDY, dx, dy)
			}
		})
	}
}

func TestIsTargetOnRight(t *testing.T) {
	tests := []struct {
		name     string
		t1       *components.Transform
		t2       *components.Transform
		expected bool
	}{
		{
			name:     "Target on right",
			t1:       &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			t2:       &components.Transform{X: 20, Y: 0, W: 10, H: 10},
			expected: true,
		},
		{
			name:     "Target on left",
			t1:       &components.Transform{X: 20, Y: 0, W: 10, H: 10},
			t2:       &components.Transform{X: 0, Y: 0, W: 10, H: 10},
			expected: false,
		},
		{
			name:     "Same X position",
			t1:       &components.Transform{X: 10, Y: 0, W: 10, H: 10},
			t2:       &components.Transform{X: 10, Y: 20, W: 10, H: 10},
			expected: false,
		},
		{
			name:     "Slightly to the right",
			t1:       &components.Transform{X: 10, Y: 0, W: 10, H: 10},
			t2:       &components.Transform{X: 10.1, Y: 0, W: 10, H: 10},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTargetOnRight(tt.t1, tt.t2)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
