package ecs_test

import (
	"testing"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/pkg/config"
	"game/resources"
)

// TestNewWorld verifies world creation and resource registration
func TestNewWorld(t *testing.T) {
	world := ecs.NewWorld(nil)

	// Verify world created
	if world == nil {
		t.Fatal("NewWorld should not return nil")
	}

	// Verify core resources registered
	if cfg := ecs.Resource[config.Config](world); cfg == nil {
		t.Error("Config resource should be registered")
	}

	if camera := ecs.Resource[resources.Camera](world); camera == nil {
		t.Error("Camera resource should be registered")
	}

	if space := ecs.Resource[bump.Space](world); space == nil {
		t.Error("Space resource should be registered")
	}

	if mapRef := ecs.Resource[resources.MapRef](world); mapRef == nil {
		t.Error("MapRef resource should be registered")
	}

	if timeControl := ecs.Resource[resources.TimeControl](world); timeControl == nil {
		t.Error("TimeControl resource should be registered")
	}
}

// TestAddComponent verifies component addition and retrieval
func TestAddComponent(t *testing.T) {
	world := ecs.NewWorld(nil)
	eid := world.NewEntity()

	transform := &components.Transform{X: 10, Y: 20, W: 32, H: 32}
	world.AddComponent(eid, transform)

	// Verify component retrievable
	retrieved := ecs.GetComponent[components.Transform](world, eid)
	if retrieved == nil {
		t.Fatal("Component should be retrievable after adding")
	}

	// Verify correct values
	if retrieved.X != 10 || retrieved.Y != 20 {
		t.Errorf("got (%f, %f), want (10, 20)", retrieved.X, retrieved.Y)
	}

	if retrieved.W != 32 || retrieved.H != 32 {
		t.Errorf("got (%f, %f), want (32, 32)", retrieved.W, retrieved.H)
	}

	// Verify pointer semantics (shared state)
	transform.X = 100
	if retrieved.X != 100 {
		t.Error("Component should be stored as pointer (shared state)")
	}
}

// TestRemoveComponent verifies component removal
func TestRemoveComponent(t *testing.T) {
	world := ecs.NewWorld(nil)
	eid := world.NewEntity()

	world.AddComponent(eid, &components.Transform{X: 10, Y: 20, W: 32, H: 32})

	// Verify exists
	if !ecs.HasComponent[components.Transform](world, eid) {
		t.Fatal("Component should exist after adding")
	}

	// Remove
	ecs.RemoveComponent[components.Transform](world, eid)

	// Verify removed
	if ecs.HasComponent[components.Transform](world, eid) {
		t.Error("Component should not exist after removal")
	}

	retrieved := ecs.GetComponent[components.Transform](world, eid)
	if retrieved != nil {
		t.Error("GetComponent should return nil for removed component")
	}
}

// TestEntitiesWith verifies entity queries by component type
func TestEntitiesWith(t *testing.T) {
	world := ecs.NewWorld(nil)

	// Create 3 entities with different components
	e1 := world.NewEntity()
	world.AddComponent(e1, &components.Transform{X: 1, Y: 1, W: 10, H: 10})
	world.AddComponent(e1, &components.Health{Current: 100, Max: 100})

	e2 := world.NewEntity()
	world.AddComponent(e2, &components.Transform{X: 2, Y: 2, W: 10, H: 10})

	e3 := world.NewEntity()
	world.AddComponent(e3, &components.Health{Current: 50, Max: 50})

	// Query single component
	transformEntities := world.EntitiesWith((*components.Transform)(nil))
	if len(transformEntities) != 2 {
		t.Errorf("Expected 2 entities with Transform, got %d", len(transformEntities))
	}

	// Verify correct entities returned
	hasE1 := false
	hasE2 := false
	for _, eid := range transformEntities {
		if eid == e1 {
			hasE1 = true
		}
		if eid == e2 {
			hasE2 = true
		}
	}
	if !hasE1 || !hasE2 {
		t.Error("Transform query should return entities e1 and e2")
	}

	// Query multiple components (AND logic)
	bothEntities := world.EntitiesWith(
		(*components.Transform)(nil),
		(*components.Health)(nil),
	)
	if len(bothEntities) != 1 {
		t.Errorf("Expected 1 entity with both components, got %d", len(bothEntities))
	}
	if len(bothEntities) > 0 && bothEntities[0] != e1 {
		t.Errorf("Expected entity %d, got %d", e1, bothEntities[0])
	}

	// Query health only
	healthEntities := world.EntitiesWith((*components.Health)(nil))
	if len(healthEntities) != 2 {
		t.Errorf("Expected 2 entities with Health, got %d", len(healthEntities))
	}
}

// TestEntityLifecycle verifies entity lifecycle management
func TestEntityLifecycle(t *testing.T) {
	world := ecs.NewWorld(nil)

	// Create entity
	eid := world.NewEntity()
	world.AddComponent(eid, &components.Transform{X: 10, Y: 20, W: 32, H: 32})

	// Queue for initialization
	world.QueueInit(eid)
	initQueue := world.DrainInitQueue()
	if len(initQueue) != 1 || initQueue[0] != eid {
		t.Errorf("Init queue should contain entity %d", eid)
	}

	// Drain again should be empty
	initQueue2 := world.DrainInitQueue()
	if len(initQueue2) != 0 {
		t.Error("Init queue should be empty after draining")
	}

	// Register as active
	world.RegisterActive(eid)

	// Queue for removal
	world.QueueRemoval(eid)
	removalQueue := world.DrainRemovalQueue()
	if len(removalQueue) != 1 || removalQueue[0] != eid {
		t.Errorf("Removal queue should contain entity %d", eid)
	}

	// Finalize removal (moves to removed list)
	world.FinalizeRemoval(eid)

	// Actually destroy the entity to remove components
	world.DestroyEntity(eid)

	// Verify entity components are removed
	if ecs.HasComponent[components.Transform](world, eid) {
		t.Error("Component should be removed after destroying entity")
	}
}

// TestComponentPointerSemantics verifies shared state between component references
func TestComponentPointerSemantics(t *testing.T) {
	world := ecs.NewWorld(nil)
	eid := world.NewEntity()

	// Add component
	transform := &components.Transform{X: 10, Y: 20, W: 32, H: 32}
	world.AddComponent(eid, transform)

	// Retrieve component
	retrieved := ecs.GetComponent[components.Transform](world, eid)
	if retrieved == nil {
		t.Fatal("Component should be retrievable")
	}

	// Modify original
	transform.X = 100

	// Verify retrieved reflects change (same pointer)
	if retrieved.X != 100 {
		t.Error("Components should share state (pointer semantics)")
	}

	// Modify retrieved
	retrieved.Y = 200

	// Verify original reflects change
	if transform.Y != 200 {
		t.Error("Components should share state (pointer semantics)")
	}

	// Verify both point to same memory
	if transform != retrieved {
		t.Error("Component pointers should be identical")
	}
}

// TestHasComponents verifies checking for multiple components
func TestHasComponents(t *testing.T) {
	world := ecs.NewWorld(nil)
	eid := world.NewEntity()

	// Initially has no components
	if ecs.HasComponent[components.Transform](world, eid) {
		t.Error("New entity should not have Transform")
	}

	// Add transform
	world.AddComponent(eid, &components.Transform{X: 0, Y: 0, W: 32, H: 32})

	// Should have transform
	if !ecs.HasComponent[components.Transform](world, eid) {
		t.Error("Entity should have Transform after adding")
	}

	// Should not have health
	if ecs.HasComponent[components.Health](world, eid) {
		t.Error("Entity should not have Health")
	}

	// Add health
	world.AddComponent(eid, &components.Health{Current: 100, Max: 100})

	// Should have both
	if !ecs.HasComponent[components.Transform](world, eid) {
		t.Error("Entity should still have Transform")
	}
	if !ecs.HasComponent[components.Health](world, eid) {
		t.Error("Entity should have Health after adding")
	}

	// Test HasComponents with multiple types
	if !ecs.HasComponents(world, eid, (*components.Transform)(nil), (*components.Health)(nil)) {
		t.Error("Entity should have both Transform and Health")
	}

	if ecs.HasComponents(world, eid, (*components.Transform)(nil), (*components.Stamina)(nil)) {
		t.Error("Entity should not have Stamina")
	}
}

// TestMultipleEntitiesWithSameComponents verifies query correctness with multiple entities
func TestMultipleEntitiesWithSameComponents(t *testing.T) {
	world := ecs.NewWorld(nil)

	// Create 10 entities with Transform
	entities := make([]entities.EntityId, 10)
	for i := 0; i < 10; i++ {
		eid := world.NewEntity()
		world.AddComponent(eid, &components.Transform{
			X: float64(i * 10),
			Y: float64(i * 10),
			W: 32,
			H: 32,
		})
		entities[i] = eid
	}

	// Query should return all 10
	results := world.EntitiesWith((*components.Transform)(nil))
	if len(results) != 10 {
		t.Errorf("Expected 10 entities with Transform, got %d", len(results))
	}

	// Add Health to half of them
	for i := 0; i < 5; i++ {
		world.AddComponent(entities[i], &components.Health{Current: 100, Max: 100})
	}

	// Query Transform should still return 10
	results = world.EntitiesWith((*components.Transform)(nil))
	if len(results) != 10 {
		t.Errorf("Expected 10 entities with Transform, got %d", len(results))
	}

	// Query Transform + Health should return 5
	results = world.EntitiesWith((*components.Transform)(nil), (*components.Health)(nil))
	if len(results) != 5 {
		t.Errorf("Expected 5 entities with both components, got %d", len(results))
	}

	// Query Health should return 5
	results = world.EntitiesWith((*components.Health)(nil))
	if len(results) != 5 {
		t.Errorf("Expected 5 entities with Health, got %d", len(results))
	}
}

// TestEntityIDMappings verifies entity ID to uint mapping
func TestEntityIDMappings(t *testing.T) {
	world := ecs.NewWorld(nil)

	eid := world.NewEntity()
	externalID := uint(123)

	// Queue with ID
	world.QueueInitWithID(eid, externalID)

	// Verify mapping
	retrievedEID := world.GetEntityByID(externalID)
	if retrievedEID != eid {
		t.Errorf("GetEntityByID(%d) = %d, want %d", externalID, retrievedEID, eid)
	}

	// Verify reverse mapping
	retrievedID := world.GetID(eid)
	if retrievedID != externalID {
		t.Errorf("GetID(%d) = %d, want %d", eid, retrievedID, externalID)
	}
}

// TestDestroyEntity verifies complete entity destruction
func TestDestroyEntity(t *testing.T) {
	world := ecs.NewWorld(nil)

	eid := world.NewEntity()
	world.AddComponent(eid, &components.Transform{X: 10, Y: 20, W: 32, H: 32})
	world.AddComponent(eid, &components.Health{Current: 100, Max: 100})

	// Verify components exist
	if !ecs.HasComponent[components.Transform](world, eid) {
		t.Fatal("Transform should exist before destruction")
	}
	if !ecs.HasComponent[components.Health](world, eid) {
		t.Fatal("Health should exist before destruction")
	}

	// Destroy entity
	world.DestroyEntity(eid)

	// Verify components removed
	if ecs.HasComponent[components.Transform](world, eid) {
		t.Error("Transform should not exist after destruction")
	}
	if ecs.HasComponent[components.Health](world, eid) {
		t.Error("Health should not exist after destruction")
	}

	// GetComponent should return nil
	if transform := ecs.GetComponent[components.Transform](world, eid); transform != nil {
		t.Error("GetComponent should return nil for destroyed entity")
	}
}
