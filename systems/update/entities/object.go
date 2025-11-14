package entities

import (
	"math"
	"math/rand/v2"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/prefabs"
	"game/resources"
	physicsHelper "game/systems/update/physics"
)

// UpdateObject handles object-specific behavior.
// Processes hit events to destroy objects and handles velocity propagation to stacked objects.
func UpdateObject(world *ecs.World, events []resources.HitEvent, _ float64) {
	if world == nil {
		return
	}

	// Process hit events for object destruction
	for i := range events {
		event := &events[i]
		if event.Handled {
			continue
		}

		// Only process hits (contact type 0 = Hit)
		if event.Contact != int(components.Hit) {
			continue
		}

		// Check if target is an object
		object := ecs.GetComponent[components.Object](world, event.Target)
		if object == nil {
			continue
		}

		// Destroy the object
		HandleObjectDestruction(world, event.Target)

		// Mark event as handled to prevent generic combat from processing
		event.Handled = true
	}

	space := ecs.Resource[bump.Space](world)
	if space == nil {
		return
	}

	// Iterate all objects
	for _, eid := range world.EntitiesWith((*components.Object)(nil)) {
		object := ecs.GetComponent[components.Object](world, eid)
		transform := ecs.GetComponent[components.Transform](world, eid)
		physics := ecs.GetComponent[components.Physics](world, eid)
		if object == nil || transform == nil || physics == nil {
			continue
		}

		// Check for objects stacked above this one
		// Query 1 pixel above this object
		queryRect := bump.Rect{
			X: transform.X,
			Y: transform.Y - 1,
			W: transform.W,
			H: 1,
		}

		// Query for other objects
		items := physicsHelper.QueryItems[entities.EntityId](space, eid, queryRect, "object")
		for _, otherEid := range items {

			otherTransform := ecs.GetComponent[components.Transform](world, otherEid)
			otherPhysics := ecs.GetComponent[components.Physics](world, otherEid)
			if otherTransform == nil || otherPhysics == nil {
				continue
			}

			// If the other object is directly above and this object is moving faster,
			// propagate velocity
			if otherTransform.Y+otherTransform.H <= transform.Y &&
				math.Abs(physics.Velocity.X) > math.Abs(otherPhysics.Velocity.X) {
				otherPhysics.Velocity.X = physics.Velocity.X
			}
		}
	}
}

// HandleObjectDestruction processes object destruction when hit.
// This is called by combat systems when an object takes damage.
func HandleObjectDestruction(world *ecs.World, eid entities.EntityId) {
	if world == nil || eid == 0 {
		return
	}

	object := ecs.GetComponent[components.Object](world, eid)
	transform := ecs.GetComponent[components.Transform](world, eid)
	if object == nil || transform == nil {
		return
	}

	// Spawn smoke and debris particles
	numParticles := 5 + rand.IntN(5)
	for range numParticles {
		prefabs.NewSmokeFrom(world, eid)
		prefabs.NewDebrisPrefab(world, eid)
	}

	// Spawn reward flakes
	for range object.Reward {
		// Get player entity as target
		players := world.EntitiesWith((*components.Player)(nil))
		var targetID entities.EntityId
		if len(players) > 0 {
			targetID = players[0]
		}

		// Spawn flake at object center
		x := transform.X + transform.W/2
		y := transform.Y + transform.H/2
		prefabs.NewFlakePrefab(world, x, y, eid, targetID)
	}

	// Destroy the object entity
	world.DestroyEntity(eid)
}
