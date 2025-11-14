package decision

import (
	"bytes"

	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/prefabs"
	"game/systems/update/entities/animation"
	"game/systems/update/combat"
)

func knightRegisterAttackSlice(world *ecs.World, eid entities.EntityId, knight *components.Knight, facing *components.Facing, anim *components.Animation) {
	if knight == nil || anim == nil || world == nil || facing == nil {
		return
	}
	hitbox := ecs.GetComponent[components.Hitbox](world, eid)
	var contactedPrev []*components.Hitbox
	animation.RegisterSliceCallback(anim, components.HitboxSliceName, facing.FlipX, false, func(x, y, w, h float64, firstFrame bool) {
		if firstFrame {
			contactedPrev = contactedPrev[:0]
		}
		rect := bump.Rect{X: x, Y: y, W: w, H: h}
		contact, contacted := combat.ResolveHitboxArea(world, eid, rect, knightAttackDamage, BuildAttackFilters(hitbox, contactedPrev))
		contactedPrev = CollectUniqueContacts(contactedPrev, contacted)
		if len(contacted) > 0 {
			ApplyMeleeImpulse(world, eid, facing, contact, knightAttackPush, knightAttackReact)
		}
	})
}

func knightForwardStep(world *ecs.World, eid entities.EntityId, knight *components.Knight, facing *components.Facing, delta float64) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	dir := KnightDirectionTowardsTarget(world, eid, knight, facing)
	physics.Velocity.X += dir * delta
	if !physics.Grounded {
		physics.Velocity.X += dir * delta
	}
}

func knightBackStep(world *ecs.World, eid entities.EntityId, knight *components.Knight, facing *components.Facing, delta float64) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	dir := KnightDirectionTowardsTarget(world, eid, knight, facing)
	physics.Velocity.X -= dir * delta
	if !physics.Grounded {
		physics.Velocity.X -= dir * delta
	}
}

// KnightSetVelocity sets the knight's body velocity.
// Used by the knight behavior tree.
func KnightSetVelocity(world *ecs.World, eid entities.EntityId, vx, vy float64) {
	SetBodyVelocity(world, eid, vx, vy)
}

// KnightSpawnRock spawns a rock projectile from the knight's position.
// Used by the knight behavior tree for throw attacks.
func KnightSpawnRock(world *ecs.World, eid entities.EntityId, knight *components.Knight) {
	if world == nil || knight == nil {
		return
	}
	transform := ecs.GetComponent[components.Transform](world, eid)
	if transform == nil {
		return
	}
	x, y := transform.X, transform.Y
	prefabs.SpawnRockProjectile(world, x-2, y-4, eid)
}

// KnightEnsureFacingTarget ensures the knight is facing its target.
// Returns the direction to the target (1.0 for right, -1.0 for left).
// Used by the knight behavior tree.
func KnightEnsureFacingTarget(world *ecs.World, eid entities.EntityId, knight *components.Knight, facing *components.Facing) float64 {
	ai := ecs.GetComponent[components.AI](world, eid)
	if knight == nil || ai == nil || ai.TargetID == 0 {
		if facing != nil && facing.FlipX {
			return 1
		}
		return -1
	}
	transform := ecs.GetComponent[components.Transform](world, eid)
	if transform == nil {
		if facing != nil && facing.FlipX {
			return 1
		}
		return -1
	}
	// Get target's Transform via ECS directly
	targetTransform := ecs.GetComponent[components.Transform](world, ai.TargetID)
	if targetTransform == nil {
		return 0
	}
	x, w := transform.X, transform.W
	targetMid := targetTransform.X + targetTransform.W/2
	selfMid := x + w/2
	facingRight := targetMid > selfMid
	if facing != nil {
		facing.FlipX = facingRight
	}
	if facingRight {
		return 1
	}
	return -1
}

// KnightDirectionTowardsTarget calculates the direction to the knight's target.
// Returns 1.0 for right, -1.0 for left.
// Used by the knight behavior tree.
func KnightDirectionTowardsTarget(world *ecs.World, eid entities.EntityId, knight *components.Knight, facing *components.Facing) float64 {
	ai := ecs.GetComponent[components.AI](world, eid)
	if knight == nil || ai == nil {
		return -1
	}
	dir := KnightEnsureFacingTarget(world, eid, knight, facing)
	if dir == 0 {
		if facing != nil && facing.FlipX {
			return 1
		}
		return -1
	}
	return dir
}

func knightTintRed(anim *components.Animation) {
	if anim == nil || anim.Image == nil {
		return
	}
	image := anim.Image
	size := image.Bounds().Size()
	pixels := make([]byte, size.X*size.Y*4)
	image.ReadPixels(pixels)
	image.WritePixels(bytes.ReplaceAll(pixels, []byte{91, 110, 225}, []byte{172, 50, 50}))
}
