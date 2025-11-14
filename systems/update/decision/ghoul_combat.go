package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/prefabs"
	"game/systems/update/entities/animation"
	"game/systems/update/combat"
)

const (
	ghoulDamage      = 18.0
	ghoulAttackReact = 10.0
	ghoulAttackPush  = 10.0
)

func ghoulEnterAttack(world *ecs.World, eid entities.EntityId, ghoul *components.Ghoul, facing *components.Facing, tag string, anim *components.Animation) {
	if ghoul == nil || anim == nil {
		return
	}
	animation.SetAnimationState(anim, tag)
	animation.SetStateEffect(anim, func() func() {
		ghoul.Paused = true
		return func() { ghoul.Paused = false }
	}, tag)
	ghoulRegisterAttackSlice(world, eid, facing, anim)
	ghoulSetVelocity(world, eid, 0, 0)
	ai := ecs.GetComponent[components.AI](world, eid)
	ghoulEnsureFacingTarget(world, eid, facing, ai)
}

func ghoulRegisterAttackSlice(world *ecs.World, eid entities.EntityId, facing *components.Facing, anim *components.Animation) {
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
		contact, contacted := combat.ResolveHitboxArea(world, eid, rect, ghoulDamage, BuildAttackFilters(hitbox, contactedPrev))
		contactedPrev = CollectUniqueContacts(contactedPrev, contacted)
		if len(contacted) > 0 {
			ApplyMeleeImpulse(world, eid, facing, contact, ghoulAttackPush, ghoulAttackReact)
		}
	})
}

func ghoulStartThrow(world *ecs.World, eid entities.EntityId, ghoul *components.Ghoul, facing *components.Facing, anim *components.Animation) {
	if ghoul == nil || anim == nil {
		return
	}
	if ghoul.Rocks <= 0 || ghoul.ThrowCooldown > 0 {
		return
	}

	// Pause AI and stop movement during throw
	ghoul.Paused = true
	ghoulSetVelocity(world, eid, 0, 0)

	fired := false
	animation.SetAnimationState(anim, "Throw")
	animation.SetStateEffect(anim, func() func() {
		// Already paused above, but keep for consistency
		ghoul.Paused = true
		return func() { ghoul.Paused = false }
	}, "Throw")
	animation.RegisterFrameCallback(anim, ghoulThrowFrame, func() {
		if fired || ghoul.Rocks <= 0 {
			return
		}
		ghoulSpawnRock(world, eid, ghoul)
		fired = true
	})
}

func ghoulSpawnRock(world *ecs.World, eid entities.EntityId, ghoul *components.Ghoul) {
	if world == nil || eid == 0 || ghoul == nil {
		return
	}
	t := ecs.GetComponent[components.Transform](world, eid)
	facing := ecs.GetComponent[components.Facing](world, eid)
	if t == nil || facing == nil {
		return
	}

	// Determine throw direction from facing
	// FlipX=true means facing RIGHT (+X direction), FlipX=false means facing LEFT (-X direction)
	directionX := -1.0 // default: left (negative X)
	if facing.FlipX {
		directionX = 1.0 // facing right (positive X)
	}

	// Calculate spawn offset based on facing direction
	// Spawn rock ahead of ghoul (in the direction they're facing)
	spawnOffsetX := 8.0 * directionX // 8 pixels ahead in facing direction
	spawnX := t.X + spawnOffsetX
	spawnY := t.Y - 4 // 4 pixels above ghoul's feet

	// Spawn rock with proper direction and owner
	prefabs.SpawnRockWithDirection(world, spawnX, spawnY, eid, directionX)
	ghoul.Rocks--
	ghoul.ThrowCooldown = ghoulThrowCooldown
}
