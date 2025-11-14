package decision

import (
	"math"

	"game/components"
	"game/components/ai"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/systems/update/entities/animation"
	"game/systems/update/combat"
	"game/systems/update/decision/nodes"
)

// buildBatBehaviorTree creates the behavior tree for bat enemies.
// Pattern: Circle player → Dive attack → Back off → Repeat
func buildBatBehaviorTree() ai.Node {
	// Pre-wrap actions with human-friendly names for debug visualization
	idle := ai.WrapForDebug(nodes.Idle(), "Idle", 2)

	// Circle around player for 3 seconds (Succeeder converts timeout Failure to Success)
	circle := ai.WrapForDebug(&ai.Succeeder{
		Child: &ai.Timeout{
			Duration: 3.0,
			Child:    newBatCircle(batRangeMid),
		},
	}, "Circle(3s)", 2)

	// Dive attack at player
	attack := ai.WrapForDebug(newBatDiveAttack(), "DiveAttack", 2)

	// Back off from player
	backOff := ai.WrapForDebug(newBatBackOff(), "BackOff", 2)

	// Assemble the tree: Circle → Attack → BackOff → Repeat
	tree := &ai.Repeat{
		Count: 0, // Infinite loop
		Child: &ai.Sequence{
			Children: []ai.Node{
				idle,
				circle,
				attack,
				backOff,
			},
		},
	}

	// Wrap tree with debug tracking for visualization (only in debug builds)
	return ai.WrapForDebug(tree, "BatBT", 0)
}

// newBatCircle creates a circular stalk pattern around the player.
// The bat flies perpendicular to the player direction, creating a circular orbit.
// Maintains altitude by adding upward bias when too close to ground or player.
func newBatCircle(targetRange float64) *ai.Action {
	return &ai.Action{
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			bat := ecs.GetComponent[components.Bat](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)
			transform := ecs.GetComponent[components.Transform](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			if bat == nil || aiComp == nil || transform == nil || physics == nil {
				return ai.Failure
			}

			// Check if target lost
			if aiComp.TargetID == 0 {
				return ai.Failure
			}

			// Don't move if paused
			if bat.Paused {
				return ai.Running
			}

			targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
			if targetTransform == nil {
				return ai.Failure
			}

			// Calculate vector from bat to player
			myX, myY := transform.X+transform.W/2, transform.Y+transform.H/2
			targetX, targetY := targetTransform.X+targetTransform.W/2, targetTransform.Y+targetTransform.H/2
			dx := targetX - myX
			dy := targetY - myY
			distance := math.Sqrt(dx*dx + dy*dy)

			// Normalize direction vector
			if distance > 0.1 {
				dx /= distance
				dy /= distance
			}

			// Calculate tangent vector (perpendicular to direction)
			// Rotate 90 degrees: (x,y) -> (-y,x)
			tangentX := -dy
			tangentY := dx

			// Calculate desired altitude (fly ABOVE player, not at same height)
			const minAltitudeAbovePlayer = 30.0
			altitudeDiff := (targetY - myY) // Negative if bat is above player

			// Mix tangent motion with approach/retreat based on distance
			var moveX, moveY float64
			if distance < targetRange-5 {
				// Too close - move away from player + tangent + stay high
				moveX = -dx*0.4 + tangentX*0.6
				moveY = -dy*0.4 + tangentY*0.6
			} else if distance > targetRange+5 {
				// Too far - move toward player + tangent + stay high
				moveX = dx*0.6 + tangentX*0.4
				moveY = dy*0.6 + tangentY*0.4
			} else {
				// Perfect range - pure circular motion
				moveX = tangentX
				moveY = tangentY
			}

			// CRITICAL: Add strong upward bias if bat is at or below player altitude
			if altitudeDiff >= -5 { // Bat is at or below player
				moveY -= 1.0 // Strong upward force
			} else if altitudeDiff >= -minAltitudeAbovePlayer { // Not high enough yet
				moveY -= 0.5 // Moderate upward force
			}

			// Screen boundary avoidance: push away from edges
			const edgeBuffer = 20.0 // Distance from edge to start avoiding

			// Get viewport/camera info (assume 320x180 viewport - standard game resolution)
			const viewportWidth = 320.0
			const viewportHeight = 180.0

			// Check proximity to left/right edges (in world space relative to player)
			horizontalDistToPlayer := math.Abs(myX - targetX)
			if horizontalDistToPlayer > viewportWidth/2-edgeBuffer {
				// Too close to horizontal edge - push toward center (toward player)
				if myX < targetX {
					moveX += 0.5 // Push right (toward player)
				} else {
					moveX -= 0.5 // Push left (toward player)
				}
			}

			// Check proximity to top edge (don't fly too high)
			if myY < targetY-viewportHeight/2+edgeBuffer {
				moveY += 1.0 // Push down
			}

			// Check proximity to bottom edge (don't fly too low)
			if myY > targetY+viewportHeight/2-edgeBuffer {
				moveY -= 1.0 // Push up
			}

			// Tile collision avoidance: check if bat is about to hit a tile
			// Simple check: if bat is very close to ground level, push upward strongly
			const groundLevel = 16.0 // Typical tile height
			if transform.Y+transform.H > targetY+targetTransform.H+groundLevel {
				// Bat is descending toward ground - push up aggressively
				moveY -= 2.0
			}

			// Normalize and apply speed
			moveLen := math.Sqrt(moveX*moveX + moveY*moveY)
			if moveLen > 0.01 {
				moveX = (moveX / moveLen) * batSpeed
				moveY = (moveY / moveLen) * batSpeed
			}

			physics.Velocity.X = moveX
			physics.Velocity.Y = moveY

			return ai.Running // Timeout will end this
		},
	}
}

// newBatDiveAttack creates the dive attack action for the bat.
// This triggers the attack animation and dives toward the player at 2x speed.
func newBatDiveAttack() *ai.Action {
	elapsed := 0.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			elapsed = 0.0

			bat := ecs.GetComponent[components.Bat](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			anim := ecs.GetComponent[components.Animation](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)

			if bat == nil || anim == nil || facing == nil || aiComp == nil {
				return
			}

			// Start attack animation
			animation.SetAnimationState(anim, "Attack")

			// Pause bat during attack setup
			animation.SetStateEffect(anim, func() func() {
				bat.Paused = true
				return func() { bat.Paused = false }
			}, "Attack")

			// Register hitbox slice for damage
			batRegisterAttackSlice(world, eid, facing, anim)

			// Register frame callback for dive trigger (frame 1)
			animation.RegisterFrameCallback(anim, 1, func() {
				vx, vy := batTargetAngleComponents(world, eid, aiComp, batSpeed)
				SetBodyVelocity(world, eid, vx, vy)
			})
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			elapsed += dt

			aiComp := ecs.GetComponent[components.AI](world, eid)
			facing := ecs.GetComponent[components.Facing](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			// Check if target lost
			if aiComp == nil || aiComp.TargetID == 0 {
				return ai.Success
			}

			if physics == nil {
				return ai.Failure
			}

			// Dive faster during attack (2x speed)
			compX, compY := batTargetAngleComponents(world, eid, aiComp, batSpeed*2)
			physics.Velocity.X = compX
			physics.Velocity.Y = compY

			// Update facing direction toward target
			if facing != nil {
				transform := ecs.GetComponent[components.Transform](world, eid)
				targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
				if transform != nil && targetTransform != nil {
					x, _, w, _ := transform.Rect()
					facing.FlipX = targetTransform.X+targetTransform.W/2 > x+w/2
				}
			}

			// Attack completes after duration
			if elapsed >= batAttackTime {
				return ai.Success
			}

			return ai.Running
		},
	}
}

// newBatBackOff creates a back off action that moves the bat away from the player.
// The bat flies upward and away from the player to create distance.
func newBatBackOff() *ai.Action {
	elapsed := 0.0
	const backOffDuration = 1.0

	return &ai.Action{
		OnStart: func(world *ecs.World, eid entities.EntityId) {
			elapsed = 0.0
		},
		OnTick: func(world *ecs.World, eid entities.EntityId, dt float64) ai.Status {
			elapsed += dt

			bat := ecs.GetComponent[components.Bat](world, eid)
			aiComp := ecs.GetComponent[components.AI](world, eid)
			transform := ecs.GetComponent[components.Transform](world, eid)
			physics := ecs.GetComponent[components.Physics](world, eid)

			if bat == nil || aiComp == nil || transform == nil || physics == nil {
				return ai.Failure
			}

			// Check if target lost
			if aiComp.TargetID == 0 {
				return ai.Failure
			}

			// Don't move if paused
			if bat.Paused {
				return ai.Running
			}

			targetTransform := ecs.GetComponent[components.Transform](world, aiComp.TargetID)
			if targetTransform == nil {
				return ai.Failure
			}

			// Calculate direction away from player
			myX, myY := transform.X+transform.W/2, transform.Y+transform.H/2
			targetX, targetY := targetTransform.X+targetTransform.W/2, targetTransform.Y+targetTransform.H/2
			dx := myX - targetX // Reversed: away from target
			dy := myY - targetY

			distance := math.Sqrt(dx*dx + dy*dy)
			if distance > 0.1 {
				dx /= distance
				dy /= distance
			}

			// Fly away and upward
			physics.Velocity.X = dx * batSpeed * 0.8
			physics.Velocity.Y = (dy - 0.5) * batSpeed // Extra upward bias

			// Complete after duration
			if elapsed >= backOffDuration {
				return ai.Success
			}

			return ai.Running
		},
	}
}

// batRegisterAttackSlice registers the hitbox slice callback for the bat's attack.
// Used by both BT and legacy systems.
func batRegisterAttackSlice(world *ecs.World, eid entities.EntityId, facing *components.Facing, anim *components.Animation) {
	if anim == nil || world == nil || facing == nil {
		return
	}
	hitbox := ecs.GetComponent[components.Hitbox](world, eid)
	var contactedPrev []*components.Hitbox
	animation.RegisterSliceCallback(anim, components.HitboxSliceName, facing.FlipX, false, func(x, y, w, h float64, firstFrame bool) {
		if firstFrame {
			contactedPrev = contactedPrev[:0]
		}
		rect := bump.Rect{X: x, Y: y, W: w, H: h}
		contact, contacted := combat.ResolveHitboxArea(world, eid, rect, batAttackDamage, BuildAttackFilters(hitbox, contactedPrev))
		contactedPrev = CollectUniqueContacts(contactedPrev, contacted)
		if len(contacted) > 0 {
			// Bat uses AddVelocity for both X and Y (flying enemy), so we handle Y separately
			ApplyMeleeImpulse(world, eid, facing, contact, batAttackPush, batAttackReact)
		}
	})
}

// batTargetAngleComponents calculates X and Y velocity components toward the target.
// Used by both BT and legacy systems.
func batTargetAngleComponents(world *ecs.World, eid entities.EntityId, ai *components.AI, speed float64) (float64, float64) {
	if ai == nil || ai.TargetID == 0 {
		return 0, 0
	}
	transform := ecs.GetComponent[components.Transform](world, eid)
	if transform == nil {
		return 0, 0
	}
	// Get target's Transform via ECS
	targetTransform := ecs.GetComponent[components.Transform](world, ai.TargetID)
	if targetTransform == nil {
		return 0, 0
	}
	bx, by, bw, bh := transform.Rect()
	angle := math.Atan2((targetTransform.Y)-(by+bh), (targetTransform.X+targetTransform.W/2)-(bx+bw/2))
	return speed * math.Cos(angle), speed * math.Sin(angle)
}
