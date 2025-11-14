// ecs/world.go
package ecs

import (
	"game/entities"
	"game/pkg/bump"
	"game/pkg/config"
	"game/pkg/tilemap"
	"game/resources"
	"log"
	"reflect"
	"sort"
	"sync"
)

// World holds entities, components, and resources.
// This is the single source of truth for the entire game.
//
// Non-ECS concerns (Camera, Map, Space, TimeControl) are stored
// as resources and accessed via Resource[T](world) instead of direct fields.
//
// Resources:
//   - *resources.Camera: Camera system for viewport and room management
//   - *resources.MapRef: Current tilemap reference
//   - *bump.Space: Collision space for physics
//   - *resources.TimeControl: Game speed and freeze effects
type World struct {
	mu         sync.RWMutex
	nextID     entities.EntityId
	components map[entities.EntityId]map[reflect.Type]any
	resources  map[reflect.Type]any

	// Entity lifecycle management
	entities          []entities.EntityId
	idToEntity        map[uint]entities.EntityId
	entityToID        map[entities.EntityId]uint
	toInit            []entities.EntityId
	toRemove          []entities.EntityId
	removed           []entityIDPair
	pendingRemovalIDs map[entities.EntityId]uint

	// typeIndex maps a component *type* to the set of entities that own that
	// component. This is an inverse index used to accelerate EntitiesWith
	// queries by iterating the smallest candidate set instead of scanning all
	// entities each time.
	typeIndex map[reflect.Type]map[entities.EntityId]struct{}
}

type entityIDPair struct {
	entityID entities.EntityId
	id       uint
}

// NewWorld constructs an ECS world with all game systems initialized.
// Core game systems (Camera, Map, Space, TimeControl) are initialized as resources.
// If cfg is nil, uses default config.
func NewWorld(cfg *config.Config) *World {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	world := &World{
		components:        make(map[entities.EntityId]map[reflect.Type]any),
		resources:         make(map[reflect.Type]any),
		typeIndex:         make(map[reflect.Type]map[entities.EntityId]struct{}),
		idToEntity:        make(map[uint]entities.EntityId),
		entityToID:        make(map[entities.EntityId]uint),
		pendingRemovalIDs: make(map[entities.EntityId]uint),
	}

	// Initialize core resources
	world.SetResource(cfg)
	world.SetResource(bump.NewSpace())
	world.SetResource(resources.NewCamera(cfg.Screen.Width, cfg.Screen.Height, cfg))
	world.SetResource(&resources.MapRef{}) // Empty map reference, will be set by BindMap
	world.SetResource(resources.NewTimeControl())

	// Store config as a resource so systems can access it via Resource[config.Config](world)
	world.SetResource(cfg)

	return world
}

// NewEntity hands out a fresh ID
func (w *World) NewEntity() entities.EntityId {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.nextID++
	id := w.nextID
	w.components[id] = make(map[reflect.Type]any)
	return id
}

// DestroyEntity erases the entity and all of its components from the world.
// Any future calls that reference this ID will see no components.
func (w *World) DestroyEntity(e entities.EntityId) {
	w.mu.Lock()
	defer w.mu.Unlock()
	// remove it from the queue
	// zq := Resource[resources.ZIndex](w)
	// idx := slices.Index(zq.Entities, e)
	// if idx >= 0 {
	// 	zq.Entities = append(zq.Entities[:idx], zq.Entities[idx+1:]...)
	// }
	// Remove entity from typeIndex for each owned component, then remove
	// the component map entirely.
	if comps := w.components[e]; comps != nil {
		for typ := range comps {
			if idxSet, ok := w.typeIndex[typ]; ok {
				delete(idxSet, e)
				if len(idxSet) == 0 {
					delete(w.typeIndex, typ)
				}
			}
		}
	}
	delete(w.components, e)
	// Note: we deliberately leave `nextID` unchanged so IDs are never reused
	// within the same session. If you need ID recycling in the future,
	// implement a free‑list here.
}

func (w *World) SetResource(res any) {
	if reflect.TypeOf(res).Kind() != reflect.Ptr {
		panic("SetResource expects a pointer, got value")
	}
	if w.resources == nil {
		w.resources = make(map[reflect.Type]any)
	}
	w.resources[reflect.TypeOf(res)] = res // res must be a *pointer*
}

func Resource[T any](w *World) *T {
	if w.resources == nil {
		return nil
	}
	if r, ok := w.resources[reflect.TypeOf((*T)(nil))]; ok {
		return r.(*T)
	}
	return nil
}

// AddComponent attaches a *pointer* to a component struct.
// Always pass `&MyComponent{...}`, never a value, so the same
// memory is shared across systems and no copies are made.
func (w *World) AddComponent(e entities.EntityId, comp any) { // expect *T here
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.components[e] == nil {
		w.components[e] = make(map[reflect.Type]any)
	}
	typ := reflect.TypeOf(comp)
	w.components[e][typ] = comp // comp is already a pointer

	// Maintain inverse index for quick queries
	if w.typeIndex == nil {
		w.typeIndex = make(map[reflect.Type]map[entities.EntityId]struct{})
	}
	set := w.typeIndex[typ]
	if set == nil {
		set = make(map[entities.EntityId]struct{})
		w.typeIndex[typ] = set
	}
	set[e] = struct{}{}
}

// EntitiesWith returns all entities that have _all_ of the given component types.
// Pass a nil *pointer* of each type you care about, e.g.:
//
//	world.EntitiesWith((*Position)(nil), (*Velocity)(nil))
//
// This keeps all component handling strictly pointer‑based.
func (w *World) EntitiesWith(components ...any) []entities.EntityId {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// If no components specified, return all entities that have any component.
	if len(components) == 0 {
		res := make([]entities.EntityId, 0, len(w.components))
		for eid := range w.components {
			res = append(res, eid)
		}
		return res
	}

	// Convert to reflect.Type keys and find the smallest candidate set via typeIndex.
	types := make([]reflect.Type, 0, len(components))
	for _, c := range components {
		types = append(types, reflect.TypeOf(c))
	}

	// Find the smallest indexed set among the requested types.
	var smallestSet map[entities.EntityId]struct{}
	for _, t := range types {
		set := w.typeIndex[t]
		if set == nil {
			// No entity has this component -> empty result.
			return nil
		}
		if smallestSet == nil || len(set) < len(smallestSet) {
			smallestSet = set
		}
	}

	// For each candidate in the smallest set, verify it has all requested types.
	result := make([]entities.EntityId, 0, len(smallestSet))
	for eid := range smallestSet {
		compMap := w.components[eid]
		if compMap == nil {
			continue
		}
		ok := true
		for _, t := range types {
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

// GetComponent returns a *pointer* to the component of type T for the
// given entity, or nil if the entity does not possess that component.
// Use it like:
//
//	pos := ecs.GetComponent[Position](world, eid)
//	if pos != nil { ... }
//
// Note: The generic parameter T is the **value** type (e.g. Position),
// and the function returns *T so that you always work on the shared
// instance stored in the world.
func GetComponent[T any](w *World, e entities.EntityId) *T {
	w.mu.RLock()
	defer w.mu.RUnlock()

	comps := w.components[e]
	if comps == nil {
		return nil
	}

	// We stored *T, so look it up by that exact reflect.Type.
	typ := reflect.TypeOf((*T)(nil)) // *T
	if raw, ok := comps[typ]; ok {
		return raw.(*T) // guaranteed safe
	}
	return nil
}

// HasComponent returns true if the entity owns a component of type T.
func HasComponent[T any](w *World, e entities.EntityId) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	comps := w.components[e]
	if comps == nil {
		return false
	}
	typ := reflect.TypeOf((*T)(nil)) // *T
	_, ok := comps[typ]
	return ok
}

// HasComponents returns true iff the entity owns *all* of the component
// types passed in `comps`.  Pass nil *pointers* of the component structs you
// want to test for—exactly the same way you call EntitiesWith:
//
//	if ecs.HasComponents(world, eid, (*Position)(nil), (*Velocity)(nil)) {
//		...
//	}
//
// An empty list of comps returns true (vacuously).
func HasComponents(w *World, e entities.EntityId, comps ...any) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	compMap := w.components[e]
	if compMap == nil {
		return false
	}
	for _, c := range comps {
		if c == nil {
			continue
		}
		t := reflect.TypeOf(c) // *T
		if _, ok := compMap[t]; !ok {
			return false
		}
	}
	return true
}

// RemoveComponent removes the component of type T from the entity (no‑op if missing).
func RemoveComponent[T any](w *World, e entities.EntityId) {
	w.mu.Lock()
	defer w.mu.Unlock()
	// remove from components
	comps := w.components[e]
	if comps == nil {
		return
	}
	typ := reflect.TypeOf((*T)(nil)) // *T
	delete(comps, typ)

	// maintain typeIndex
	if w.typeIndex != nil {
		if set, ok := w.typeIndex[typ]; ok {
			delete(set, e)
			if len(set) == 0 {
				delete(w.typeIndex, typ)
			}
		}
	}
}

// ========================================
// World State Management Methods
// ========================================

// BindMap attaches a map to the world, wiring camera rooms and collision objects.
func (w *World) BindMap(worldMap *tilemap.Map, roomsLayer, collisionsLayer string) {
	if w == nil || worldMap == nil {
		return
	}

	// Update map resource
	mapRef := Resource[resources.MapRef](w)
	if mapRef == nil {
		mapRef = &resources.MapRef{}
		w.SetResource(mapRef)
	}
	mapRef.Map = worldMap

	// Create fresh collision space
	w.SetResource(bump.NewSpace())
	space := Resource[bump.Space](w)

	// Get or create camera
	camera := Resource[resources.Camera](w)
	if camera == nil {
		cfg := Resource[config.Config](w)
		if cfg == nil {
			cfg = config.NewDefaultConfig()
		}
		camera = resources.NewCamera(cfg.Screen.Width, cfg.Screen.Height, cfg)
		w.SetResource(camera)
	}

	// Wire camera rooms
	if roomsLayer != "" && camera != nil {
		rooms := tilemap.GetObjectsRects(worldMap, roomsLayer)
		if rooms != nil {
			camera.SetRooms(rooms)
		}
	}

	// Populate collision geometry
	if space != nil {
		tilemap.LoadTilesetCollisionObjects(worldMap, space)
		if collisionsLayer != "" {
			tilemap.LoadBumpObjects(worldMap, space, collisionsLayer)
		}
	}

	// Load ladder tiles into registry AND collision space
	ladderRegistry := Resource[resources.LadderRegistry](w)
	if ladderRegistry == nil {
		ladderRegistry = resources.NewLadderRegistry()
		w.SetResource(ladderRegistry)
	}
	err := tilemap.LoadLadderTiles(worldMap, ladderRegistry, space)
	if err != nil {
		log.Printf("Error loading ladder tiles: %v", err)
	}
}

// Follow makes the camera follow the given target.
func (w *World) Follow(target resources.Recter) {
	camera := Resource[resources.Camera](w)
	if target == nil || camera == nil {
		return
	}
	camera.Follow(target)
}

// ShakeCamera applies a shake effect to the camera.
func (w *World) ShakeCamera(duration float32, magnitude float64) {
	if camera := Resource[resources.Camera](w); camera != nil {
		camera.Shake(duration, magnitude)
	}
}

// UpdateMap updates animated map tiles.
func (w *World) UpdateMap(dt float64) {
	mapRef := Resource[resources.MapRef](w)
	if mapRef == nil || mapRef.Map == nil {
		return
	}
	tilemap.Update(mapRef.Map, dt)
}

// CameraInFrame checks if a Recter is visible within the camera frame.
func (w *World) CameraInFrame(r resources.Recter, marginX, marginY float64) bool {
	camera := Resource[resources.Camera](w)
	if camera == nil || r == nil {
		return false
	}
	return camera.InFrame(r, marginX, marginY)
}

// CameraInFrameRecter is an alias for CameraInFrame for compatibility with draw systems
func (w *World) CameraInFrameRecter(r resources.Recter, marginX, marginY float64) bool {
	return w.CameraInFrame(r, marginX, marginY)
}

// CameraPosition returns the current camera position.
func (w *World) CameraPosition() (float64, float64) {
	camera := Resource[resources.Camera](w)
	if camera == nil {
		return 0, 0
	}
	return camera.Position()
}

// ========================================
// Entity Lifecycle Management
// ========================================

// QueueInit adds an entity to the initialization queue.
func (w *World) QueueInit(entityID entities.EntityId) {
	if entityID == 0 {
		return
	}
	w.mu.Lock()
	w.toInit = append(w.toInit, entityID)
	w.mu.Unlock()
}

// QueueInitWithID adds an entity with a specific ID mapping.
func (w *World) QueueInitWithID(entityID entities.EntityId, id uint) {
	if entityID == 0 {
		return
	}
	w.mu.Lock()
	w.idToEntity[id] = entityID
	w.entityToID[entityID] = id
	w.toInit = append(w.toInit, entityID)
	w.mu.Unlock()
}

// ScheduleAdd enqueues an entity, wiring the provided ID when non-zero.
func (w *World) ScheduleAdd(entityID entities.EntityId, id uint) {
	if id != 0 {
		w.QueueInitWithID(entityID, id)
		return
	}
	w.QueueInit(entityID)
}

// DrainInitQueue returns and clears the initialization queue.
func (w *World) DrainInitQueue() []entities.EntityId {
	w.mu.Lock()
	defer w.mu.Unlock()
	pending := w.toInit
	w.toInit = nil
	return pending
}

// GetID returns the uint ID for an entity (if mapped).
func (w *World) GetID(e entities.EntityId) uint {
	return w.entityToID[e]
}

// GetEntityByID returns the entity for a given uint ID.
func (w *World) GetEntityByID(id uint) entities.EntityId {
	return w.idToEntity[id]
}

// ActiveEntityCount returns the number of active entities.
func (w *World) ActiveEntityCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.entities)
}

// ActiveIDs returns a sorted list of active entity IDs.
func (w *World) ActiveIDs() []uint {
	w.mu.Lock()
	ids := make([]uint, 0, len(w.entityToID))
	for _, id := range w.entityToID {
		if id != 0 {
			ids = append(ids, id)
		}
	}
	w.mu.Unlock()
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Clear removes all tracked entities and queues.
func (w *World) Clear() {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entities = nil
	w.toInit = nil
	w.toRemove = nil
	w.removed = nil
	w.idToEntity = make(map[uint]entities.EntityId)
	w.entityToID = make(map[entities.EntityId]uint)
	w.pendingRemovalIDs = make(map[entities.EntityId]uint)

	// Reset time control resource
	if tc := Resource[resources.TimeControl](w); tc != nil {
		tc.Reset()
	}
}

// QueueRemoval adds an entity to the removal queue.
func (w *World) QueueRemoval(entityID entities.EntityId) {
	if entityID == 0 {
		return
	}
	w.mu.Lock()
	w.toRemove = append(w.toRemove, entityID)
	id := w.entityToID[entityID]
	delete(w.entityToID, entityID)
	for id, stored := range w.idToEntity {
		if stored == entityID {
			delete(w.idToEntity, id)
			break
		}
	}
	if w.pendingRemovalIDs == nil {
		w.pendingRemovalIDs = make(map[entities.EntityId]uint)
	}
	w.pendingRemovalIDs[entityID] = id
	w.mu.Unlock()
}

// DrainRemovalQueue returns and clears the removal queue.
func (w *World) DrainRemovalQueue() []entities.EntityId {
	w.mu.Lock()
	removals := append([]entities.EntityId(nil), w.toRemove...)
	w.toRemove = nil
	w.mu.Unlock()
	return removals
}

// RemovedEntities returns the list of removed entities.
func (w *World) RemovedEntities() []entities.EntityId {
	w.mu.Lock()
	defer w.mu.Unlock()
	removed := make([]entities.EntityId, 0, len(w.removed))
	for _, entry := range w.removed {
		removed = append(removed, entry.entityID)
	}
	return removed
}

// RangeEntities iterates over all active entities.
func (w *World) RangeEntities(fn func(entities.EntityId, uint)) {
	if fn == nil {
		return
	}
	w.mu.Lock()
	pairs := make([]entityIDPair, 0, len(w.entityToID))
	for entityID, id := range w.entityToID {
		pairs = append(pairs, entityIDPair{entityID: entityID, id: id})
	}
	w.mu.Unlock()
	for _, pair := range pairs {
		fn(pair.entityID, pair.id)
	}
}

// RangeRemoved iterates over all removed entities.
func (w *World) RangeRemoved(fn func(entities.EntityId, uint)) {
	if fn == nil {
		return
	}
	w.mu.Lock()
	removed := append([]entityIDPair(nil), w.removed...)
	w.mu.Unlock()
	for _, entry := range removed {
		fn(entry.entityID, entry.id)
	}
}

// RegisterActive adds an entity to the active list.
func (w *World) RegisterActive(entityID entities.EntityId) bool {
	if entityID == 0 {
		return false
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, existing := range w.entities {
		if existing == entityID {
			return false
		}
	}
	w.entities = append(w.entities, entityID)
	return true
}

// ActiveEntities returns a copy of the active entity list.
func (w *World) ActiveEntities() []entities.EntityId {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]entities.EntityId(nil), w.entities...)
}

// FinalizeRemoval completes the removal of an entity.
func (w *World) FinalizeRemoval(target entities.EntityId) {
	if target == 0 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	id := w.pendingRemovalIDs[target]
	delete(w.pendingRemovalIDs, target)
	for i, entityID := range w.entities {
		if entityID != target {
			continue
		}
		w.entities[i] = w.entities[len(w.entities)-1]
		w.entities = w.entities[:len(w.entities)-1]
		w.removed = append(w.removed, entityIDPair{entityID: entityID, id: id})
		return
	}
}
