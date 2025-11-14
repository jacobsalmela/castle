package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/resources"
	"game/systems/update/physics"
)

// UpdateTargetDetection finds targets for entities with DetectionRange component.
// Sets AI.TargetID when valid target found within detection rect.
// Only searches when AI.TargetID is 0 (no current target) - implements "sticky targeting".
//
// Once an enemy locks onto a target, it chases forever until:
//   - Target dies (handled by PruneDeadTarget in ai.go)
//   - Enemy dies
//
// This creates classic Castlevania-style behavior where enemies pursue
// the player across multiple rooms once aggroed.
//
// This system eliminates the need for enemy-specific FindTargets functions
// by using configurable DetectionRange components.
func UpdateTargetDetection(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	// Get collision space for spatial queries
	collision := physics.GetCollisionSpace(world)
	if collision == nil {
		return
	}

	// Query all entities with detection capability
	entities := world.EntitiesWith(
		(*components.DetectionRange)(nil),
		(*components.AI)(nil),
		(*components.Transform)(nil),
		(*components.Facing)(nil),
	)

	for _, eid := range entities {
		detection := ecs.GetComponent[components.DetectionRange](world, eid)
		ai := ecs.GetComponent[components.AI](world, eid)
		transform := ecs.GetComponent[components.Transform](world, eid)
		facing := ecs.GetComponent[components.Facing](world, eid)

		if detection == nil || ai == nil || transform == nil {
			continue
		}

		// **FIX FOR ISSUE #2**: Skip dead entities - they should not detect targets
		if isEntityDead(world, eid) {
			continue
		}

		// Build detection rectangle based on facing direction
		rect := buildDetectionRect(transform, facing, detection)

		// **DEBUG VISUALIZATION**: Attach detection rectangle as DebugOverlay for AI debug rendering
		// This creates yellow boxes showing enemy detection ranges when Cmd+4 is pressed
		debugState := ecs.Resource[resources.DebugState](world)
		if debugState != nil && debugState.IsEnabled(resources.DebugCategoryAI) {
			components.SetDebugRect(world, eid, &rect)
		}

		// Only search for targets if we don't already have one (sticky targeting)
		// Once locked on, enemies chase forever until target dies
		if ai.TargetID == 0 {
			// Query collision space for entities in rect
			targets := findTargetsInRect(world, collision, eid, rect, detection.TeamFilter)

			// Acquire new target if found
			if len(targets) > 0 {
				ai.TargetID = targets[0]
			}
		}
		// Note: Target validation (death check) is handled by PruneDeadTarget in ai.go
	}
}

// buildDetectionRect creates vision rect from entity position, facing, and detection config.
// Returns a bump.Rect covering the detection area in world space.
func buildDetectionRect(transform *components.Transform, facing *components.Facing, detection *components.DetectionRange) bump.Rect {
	x, y, w, h := transform.X, transform.Y, transform.W, transform.H

	// Facing determines which direction is "front"
	// FlipX = false (facing left): front extends LEFT, back extends RIGHT
	// FlipX = true (facing right): front extends RIGHT, back extends LEFT
	var rectX, rectW float64
	if facing != nil && facing.FlipX {
		// Facing RIGHT: back extends left, front extends right
		rectX = x - detection.BackDistance
		rectW = detection.BackDistance + w + detection.FrontDistance
	} else {
		// Facing LEFT: front extends left, back extends right
		rectX = x - detection.FrontDistance
		rectW = detection.FrontDistance + w + detection.BackDistance
	}

	return bump.Rect{
		X: rectX,
		Y: y - detection.UpDistance,
		W: rectW,
		H: detection.UpDistance + detection.DownDistance + h,
	}
}

// findTargetsInRect queries collision space and filters by team.
// Returns list of valid target EntityIds within the rect.
func findTargetsInRect(world *ecs.World, collision *bump.Space, eid entities.EntityId, rect bump.Rect, teamFilter string) []entities.EntityId {
	// Query collision space
	cols := collision.Query(rect, func(other bump.Item) bool {
		return other != eid // Exclude self
	}, "body")

	// Filter by team and alive status
	targets := make([]entities.EntityId, 0, len(cols))
	for _, col := range cols {
		targetEID, ok := col.Other.(entities.EntityId)
		if !ok {
			continue
		}

		// **FIX FOR ISSUE #1**: Skip dead entities
		// Dead entities should not be targetable even if still in collision space
		if isEntityDead(world, targetEID) {
			continue
		}

		// Apply team filter
		if !matchesTeamFilter(world, targetEID, teamFilter) {
			continue
		}

		targets = append(targets, targetEID)
	}

	return targets
}

// isEntityDead checks if an entity is dead.
func isEntityDead(world *ecs.World, eid entities.EntityId) bool {
	// Check Health component
	if health := ecs.GetComponent[components.Health](world, eid); health != nil && health.Current <= 0 {
		return true
	}

	return false
}

// matchesTeamFilter checks if targetEID matches the teamFilter criteria.
func matchesTeamFilter(world *ecs.World, targetEID entities.EntityId, teamFilter string) bool {
	// "all" filter accepts everything
	if teamFilter == "all" {
		return true
	}

	// Get team component
	team := ecs.GetComponent[components.Team](world, targetEID)
	if team == nil {
		return false
	}

	// Check specific team filters
	if teamFilter == "player" {
		return team.Type == components.TeamPlayer
	}
	if teamFilter == "enemy" {
		return team.Type != components.TeamPlayer
	}

	return false
}
