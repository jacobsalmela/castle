package initialization

import (
	"game/components"
	"game/ecs"
	"game/resources"
)

type cameraFollowState interface {
	Follow(resources.Recter)
}

// RunCameraFollow ensures the camera is following the player via its Transform component.
// This is idempotent and safe to call every frame.
func RunCameraFollow(world *ecs.World, state cameraFollowState) {
	if world == nil || state == nil {
		return
	}
	// Find the player in ECS and follow its Transform component.
	for _, eid := range world.EntitiesWith((*components.Player)(nil)) {
		p := ecs.GetComponent[components.Player](world, eid)
		if p == nil {
			continue
		}
		// Use ECS-native Transform for camera following
		if t := ecs.GetComponent[components.Transform](world, eid); t != nil {
			state.Follow(t)
			return
		}
		// If no Transform found, this is an error - player should always have Transform
		return
	}
}
