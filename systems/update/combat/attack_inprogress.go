package combat

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/resources"
	"game/systems/update/entities/animation"
)

// AttackInProgress should be run after animation updates. It processes any
// entities with an AttackActive component by checking the Anim slice named
// components.HitboxSliceName on the current frame and resolving contacts.
func AttackInProgress(world *ecs.World, dt float64) {
	if world == nil {
		return
	}

	activeAttacks := world.EntitiesWith((*components.AttackActive)(nil))

	for _, eid := range activeAttacks {
		aa := ecs.GetComponent[components.AttackActive](world, eid)
		anim := ecs.GetComponent[components.Animation](world, eid)
		if aa == nil || anim == nil {
			continue
		}

		// Initialize transient state if needed (lazy initialization)
		if aa.Contacted == nil {
			aa.Contacted = make(map[*components.Hitbox]struct{})
			aa.OnceApplied = false
			aa.LastSlicePresent = false
		}

		// Get flip state for hitbox positioning
		facing := ecs.GetComponent[components.Facing](world, eid)
		flipX := false
		if facing != nil {
			flipX = facing.FlipX
		}

		// Determine if the hitbox slice is present on this frame
		sliceRect, err := animation.ExtractSlice(anim, components.HitboxSliceName, flipX, false)
		present := err == nil

		// If the slice just became present after being absent, treat this as a segmented
		// start and clear the contacted set to allow a fresh burst.
		if present && !aa.LastSlicePresent {
			aa.Contacted = make(map[*components.Hitbox]struct{})
		}

		if present {
			processPresentSlice(world, eid, aa, sliceRect)
		}

		// Heuristic: if the anim's state no longer matches the AttackTag, consider attack finished
		if aa.AttackTag != "" && anim.State != aa.AttackTag {
			ecs.RemoveComponent[components.AttackActive](world, entities.EntityId(eid))
		}

		// remember last slice presence
		aa.LastSlicePresent = present
	}
}

// processPresentSlice handles an attack slice when present: resolves contacts,
// coalesces them, triggers camera shake on new contacts, and applies once-only
// effects such as stamina deduction and body impulse.
func processPresentSlice(world *ecs.World, eid entities.EntityId, aa *components.AttackActive, area bump.Rect) {
	attackMult := aa.Mult + 1
	totalDamage := aa.Damage * attackMult
	maxContact, contacted := ResolveHitboxArea(world, eid, area, totalDamage, aa.FilterOut)

	// Hitbox visualization is now handled in combat.go's enqueueHitEvents
	// which has access to both attacker and victim positions

	// Add newly contacted targets and compute contact count
	coalesceContacts(aa, contacted)
	contactCount := len(aa.Contacted)

	// Cache common components
	stamina := ecs.GetComponent[components.Stamina](world, eid)
	physics := ecs.GetComponent[components.Physics](world, eid)

	// Player-only camera shake behavior
	players := world.EntitiesWith((*components.Player)(nil))
	isPlayer := len(players) > 0 && entities.EntityId(eid) == players[0]
	if isPlayer {
		if aa.ShakeNum != contactCount {
			aa.ShakeNum = contactCount
			if contactCount > 0 {
				if camera := ecs.Resource[resources.Camera](world); camera != nil {
					camera.Shake(0.1*float32(attackMult), 0.5*(attackMult))
				}
			}
		}
	}

	// Parry/poise handling.
	handleParryPoise(world, eid, aa, maxContact, totalDamage)

	// Once-only attacker-side effects (stamina deduction and body impulse).
	applyOnceEffects(world, eid, aa, maxContact, attackMult, stamina, physics)
}

// handleParryPoise applies poise reduction and stagger on the attacker when a ParryBlock occurs.
func handleParryPoise(world *ecs.World, eid entities.EntityId, aa *components.AttackActive, maxContact components.ContactType, totalDamage float64) {
	if maxContact != components.ParryBlock {
		return
	}
	// fetch components lazily to keep the helper signature small
	anim := ecs.GetComponent[components.Animation](world, eid)
	poise := ecs.GetComponent[components.Poise](world, eid)
	health := ecs.GetComponent[components.Health](world, eid)
	physics := ecs.GetComponent[components.Physics](world, eid)
	hb := ecs.GetComponent[components.Hitbox](world, eid)

	if poise == nil {
		return
	}
	poise.Current -= totalDamage
	if poise.Current > 0 {
		return
	}
	// Emulate ShieldDown parity and stagger application
	emulateShieldDownAndStagger(world, eid, anim, physics, hb, aa, totalDamage, health)
}

func emulateShieldDownAndStagger(world *ecs.World, eid entities.EntityId, anim *components.Animation, physics *components.Physics, hb *components.Hitbox, aa *components.AttackActive, totalDamage float64, health *components.Health) {
	if anim == nil {
		return
	}
	if anim.State == components.ParryBlockTag || anim.State == components.BlockTag {
		animation.SetAnimationState(anim, components.IdleTag)
		if hb != nil && len(hb.Boxes) > 0 {
			hb.Boxes = hb.Boxes[:len(hb.Boxes)-1]
		}
	}
	// Apply stagger state
	animation.SetAnimationState(anim, components.StaggerTag)
	animation.SetStateEffect(anim, func() func() {
		if anim.Data == nil {
			// No anim data available; return a no-op restore function.
			return func() { /* no-op: no anim data to restore */ }
		}
		prev := anim.Data.PlaySpeed
		anim.Data.PlaySpeed = float32(1.0) // timeMult=1 for parity

		return func() { anim.Data.PlaySpeed = prev }
	}, components.StaggerTag)

	if physics != nil && health != nil {
		force := aa.ReactForce * (totalDamage / health.Max)
		facing := ecs.GetComponent[components.Facing](world, eid)
		if facing != nil && facing.FlipX {
			force *= -1
		}
		physics.Velocity.X += force
	}
}

// applyOnceEffects deducts stamina once and applies the attacker-side impulse.
func applyOnceEffects(world *ecs.World, eid entities.EntityId, aa *components.AttackActive, maxContact components.ContactType, attackMult float64, stamina *components.Stamina, physics *components.Physics) {
	if aa.OnceApplied {
		return
	}
	if stamina != nil && aa.StaminaDamage != 0 {
		stamina.Current -= aa.StaminaDamage * attackMult
		if stamina.Current < 0 {
			stamina.Current = 0
		}
	}
	force := aa.PushForce
	if maxContact >= components.Block {
		force = aa.ReactForce
	}
	// Use Facing component for direction
	facing := ecs.GetComponent[components.Facing](world, eid)
	if facing != nil {
		if (maxContact >= components.Block && facing.FlipX) || (maxContact < components.Block && !facing.FlipX) {
			force *= -1
		}
	}
	if physics != nil {
		physics.Velocity.X += force
	}
	aa.OnceApplied = true
}

// coalesceContacts adds newly contacted targets to aa.Contacted and returns
// the number of newly added contacts.
func coalesceContacts(aa *components.AttackActive, contacted []*components.Hitbox) int {
	var newContacts int
	for _, t := range contacted {
		if t == nil {
			continue
		}
		if _, seen := aa.Contacted[t]; !seen {
			aa.Contacted[t] = struct{}{}
			newContacts++
		}
	}
	return newContacts
}

// applyNewContactsEffects triggers camera shake when new contacts occurred.
func applyNewContactsEffects(world *ecs.World, aa *components.AttackActive, newContacts int) {
	if newContacts > 0 {
		if camera := ecs.Resource[resources.Camera](world); camera != nil {
			camera.Shake(0.1*float32(1+aa.Mult), 0.5*(1+aa.Mult))
		}
	}
}
