package entities

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/resources"
	"game/systems/update/physics"
)

// ApplyDoorOpen processes combat events to open doors when hit by the player
// from the correct direction. Simplified Pattern C implementation.
func ApplyDoorOpen(world *ecs.World, events []resources.HitEvent) {
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

		// Check if target is a door
		door := ecs.GetComponent[components.Door](world, event.Target)
		if door == nil {
			continue
		}

		// Skip if door is already open
		if door.Opened {
			continue
		}

		// Only player can open doors
		attackerEntity := event.Attacker
		if attackerEntity == 0 {
			continue
		}
		team := ecs.GetComponent[components.Team](world, attackerEntity)
		if team == nil || team.Type != components.TeamPlayer {
			continue
		}

		// Get door transform for position check
		doorTransform := ecs.GetComponent[components.Transform](world, event.Target)
		if doorTransform == nil {
			continue
		}

		// Get attacker position from transform
		attackerTransform := ecs.GetComponent[components.Transform](world, attackerEntity)
		if attackerTransform == nil {
			continue
		}
		attackerX := attackerTransform.X
		doorX := doorTransform.X

		correctDirection := false
		if door.OpensFromRight && attackerX > doorX {
			correctDirection = true
		} else if !door.OpensFromRight && attackerX < doorX {
			correctDirection = true
		}

		if !correctDirection {
			continue
		}

		// Open the door using system function (handles door chain)
		OpenDoor(world, event.Target)

		event.Handled = true

		// Camera shake for door opening
		if camera := ecs.Resource[resources.Camera](world); camera != nil {
			camera.Shake(0.3, 0.8)
		}
	}
}

// OpenDoor opens a door entity, making it non-solid and updating its appearance.
func OpenDoor(world *ecs.World, doorID entities.EntityId) {
	if world == nil {
		return
	}

	// Get door component
	door := ecs.GetComponent[components.Door](world, doorID)
	if door == nil || door.Opened {
		return // Already open or not a door
	}

	// Mark as opened
	door.Opened = true

	// Remove collision entirely - player should be able to walk through
	if space := physics.GetCollisionSpace(world); space != nil {
		space.Remove(doorID)
	}

	// Remove the Collider component entirely (not just disable it)
	ecs.RemoveComponent[components.Collider](world, doorID)

	// Clear hitboxes (door can no longer be hit)
	hitbox := ecs.GetComponent[components.Hitbox](world, doorID)
	if hitbox != nil {
		hitbox.Boxes = hitbox.Boxes[:0]
	}

	// Trigger the "activate" animation to transition to open state
	anim := ecs.GetComponent[components.Animation](world, doorID)
	if anim != nil {
		if anim.Data != nil {
			if err := anim.Data.Play("activate"); err == nil {
				anim.State = "activate"
				anim.Frame = 0
				anim.Timer = 0
				// Animation system will automatically transition to "open" state when "activate" finishes
				// (configured via FSMTransitions: "activate" -> "open")
			}
		}
	}

	// Door chain: Open subsequent doors in front (Pure ECS implementation)
	OpenDoorChain(world, doorID)
}

// InitializeDoors processes doors that need initialization (e.g., pre-opened doors).
// This is called each frame to apply the correct state to doors with NeedsInit flag.
func InitializeDoors(world *ecs.World) {
	if world == nil {
		return
	}

	for _, eid := range world.EntitiesWith((*components.Door)(nil)) {
		door := ecs.GetComponent[components.Door](world, eid)
		if door == nil || !door.NeedsInit {
			continue
		}

		// Clear the init flag
		door.NeedsInit = false

		// Apply open state if door should start open
		if door.Opened {
			applyOpenDoorState(world, eid)
		}
	}
}

// applyOpenDoorState applies the visual and collision state for an open door.
// This is used for pre-opened doors (skips activation animation).
func applyOpenDoorState(world *ecs.World, doorID entities.EntityId) {
	// Set animation to open state (skip activation animation for pre-opened)
	anim := ecs.GetComponent[components.Animation](world, doorID)
	if anim != nil && anim.Data != nil {
		if err := anim.Data.Play("open"); err == nil {
			anim.State = "open"
		}
	}

	// Remove collision entirely
	if space := physics.GetCollisionSpace(world); space != nil {
		space.Remove(doorID)
	}
	ecs.RemoveComponent[components.Collider](world, doorID)

	// Clear hitboxes
	hitbox := ecs.GetComponent[components.Hitbox](world, doorID)
	if hitbox != nil {
		hitbox.Boxes = hitbox.Boxes[:0]
	}
}

// OpenDoorChain opens all doors in front of the given door (door chain behavior).
// Pure ECS implementation - replaces spatial.QueryFront.
func OpenDoorChain(world *ecs.World, doorID entities.EntityId) {
	if world == nil {
		return
	}

	// Get the door that was just opened
	door := ecs.GetComponent[components.Door](world, doorID)
	doorTransform := ecs.GetComponent[components.Transform](world, doorID)
	if door == nil || doorTransform == nil {
		return
	}

	// Define search area in front of the door
	const searchDist = 8.0 // TileSize
	searchHeight := doorTransform.H / 2

	// Calculate search rectangle based on door orientation
	var searchX, searchY, searchW, searchH float64
	if door.OpensFromRight {
		// Search to the right
		searchX = doorTransform.X + doorTransform.W
		searchY = doorTransform.Y - searchHeight
		searchW = searchDist
		searchH = searchHeight * 2
	} else {
		// Search to the left
		searchX = doorTransform.X - searchDist
		searchY = doorTransform.Y - searchHeight
		searchW = searchDist
		searchH = searchHeight * 2
	}

	// Create search rectangle
	searchRect := bump.Rect{X: searchX, Y: searchY, W: searchW, H: searchH}

	// Query all doors in the search area
	for _, otherID := range world.EntitiesWith((*components.Door)(nil)) {
		if otherID == doorID {
			continue // Skip self
		}

		otherDoor := ecs.GetComponent[components.Door](world, otherID)
		otherTransform := ecs.GetComponent[components.Transform](world, otherID)

		if otherDoor == nil || otherTransform == nil || otherDoor.Opened {
			continue // Skip non-doors, missing transforms, or already-open doors
		}

		// Check if other door overlaps with search area
		doorRect := bump.Rect{
			X: otherTransform.X,
			Y: otherTransform.Y,
			W: otherTransform.W,
			H: otherTransform.H,
		}

		if bump.Overlaps(searchRect, doorRect) {
			// Open this door recursively (will trigger its own chain)
			OpenDoor(world, otherID)
		}
	}
}
