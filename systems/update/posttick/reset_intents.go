package posttick

import (
	"game/components"
	"game/ecs"
)

// ResetActionIntents clears frame-specific action intent flags at the end of each frame.
// This ensures intents like Heal, ShieldRelease, ClimbDrop, and ClimbRelease don't persist
// across frames when they should be single-frame actions.
//
// This runs in the posttick phase after all game logic has consumed the intents.
func ResetActionIntents(world *ecs.World) {
	if world == nil {
		return
	}

	// Find all entities with ActionIntents component
	for _, eid := range world.EntitiesWith((*components.ActionIntents)(nil)) {
		intents := ecs.GetComponent[components.ActionIntents](world, eid)
		if intents == nil {
			continue
		}

		// Clear frame-specific intent flags
		// These are "momentary" actions that should only last one frame
		intents.Heal = false
		intents.ShieldRelease = false
		intents.ClimbDrop = false
		intents.ClimbRelease = false

		// NOTE: ShieldHeld and ClimbHeld are NOT cleared here - they persist
		// across frames while the button is held down, and are managed by
		// the input processing logic in decision/player.go
	}
}
