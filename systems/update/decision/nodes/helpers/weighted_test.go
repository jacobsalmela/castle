package helpers

import (
	"testing"
)

func TestSelectWeightedIndex(t *testing.T) {
	tests := []struct {
		name    string
		weights []float64
		// For deterministic tests, we check if result is valid (in range)
		wantValid bool
		wantRange [2]int // [min, max] valid indices
	}{
		{
			name:      "Equal weights",
			weights:   []float64{1.0, 1.0, 1.0},
			wantValid: true,
			wantRange: [2]int{0, 2},
		},
		{
			name:      "Single weight",
			weights:   []float64{5.0},
			wantValid: true,
			wantRange: [2]int{0, 0},
		},
		{
			name:      "Zero weights",
			weights:   []float64{0, 0, 0},
			wantValid: false,
			wantRange: [2]int{-1, -1},
		},
		{
			name:      "Mixed weights",
			weights:   []float64{0.5, 1.0, 2.0},
			wantValid: true,
			wantRange: [2]int{0, 2},
		},
		{
			name:      "Empty weights",
			weights:   []float64{},
			wantValid: false,
			wantRange: [2]int{-1, -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectWeightedIndex(tt.weights)

			if !tt.wantValid {
				if result != -1 {
					t.Errorf("expected -1 for invalid weights, got %d", result)
				}
				return
			}

			if result < tt.wantRange[0] || result > tt.wantRange[1] {
				t.Errorf("result %d out of valid range [%d, %d]", result, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}

func TestSelectWeightedIndex_Distribution(t *testing.T) {
	// Statistical test: run many times and check distribution roughly matches weights
	weights := []float64{1.0, 2.0, 3.0} // Should get ~16.7%, ~33.3%, ~50% distribution
	iterations := 10000
	counts := make([]int, len(weights))

	for i := 0; i < iterations; i++ {
		idx := SelectWeightedIndex(weights)
		if idx >= 0 && idx < len(counts) {
			counts[idx]++
		}
	}

	// Check that index 2 (weight 3.0) got roughly 50% of selections
	// Allow 10% margin of error (45-55%)
	expectedPercent := 0.5
	actualPercent := float64(counts[2]) / float64(iterations)

	if actualPercent < expectedPercent-0.1 || actualPercent > expectedPercent+0.1 {
		t.Errorf("index 2 should be selected ~50%% of the time, got %.1f%%", actualPercent*100)
	}

	// Check that index 0 (weight 1.0) got roughly 16.7% of selections
	expectedPercent = 1.0 / 6.0 // ~16.7%
	actualPercent = float64(counts[0]) / float64(iterations)

	if actualPercent < expectedPercent-0.05 || actualPercent > expectedPercent+0.05 {
		t.Errorf("index 0 should be selected ~16.7%% of the time, got %.1f%%", actualPercent*100)
	}
}

func TestShuffleWeightedIndices(t *testing.T) {
	tests := []struct {
		name    string
		weights []float64
		wantLen int
	}{
		{
			name:    "Three equal weights",
			weights: []float64{1.0, 1.0, 1.0},
			wantLen: 3,
		},
		{
			name:    "Single weight",
			weights: []float64{5.0},
			wantLen: 1,
		},
		{
			name:    "Empty weights",
			weights: []float64{},
			wantLen: 0,
		},
		{
			name:    "Mixed weights",
			weights: []float64{0.5, 1.0, 2.0, 3.0},
			wantLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShuffleWeightedIndices(tt.weights)

			if len(result) != tt.wantLen {
				t.Errorf("expected length %d, got %d", tt.wantLen, len(result))
			}

			// Check that all indices are present exactly once
			seen := make(map[int]bool)
			for _, idx := range result {
				if idx < 0 || idx >= len(tt.weights) {
					t.Errorf("invalid index %d (weights length: %d)", idx, len(tt.weights))
				}
				if seen[idx] {
					t.Errorf("duplicate index %d in result", idx)
				}
				seen[idx] = true
			}
		})
	}
}

func TestShuffleWeightedIndices_WeightPriority(t *testing.T) {
	// Statistical test: items with higher weights should appear earlier on average
	weights := []float64{1.0, 10.0, 1.0} // Middle item has much higher weight
	iterations := 1000
	positionSums := make([]int, len(weights))

	for i := 0; i < iterations; i++ {
		result := ShuffleWeightedIndices(weights)
		for pos, idx := range result {
			positionSums[idx] += pos
		}
	}

	// Index 1 (weight 10.0) should have lowest average position (appear earlier)
	avgPos1 := float64(positionSums[1]) / float64(iterations)
	avgPos0 := float64(positionSums[0]) / float64(iterations)
	avgPos2 := float64(positionSums[2]) / float64(iterations)

	if avgPos1 > avgPos0 || avgPos1 > avgPos2 {
		t.Errorf("index 1 (high weight) should appear earlier on average. Avg positions: [%.2f, %.2f, %.2f]",
			avgPos0, avgPos1, avgPos2)
	}
}

func TestRemoveIndex(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		index    int
		expected []int
	}{
		{
			name:     "Remove first",
			slice:    []int{1, 2, 3, 4},
			index:    0,
			expected: []int{2, 3, 4},
		},
		{
			name:     "Remove middle",
			slice:    []int{1, 2, 3, 4},
			index:    2,
			expected: []int{1, 2, 4},
		},
		{
			name:     "Remove last",
			slice:    []int{1, 2, 3, 4},
			index:    3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "Single element",
			slice:    []int{1},
			index:    0,
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeIndex(tt.slice, tt.index)

			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d: expected %d, got %d", i, tt.expected[i], v)
				}
			}
		})
	}
}
