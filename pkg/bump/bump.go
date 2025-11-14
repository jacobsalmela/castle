package bump

import (
	"slices"
	"sync"
)

// Space is the spatial hash for collision detection.
type Space struct {
	Responses   map[ColType]Response
	rects       map[Item]Rect
	tags        map[Item][]Tag
	searchSpace map[location]map[Item]bool
	cellSize    float64
	mutex       sync.RWMutex
}

// NewSpace allocates a new Space with default collision responses.
func NewSpace() *Space {
	space := &Space{
		rects:       map[Item]Rect{},
		tags:        map[Item][]Tag{},
		searchSpace: map[location]map[Item]bool{},
		cellSize:    CellSize,
	}
	space.Responses = defaultResponses(space)
	return space
}

// Set inserts or updates an item's bounding rectangle and tags in the spatial hash.
func (s *Space) Set(item Item, rect Rect, tags ...Tag) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	oldRect, existed := s.rects[item]
	s.rects[item] = rect

	cells, oldCells := s.cellCoords(rect), s.cellCoords(oldRect)
	if slices.Equal(cells, oldCells) && (len(tags) == 0 || slices.Equal(tags, s.tags[item])) {
		return
	}
	if len(tags) > 0 {
		s.tags[item] = tags
	}

	// Remove from old locations
	if existed {
		for _, tag := range s.allTags(item) {
			for _, cell := range oldCells {
				s.removeFromLocation(location{tag, cell}, item)
			}
		}
	}

	// Add to new locations
	for _, tag := range s.allTags(item) {
		for _, cell := range cells {
			s.addToLocation(location{tag, cell}, item)
		}
	}
}

// Rect returns the current rectangle of an item.
func (s *Space) Rect(item Item) Rect {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.rects[item]
}

// Has checks if an item exists in the space with the given tags.
func (s *Space) Has(item Item, tags ...Tag) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, cell := range s.cellCoords(s.rects[item]) {
		for _, tag := range tags {
			if !s.searchSpace[location{tag, cell}][item] {
				return false
			}
		}
	}

	return true
}

// Remove deletes an item from the space.
func (s *Space) Remove(item Item) {
	if item == nil {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if rect, ok := s.rects[item]; ok {
		for _, cell := range s.cellCoords(rect) {
			for _, tag := range s.allTags(item) {
				s.removeFromLocation(location{tag, cell}, item)
			}
		}
	}
	delete(s.tags, item)
	delete(s.rects, item)
}

// Move moves an item to a new position, resolving collisions.
func (s *Space) Move(item Item, targetGoal Vec2, filter Filter, tags ...Tag) (Vec2, []*Collision) {
	goal, cols := s.Check(item, targetGoal, filter, tags...)
	rect := s.Rect(item)
	rect.X, rect.Y = goal.X, goal.Y
	s.Set(item, rect)

	return goal, cols
}

// Check tests a movement goal for collisions and returns the new goal and collisions.
func (s *Space) Check(item Item, goal Vec2, filter Filter, tags ...Tag) (Vec2, []*Collision) {
	if filter == nil {
		filter = DefaultResponseFilter
	}

	visited := map[Item]bool{item: true}
	visitedFilter := func(item, other Item) (ColType, bool) {
		if visited[other] {
			return 0, false
		}

		return filter(item, other)
	}

	projectedCols := s.Project(item, s.Rect(item), goal, visitedFilter, tags...)
	var cols []*Collision
	for len(projectedCols) > 0 {
		col := projectedCols[0]
		visited[col.Other] = true
		goal, projectedCols = s.Responses[col.Type](goal, col, visitedFilter, tags...)
		cols = append(cols, col)
	}

	return goal, cols
}

// Project computes movement collisions for item moving from rect to goal based on filter and tags.
func (s *Space) Project(item Item, rect Rect, goal Vec2, filter Filter, tags ...Tag) []*Collision {
	if filter == nil {
		filter = DefaultResponseFilter
	}
	if len(tags) == 0 {
		tags = []Tag{""}
	}

	candidates := s.collectCandidates(item, rect, tags)

	var cols []*Collision
	for _, other := range candidates {
		responseName, ok := filter(item, other)
		if !ok {
			continue
		}

		col, detected := detectCollision(rect, s.Rect(other), goal)
		if !detected {
			continue
		}

		col.Item, col.Other = item, other
		col.Type = responseName
		cols = append(cols, col)
	}

	sortCollisions(cols)
	return cols
}

// Query retrieves all items that overlap or are near the given rect.
func (s *Space) Query(rect Rect, filter SelectFilter, tags ...Tag) []*Collision {
	if filter == nil {
		filter = func(_ Item) bool { return true }
	}
	projectFilter := func(_, other Item) (ColType, bool) { return 0, filter(other) }

	return s.Project(nil, rect, Vec2{rect.X, rect.Y}, projectFilter, tags...)
}

// Helper methods to reduce code duplication.

// allTags returns all tags for an item including the empty tag.
func (s *Space) allTags(item Item) []Tag {
	return append(s.tags[item], "")
}

// removeFromLocation removes an item from a location and cleans up empty locations.
func (s *Space) removeFromLocation(loc location, item Item) {
	delete(s.searchSpace[loc], item)
	if len(s.searchSpace[loc]) == 0 {
		delete(s.searchSpace, loc)
	}
}

// addToLocation adds an item to a location, creating the map if needed.
func (s *Space) addToLocation(loc location, item Item) {
	if s.searchSpace[loc] == nil {
		s.searchSpace[loc] = map[Item]bool{}
	}
	s.searchSpace[loc][item] = true
}

// cellCoords returns the cell coordinates covered by a rectangle.
func (s *Space) cellCoords(rect Rect) []cell {
	cx, cy := int(rect.X/s.cellSize), int(rect.Y/s.cellSize)
	cr, cb := int((rect.X+rect.W)/s.cellSize), int((rect.Y+rect.H)/s.cellSize)

	coords := make([]cell, 0, (cr+1-cx)*(cb+1-cy))
	for y := cy; y <= cb; y++ {
		for x := cx; x <= cr; x++ {
			coords = append(coords, cell{x, y})
		}
	}

	return coords
}

// collectCandidates collects candidate items from cells for collision detection.
func (s *Space) collectCandidates(item Item, rect Rect, tags []Tag) []Item {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	seen := make(map[Item]bool)
	var items []Item

	for _, cell := range s.cellCoords(rect) {
		for _, tag := range tags {
			for other := range s.searchSpace[location{tag, cell}] {
				if other == item || seen[other] {
					continue
				}
				seen[other] = true
				items = append(items, other)
			}
		}
	}

	return items
}

// sortCollisions sorts collisions by intersection time and distance.
func sortCollisions(cols []*Collision) {
	slices.SortFunc(cols, func(a, b *Collision) int {
		if a.Intersection != b.Intersection {
			if a.Intersection < b.Intersection {
				return -1
			}
			return 1
		}

		ir := a.ItemRect
		distA := rectSquareDistance(ir, a.OtherRect)
		distB := rectSquareDistance(ir, b.OtherRect)
		if distA < distB {
			return -1
		}
		return 1
	})
}
