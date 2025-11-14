package entities

import (
	"image/color"
	"math/rand/v2"
	"time"

	"game/components"
	"game/components/debug"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/prefabs"
	"game/resources"
	"game/systems/update/physics"
)

const (
	fakeWallNeighborDelay = 0.3  // Delay before triggering neighbors (seconds)
	fakeWallDestroyDelay  = 0.6  // Delay before destroying entity (seconds)
)

// ApplyFakeWallOpen processes combat events to crumble fake walls when hit by the player.
// Simplified Pattern C implementation.
func ApplyFakeWallOpen(world *ecs.World, events []resources.HitEvent) {
	if world == nil || len(events) == 0 {
		return
	}

	for i := range events {
		event := &events[i]
		if event.Handled {
			continue
		}

		// Only process hits (contact type 0 = Hit)
		if event.Contact != int(components.Hit) {
			continue
		}

		// Check if target is a fake wall
		fakeWall := ecs.GetComponent[components.FakeWall](world, event.Target)
		if fakeWall == nil {
			continue
		}

		// Skip if fake wall is already open
		if fakeWall.Opened {
			continue
		}

		// Only player can crumble fake walls
		attackerEntity := event.Attacker
		if attackerEntity == 0 {
			continue
		}
		team := ecs.GetComponent[components.Team](world, attackerEntity)
		if team == nil || team.Type != components.TeamPlayer {
			continue
		}

		// Crumble the fake wall with chain reaction
		CrumbleFakeWall(world, event.Target)
		fakeWall.Opened = true
		event.Handled = true
	}
}

// CrumbleFakeWall handles the crumbling animation and chain reaction for a fake wall.
// This includes: spawning smoke particles, camera shake, and scheduling neighbor triggering.
func CrumbleFakeWall(world *ecs.World, wallID entities.EntityId) {
	if world == nil {
		return
	}

	fakeWall := ecs.GetComponent[components.FakeWall](world, wallID)
	if fakeWall == nil || fakeWall.Opened {
		return
	}

	// Mark as opened
	fakeWall.Opened = true

	// Get transform for position (needed for neighbor triggering)
	transform := ecs.GetComponent[components.Transform](world, wallID)
	if transform == nil {
		return
	}

	// IMMEDIATELY remove collision so player can pass through
	// (similar to door behavior - don't wait for delayed removal)
	if space := physics.GetCollisionSpace(world); space != nil {
		// Remove the fake wall entity from collision space
		space.Remove(wallID)

		// CRITICAL FIX: Also remove underlying map collision tiles at this position
		// These were registered by LoadBumpObjects with *tiled.Object items during world init.
		// Without this, the map collision tiles remain as invisible barriers.
		queryRect := bump.Rect{X: transform.X, Y: transform.Y, W: transform.W, H: transform.H}

		// Try querying with different tag combinations to catch all collision types
		tagCombinations := [][]string{
			{"map"},   // Standard map tiles
			{"solid"}, // Solid collision tiles
			{},        // No tags (all items in rect)
		}

		for _, tags := range tagCombinations {
			var collisions []*bump.Collision
			if len(tags) == 0 {
				collisions = space.Query(queryRect, nil)
			} else {
				collisions = space.Query(queryRect, nil, bump.Tag(tags[0]))
			}

			// Remove each collision tile item
			for _, col := range collisions {
				if col.Other != nil && col.Other != wallID {
					space.Remove(col.Other)
				}
			}
		}

		// Visual debug (Shift+F to toggle): Show collision state after removal
		if debugState := ecs.Resource[resources.DebugState](world); debugState != nil && debugState.IsEnabled(resources.DebugCategoryFakeWall) {
			// Verify what remains after cleanup
			verifyCollisions := space.Query(queryRect, nil)
			if len(verifyCollisions) > 0 {
				// RED filled rectangles: Remaining collision items (problem!)
				for _, col := range verifyCollisions {
					if col.Other != nil {
						otherRect := space.Rect(col.Other)
						createDebugRect(world, otherRect.X, otherRect.Y, otherRect.W, otherRect.H,
							color.RGBA{R: 255, G: 0, B: 0, A: 180}, true, 5*time.Second)
					}
				}
			} else {
				// GREEN outline: Successfully cleared area
				createDebugRect(world, transform.X, transform.Y, transform.W, transform.H,
					color.RGBA{R: 0, G: 255, B: 0, A: 100}, false, 3*time.Second)
			}
		}
	}

	// Remove the Collider component entirely
	ecs.RemoveComponent[components.Collider](world, wallID)

	// Clear hitboxes (wall can no longer be hit again)
	hitbox := ecs.GetComponent[components.Hitbox](world, wallID)
	if hitbox != nil {
		hitbox.Boxes = hitbox.Boxes[:0]
	}

	// Immediate effects: camera shake and smoke particles
	if camera := ecs.Resource[resources.Camera](world); camera != nil {
		camera.Shake(0.1, 0.1)
	}
	spawnFakeWallSmoke(world, wallID, 5+rand.IntN(5))

	// Schedule neighbor trigger and destruction using deterministic timers
	fakeWall.NeighborTriggerTime = fakeWallNeighborDelay
	fakeWall.DestroyTimer = fakeWallDestroyDelay
}

// UpdateFakeWallTimers processes fake wall timers for neighbor triggering and destruction.
// This system should be called after ApplyFakeWallOpen in the update cycle.
func UpdateFakeWallTimers(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	// Process all fake walls with active timers
	for _, eid := range world.EntitiesWith((*components.FakeWall)(nil)) {
		fakeWall := ecs.GetComponent[components.FakeWall](world, eid)
		if fakeWall == nil {
			continue
		}

		// Process neighbor trigger timer
		if fakeWall.NeighborTriggerTime > 0 {
			fakeWall.NeighborTriggerTime -= dt
			if fakeWall.NeighborTriggerTime <= 0 {
				fakeWall.NeighborTriggerTime = 0
				// Trigger neighbor chain reaction
				transform := ecs.GetComponent[components.Transform](world, eid)
				if transform != nil {
					triggerNeighborFakeWalls(world, eid, transform.X, transform.Y)
				}
			}
		}

		// Process destruction timer
		if fakeWall.DestroyTimer > 0 {
			fakeWall.DestroyTimer -= dt
			if fakeWall.DestroyTimer <= 0 {
				fakeWall.DestroyTimer = 0
				world.DestroyEntity(eid)
			}
		}
	}
}

// spawnFakeWallSmoke spawns smoke particles at the fake wall location
func spawnFakeWallSmoke(world *ecs.World, wallID entities.EntityId, count int) {
	for range count {
		smokeID := prefabs.NewSmokeFrom(world, wallID)
		if smokeID != 0 {
			world.QueueInit(smokeID)
		}
	}
}

// triggerNeighborFakeWalls finds and triggers adjacent fake walls in a chain reaction
func triggerNeighborFakeWalls(world *ecs.World, wallID entities.EntityId, x, y float64) {
	if world == nil {
		return
	}

	// Note: Since collision is removed immediately when a wall crumbles,
	// we can't use the Collider component to check. Instead, just check
	// if the wall entity still exists (hasn't been destroyed yet).
	fakeWall := ecs.GetComponent[components.FakeWall](world, wallID)
	if fakeWall == nil {
		return // Wall was already destroyed
	}

	space := physics.GetCollisionSpace(world)
	if space == nil {
		return
	}

	const TileSize = prefabs.TileSize

	// Search horizontally: left and right neighbors
	// Expand by 0.1 to ensure we catch tiles that are exactly adjacent
	searchRect := bump.Rect{X: x - TileSize - 0.1, Y: y, W: TileSize*3 + 0.2, H: TileSize}
	horizontal := physics.QueryItems(space, wallID, searchRect, "fakeWall")

	// Search vertically: above and below neighbors
	searchRect = bump.Rect{X: x, Y: y - TileSize - 0.1, W: TileSize, H: TileSize*3 + 0.2}
	vertical := physics.QueryItems(space, wallID, searchRect, "fakeWall")

	neighbors := append(horizontal, vertical...)

	// Recursively trigger neighbor fake walls
	for _, neighborID := range neighbors {
		CrumbleFakeWall(world, neighborID)
	}
}

// createDebugRect creates a temporary debug rectangle visualization
func createDebugRect(world *ecs.World, x, y, w, h float64, col color.RGBA, fill bool, duration time.Duration) {
	debugEnt := world.NewEntity()
	world.AddComponent(debugEnt, &components.Transform{X: x, Y: y, W: w, H: h})
	world.AddComponent(debugEnt, &debug.DebugVisual{
		Type:  debug.DebugVisualRect,
		Color: col,
		X:     x,
		Y:     y,
		W:     w,
		H:     h,
		Fill:  fill,
	})
	world.QueueInit(debugEnt)

	// Schedule destruction using component timer
	fakeWall := &components.FakeWall{
		DestroyTimer: duration.Seconds(),
	}
	world.AddComponent(debugEnt, fakeWall)
}
