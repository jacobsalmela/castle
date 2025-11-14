package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
)

func ghoulEnsureFacingTarget(world *ecs.World, eid entities.EntityId, facing *components.Facing, ai *components.AI) float64 {
	if world == nil || eid == 0 || ai == nil || ai.TargetID == 0 {
		return 0
	}
	t := ecs.GetComponent[components.Transform](world, eid)
	if t == nil {
		return 0
	}
	// Get target's Transform via ECS directly
	targetTransform := ecs.GetComponent[components.Transform](world, ai.TargetID)
	if targetTransform == nil {
		return 0
	}
	targetMid := targetTransform.X + targetTransform.W/2
	selfMid := t.X + t.W/2
	facingRight := targetMid > selfMid
	if facing != nil {
		facing.FlipX = facingRight
	}
	if facingRight {
		return 1
	}
	return -1
}

func ghoulForwardStep(world *ecs.World, eid entities.EntityId, facing *components.Facing, delta float64) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	if facing != nil && facing.FlipX {
		physics.Velocity.X += delta
		if !physics.Grounded {
			physics.Velocity.X -= delta * 2
		}
	} else {
		physics.Velocity.X -= delta
		if !physics.Grounded {
			physics.Velocity.X += delta * 2
		}
	}
}

func ghoulBackStep(world *ecs.World, eid entities.EntityId, facing *components.Facing, delta float64) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	if facing != nil && facing.FlipX {
		physics.Velocity.X -= delta
		if !physics.Grounded {
			physics.Velocity.X += delta * 2
		}
	} else {
		physics.Velocity.X += delta
		if !physics.Grounded {
			physics.Velocity.X -= delta * 2
		}
	}
}

func ghoulExecuteJump(world *ecs.World, eid entities.EntityId, facing *components.Facing) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	physics.MaxVelocity.X = ghoulMaxSpeed * 2
	physics.Velocity.Y = -ghoulSpeed / 2
	physics.Grounded = false
	dir := -1.0
	if facing != nil && facing.FlipX {
		dir = 1.0
	}
	physics.Velocity.X += dir * ghoulMaxSpeed * 2
}

func ghoulUpdateJump(world *ecs.World, eid entities.EntityId, facing *components.Facing, anim *components.Animation, ai *components.AI, dt float64) bool {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return true
	}
	grounded := false
	dir := ghoulEnsureFacingTarget(world, eid, facing, ai)
	if dir == 0 {
		dir = -1
		if facing != nil && facing.FlipX {
			dir = 1
		}
	}
	physics.Velocity.X += dir * ghoulSpeed * dt
	grounded = physics.Grounded
	return grounded && (anim == nil || anim.State != "AttackShort")
}

func ghoulSetVelocity(world *ecs.World, eid entities.EntityId, vx, vy float64) {
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics == nil {
		return
	}
	physics.Velocity.X = vx
	physics.Velocity.Y = vy
}
