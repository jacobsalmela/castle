package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/systems/update/entities/animation"
	"game/systems/update/combat"
)

// EnterMeleeAttack handles the generic melee attack setup logic.
//
// This function encapsulates the common melee attack pattern used by most enemies:
//  1. Set animation
//  2. Pause enemy during attack
//  3. Register hitbox slice callback
//  4. Stop movement
//  5. Face target
//
// Parameters:
//   - world: ECS world instance
//   - eid: Entity performing the attack
//   - behavior: MeleeAttackBehavior component with attack config (damage, forces, animation)
//   - facing: Facing component for direction
//   - anim: Animation component for attack animation
//   - ai: AI component for target tracking (optional, for facing target)
//   - pausedPtr: Pointer to enemy's Paused field (e.g., &ent.Paused, &crawler.Paused)
//
// Used by: Ent behavior tree
func EnterMeleeAttack(
	world *ecs.World,
	eid entities.EntityId,
	behavior *components.MeleeAttackBehavior,
	facing *components.Facing,
	anim *components.Animation,
	ai *components.AI,
	pausedPtr *bool,
) {
	if anim == nil || behavior == nil {
		return
	}

	// Set attack animation
	animation.SetAnimationState(anim, behavior.AnimationTag)

	// Pause enemy during attack
	if pausedPtr != nil {
		animation.SetStateEffect(anim, func() func() {
			*pausedPtr = true
			return func() { *pausedPtr = false }
		}, behavior.AnimationTag)
	}

	// Register hitbox slice callback
	registerMeleeAttackSlice(world, eid, behavior, facing, anim)

	// Stop movement during attack
	SetBodyVelocity(world, eid, 0, 0)

	// Face target if available
	ensureFacingTargetGeneric(world, eid, facing, ai)
}

// registerMeleeAttackSlice sets up the hitbox slice callback for melee damage detection.
func registerMeleeAttackSlice(
	world *ecs.World,
	eid entities.EntityId,
	behavior *components.MeleeAttackBehavior,
	facing *components.Facing,
	anim *components.Animation,
) {
	if anim == nil || world == nil || facing == nil || behavior == nil {
		return
	}

	hitbox := ecs.GetComponent[components.Hitbox](world, eid)
	var contactedPrev []*components.Hitbox

	sliceName := behavior.HitboxSliceName
	if sliceName == "" {
		sliceName = components.HitboxSliceName
	}

	animation.RegisterSliceCallback(anim, sliceName, facing.FlipX, false, func(x, y, w, h float64, firstFrame bool) {
		// Reset contacted list on first frame of slice
		if firstFrame {
			contactedPrev = contactedPrev[:0]
		}

		// Build hitbox rect and detect contacts
		rect := bump.Rect{X: x, Y: y, W: w, H: h}
		contact, contacted := combat.ResolveHitboxArea(
			world,
			eid,
			rect,
			behavior.Damage,
			BuildAttackFilters(hitbox, contactedPrev),
		)

		// Remember contacted targets to prevent multi-hit
		contactedPrev = CollectUniqueContacts(contactedPrev, contacted)

		// Apply impulse forces if hit anything
		if len(contacted) > 0 {
			ApplyMeleeImpulse(world, eid, facing, contact, behavior.PushForce, behavior.ReactForce)
		}
	})
}

// ensureFacingTargetGeneric makes the enemy face their AI target before attacking.
func ensureFacingTargetGeneric(world *ecs.World, eid entities.EntityId, facing *components.Facing, ai *components.AI) {
	if world == nil || eid == 0 || facing == nil || ai == nil || ai.TargetID == 0 {
		return
	}

	transform := ecs.GetComponent[components.Transform](world, eid)
	targetTransform := ecs.GetComponent[components.Transform](world, ai.TargetID)
	if transform == nil || targetTransform == nil {
		return
	}

	// Calculate midpoints
	selfMid := transform.X + transform.W/2
	targetMid := targetTransform.X + targetTransform.W/2

	// Update facing direction
	facing.FlipX = targetMid > selfMid
}
