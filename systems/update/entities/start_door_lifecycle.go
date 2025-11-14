package entities

import (
	"time"

	"game/components"
	"game/ecs"
	"game/prefabs"
	"game/resources"
)

const (
	// Effect properties from prefabs/start_door.go
	startDoorShakeDuration  = 0.1 // Camera shake duration in seconds
	startDoorShakeMagnitude = 0.1 // Camera shake magnitude
	startDoorSmokeCount     = 10  // Number of smoke particles to spawn
)

// UpdateStartDoorLifecycle handles the timed destruction of StartDoor entities.
// This is a Pure ECS replacement for the time.AfterFunc in prefabs/start_door.go.
//
// After the configured delay has elapsed:
//  1. Triggers camera shake for dramatic effect
//  2. Spawns smoke particles around the door
//  3. Queues the entity for removal
func UpdateStartDoorLifecycle(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	camera := ecs.Resource[resources.Camera](world)

	for _, eid := range world.EntitiesWith((*components.StartDoor)(nil)) {
		door := ecs.GetComponent[components.StartDoor](world, eid)
		if door == nil {
			continue
		}

		// Check if delay has elapsed
		elapsed := time.Since(door.SpawnTime).Seconds()
		if elapsed < door.DestroyDelay {
			continue
		}

		// Camera shake for dramatic effect
		if camera != nil {
			camera.Shake(float32(startDoorShakeDuration), startDoorShakeMagnitude)
		}

		// Spawn smoke particles around the door
		for i := 0; i < startDoorSmokeCount; i++ {
			smokeID := prefabs.NewSmokeFrom(world, eid)
			if smokeID != 0 {
				world.QueueInit(smokeID)
			}
		}

		// Remove the door entity
		world.DestroyEntity(eid)
	}
}
