package ecs

import (
	"fmt"
	"game/entities"
	"game/pkg/config"
	"reflect"
	"sort"
	"testing"
	"time"
)

// Dummy component types used only in benchmarks/tests
type Pos struct{ X, Y float64 }
type Vel struct{ X, Y float64 }
type TagA struct{}
type TagB struct{}

// populate deterministically adds components to entities so that
// queries have non-trivial (but predictable) cardinalities.
func populate(w *World, n int) {
	for i := 1; i <= n; i++ {
		e := w.NewEntity()
		// pattern: every 2nd has Pos, every 3rd has Vel, every 100th has TagA
		if i%2 == 0 {
			w.AddComponent(e, &Pos{X: float64(i), Y: float64(i)})
		}
		if i%3 == 0 {
			w.AddComponent(e, &Vel{X: 1, Y: 2})
		}
		if i%100 == 0 {
			w.AddComponent(e, &TagA{})
		}
		if i%257 == 0 {
			w.AddComponent(e, &TagB{})
		}
	}
}

// naiveEntitiesWith mimics the original O(n*m) implementation by scanning
// all entities and checking component presence directly on the components map.
func naiveEntitiesWith(w *World, comps ...any) []entities.EntityId {
	w.mu.RLock()
	defer w.mu.RUnlock()

	needed := make([]reflect.Type, len(comps))
	for i, c := range comps {
		needed[i] = reflect.TypeOf(c)
	}

	var result []entities.EntityId
	for eid, compMap := range w.components {
		ok := true
		for _, t := range needed {
			if _, has := compMap[t]; !has {
				ok = false
				break
			}
		}
		if ok {
			result = append(result, eid)
		}
	}
	return result
}

// compareUnordered verifies two slices contain same elements (entity ids) ignoring order.
func compareUnordered(a, b []entities.EntityId) bool {
	if len(a) != len(b) {
		return false
	}
	sa := make([]uint64, len(a))
	sb := make([]uint64, len(b))
	for i := range a {
		sa[i] = uint64(a[i])
		sb[i] = uint64(b[i])
	}
	sort.Slice(sa, func(i, j int) bool { return sa[i] < sa[j] })
	sort.Slice(sb, func(i, j int) bool { return sb[i] < sb[j] })
	for i := range sa {
		if sa[i] != sb[i] {
			return false
		}
	}
	return true
}

func TestEntitiesWithIndexMatchesNaive(t *testing.T) {
	w := NewWorld(&config.Config{})
	populate(w, 2000)

	a := w.EntitiesWith((*Pos)(nil), (*Vel)(nil))
	b := naiveEntitiesWith(w, (*Pos)(nil), (*Vel)(nil))
	if !compareUnordered(a, b) {
		t.Fatalf("mismatch between indexed and naive queries: got %d vs %d", len(a), len(b))
	}

	// Also test a single-component query
	a = w.EntitiesWith((*TagA)(nil))
	b = naiveEntitiesWith(w, (*TagA)(nil))
	if !compareUnordered(a, b) {
		t.Fatalf("mismatch for single-component query: got %d vs %d", len(a), len(b))
	}
}

func BenchmarkEntitiesWithImplementations(b *testing.B) {
	sizes := []int{1000, 5000, 20000}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("Indexed/N=%d", n), func(b *testing.B) {
			w := NewWorld(&config.Config{})
			populate(w, n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = w.EntitiesWith((*Pos)(nil), (*Vel)(nil))
			}
		})

		b.Run(fmt.Sprintf("Naive/N=%d", n), func(b *testing.B) {
			w := NewWorld(&config.Config{})
			populate(w, n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = naiveEntitiesWith(w, (*Pos)(nil), (*Vel)(nil))
			}
		})
	}
}

// Large-run tests: measure runtime for indexed vs naive on very large entity counts.
// These are intentionally heavy; they will run as part of the normal test target
// unless you run `go test -short` to skip them.
func runLargeComparison(t *testing.T, n int) {
	if testing.Short() {
		t.Skip("skipping large comparison in short mode")
	}
	t.Logf("running large comparison with N=%d", n)
	// indexed
	w := NewWorld(&config.Config{})
	populate(w, n)
	start := time.Now()
	_ = w.EntitiesWith((*Pos)(nil), (*Vel)(nil))
	durIndexed := time.Since(start)

	// naive
	// rebuild world to ensure same layout
	w2 := NewWorld(&config.Config{})
	populate(w2, n)
	start = time.Now()
	_ = naiveEntitiesWith(w2, (*Pos)(nil), (*Vel)(nil))
	durNaive := time.Since(start)

	t.Logf("N=%d indexed=%s naive=%s", n, durIndexed.String(), durNaive.String())
}

func TestEntitiesWithLarge100k(t *testing.T) { runLargeComparison(t, 100000) }
func TestEntitiesWithLarge500k(t *testing.T) { runLargeComparison(t, 500000) }
