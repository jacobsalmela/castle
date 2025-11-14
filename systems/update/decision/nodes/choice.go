package nodes

import (
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/systems/update/decision/nodes/helpers"
)

// WeightedChoice represents a single option in a weighted random selection.
type WeightedChoice struct {
	Weight float64
	Node   ai.Node
}

// RandomSelector creates a composite node that randomly selects one child based on weights.
// Unlike a standard Selector, this picks ONE child randomly (weighted) and executes only that child.
// This is equivalent to the legacy components.Choices.Play() behavior.
//
// How it works:
// - On first tick, randomly selects a child based on weights
// - Continues ticking that child until it completes (Success/Failure)
// - Returns the selected child's status
// - Resets selection when the child completes
//
// Use case: Enemy decides between multiple attack patterns with different probabilities.
// Example: 70% chance to use melee attack, 30% chance to use ranged attack.
type RandomSelector struct {
	Choices       []WeightedChoice
	selectedIndex int
	hasSelected   bool
}

func (r *RandomSelector) Tick(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
	if len(r.Choices) == 0 {
		return ai.Failure
	}

	// Select a child on first tick
	if !r.hasSelected {
		r.selectedIndex = r.selectWeightedRandom()
		r.hasSelected = true
	}

	// Tick the selected child
	status := r.Choices[r.selectedIndex].Node.Tick(world, eid, dt)

	// Reset when child completes
	if status != ai.Running {
		r.hasSelected = false
	}

	return status
}

func (r *RandomSelector) Reset() {
	r.hasSelected = false
	r.selectedIndex = 0
	for _, choice := range r.Choices {
		if resettable, ok := choice.Node.(ai.Resettable); ok {
			resettable.Reset()
		}
	}
}

// selectWeightedRandom picks a random index based on weights
func (r *RandomSelector) selectWeightedRandom() int {
	weights := make([]float64, len(r.Choices))
	for i, choice := range r.Choices {
		weights[i] = choice.Weight
	}

	idx := helpers.SelectWeightedIndex(weights)
	if idx == -1 {
		return 0 // Fallback to first choice
	}
	return idx
}

// NewRandomSelector creates a RandomSelector from a list of weighted choices.
func NewRandomSelector(choices ...WeightedChoice) *RandomSelector {
	return &RandomSelector{
		Choices:     choices,
		hasSelected: false,
	}
}

// WeightedSequence is like RandomSelector but tries children in random weighted order.
// Unlike RandomSelector which picks ONE child, this shuffles the children by weight
// and then executes them in sequence (like a normal Sequence).
//
// Use case: Randomize the order of actions while ensuring all execute.
type WeightedSequence struct {
	Choices       []WeightedChoice
	shuffledOrder []int
	current       int
	hasShuffled   bool
}

func (w *WeightedSequence) Tick(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
	if len(w.Choices) == 0 {
		return ai.Success
	}

	// Shuffle on first tick
	if !w.hasShuffled {
		w.shuffleWeighted()
		w.hasShuffled = true
		w.current = 0
	}

	// Execute children in shuffled order
	for w.current < len(w.shuffledOrder) {
		idx := w.shuffledOrder[w.current]
		status := w.Choices[idx].Node.Tick(world, eid, dt)

		switch status {
		case ai.Failure:
			w.hasShuffled = false
			w.current = 0
			return ai.Failure
		case ai.Running:
			return ai.Running
		case ai.Success:
			w.current++
		}
	}

	// All children succeeded
	w.hasShuffled = false
	w.current = 0
	return ai.Success
}

func (w *WeightedSequence) Reset() {
	w.hasShuffled = false
	w.current = 0
	w.shuffledOrder = nil
	for _, choice := range w.Choices {
		if resettable, ok := choice.Node.(ai.Resettable); ok {
			resettable.Reset()
		}
	}
}

// shuffleWeighted creates a weighted shuffle of indices
func (w *WeightedSequence) shuffleWeighted() {
	weights := make([]float64, len(w.Choices))
	for i, choice := range w.Choices {
		weights[i] = choice.Weight
	}

	w.shuffledOrder = helpers.ShuffleWeightedIndices(weights)
}

// NewWeightedSequence creates a WeightedSequence from a list of weighted choices.
func NewWeightedSequence(choices ...WeightedChoice) *WeightedSequence {
	return &WeightedSequence{
		Choices:     choices,
		hasShuffled: false,
	}
}
