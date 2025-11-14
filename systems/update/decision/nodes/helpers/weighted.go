package helpers

import "math/rand"

// SelectWeightedIndex returns a random index based on weights.
// Returns -1 if totalWeight <= 0.
func SelectWeightedIndex(weights []float64) int {
	totalWeight := sumWeights(weights)
	if totalWeight <= 0 {
		return -1
	}

	randomValue := rand.Float64() * totalWeight
	accumulated := 0.0

	for i, weight := range weights {
		accumulated += weight
		if randomValue <= accumulated {
			return i
		}
	}

	return len(weights) - 1
}

// ShuffleWeightedIndices returns indices shuffled by weight priority.
// Higher weight = more likely to appear earlier in result.
func ShuffleWeightedIndices(weights []float64) []int {
	n := len(weights)
	result := make([]int, n)
	available := make([]int, n)
	remainingWeights := make([]float64, n)

	// Initialize available indices and weights
	for i := 0; i < n; i++ {
		available[i] = i
		remainingWeights[i] = weights[i]
	}

	// Select weighted random for each position
	for i := 0; i < n; i++ {
		selected := SelectWeightedIndex(remainingWeights)
		if selected == -1 {
			// No weight left, take remaining in order
			copy(result[i:], available)
			break
		}

		result[i] = available[selected]
		available = removeIndex(available, selected)
		remainingWeights = removeIndex(remainingWeights, selected)
	}

	return result
}

// sumWeights returns the sum of all weights.
func sumWeights(weights []float64) float64 {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	return total
}

// removeIndex removes an element at the specified index from a slice.
func removeIndex[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}
