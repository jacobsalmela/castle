package combat

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
	"game/resources"
	"game/systems/update/entities/animation"
)

// ApplyPlayerBlock processes Block and ParryBlock combat events for player entities.
// Migrated from entity/actor/actor.go Control.Block() method.
//
// Block mechanics:
// - Chip damage: -damage/10 health
// - Stamina drain: -damage stamina
// - Knockback: reactForce/2 (directional based on attacker position)
// - Stagger: if stamina depleted, stagger with increased force
// - ParryBlock: only applies chip damage and stamina, no knockback or stagger
func ApplyPlayerBlock(world *ecs.World, events []resources.HitEvent) {
	if world == nil || len(events) == 0 {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		return
	}

	for _, evt := range events {
		if !shouldProcessBlockEvent(world, evt) {
			continue
		}

		processPlayerBlockEvent(world, evt, cfg)
	}
}

// shouldProcessBlockEvent determines if the event should be processed by the block system.
// Returns true only for Block/ParryBlock events targeting the player that haven't been handled.
func shouldProcessBlockEvent(world *ecs.World, evt resources.HitEvent) bool {
	if evt.Handled {
		return false
	}

	contact := components.ContactType(evt.Contact)
	if contact != components.Block && contact != components.ParryBlock {
		return false
	}

	return ecs.HasComponent[components.Player](world, evt.Target)
}

// processPlayerBlockEvent handles a block event by applying damage, knockback, and stagger.
func processPlayerBlockEvent(world *ecs.World, evt resources.HitEvent, cfg *config.Config) {
	comps := getPlayerBlockComponents(world, evt.Target)
	if comps == nil {
		return
	}

	shieldDown(comps.anim, comps.hitbox)
	applyBlockDamage(comps.health, comps.stamina, comps.physics, evt.Damage)

	// ParryBlock only applies damage, no knockback or stagger
	if components.ContactType(evt.Contact) == components.ParryBlock {
		return
	}

	applyBlockKnockback(comps.physics, evt, cfg.Actor.ReactForce)

	if comps.stamina.Current < 0 {
		applyBlockStagger(world, evt.Target, comps, evt, cfg.Actor.ReactForce)
	}
}

// playerBlockComponents holds all components needed for block processing.
type playerBlockComponents struct {
	health  *components.Health
	stamina *components.Stamina
	anim    *components.Animation
	physics *components.Physics
	hitbox  *components.Hitbox
}

// getPlayerBlockComponents retrieves all required components for block processing.
// Returns nil if any required component is missing.
func getPlayerBlockComponents(world *ecs.World, target entities.EntityId) *playerBlockComponents {
	health := ecs.GetComponent[components.Health](world, target)
	stamina := ecs.GetComponent[components.Stamina](world, target)
	anim := ecs.GetComponent[components.Animation](world, target)
	physics := ecs.GetComponent[components.Physics](world, target)
	hitbox := ecs.GetComponent[components.Hitbox](world, target)

	if health == nil || stamina == nil || anim == nil || physics == nil || hitbox == nil {
		return nil
	}

	return &playerBlockComponents{health, stamina, anim, physics, hitbox}
}

// applyBlockDamage applies chip damage to health and drains stamina.
// Disables physics when player dies.
func applyBlockDamage(health *components.Health, stamina *components.Stamina, physics *components.Physics, damage float64) {
	// Apply chip damage
	health.Current -= damage / 10
	if health.Current < 0 {
		health.Current = 0
	}

	// Disable physics on death
	if health.Current <= 0 {
		physics.GravityEnabled = false
		physics.Velocity.X = 0
		physics.Velocity.Y = 0
	}

	// Drain stamina
	stamina.Current -= damage
	if stamina.Current < 0 {
		stamina.Current = 0
	}
}

// applyBlockKnockback applies knockback force to the player when blocking.
// Force is directional based on attacker position.
func applyBlockKnockback(physics *components.Physics, evt resources.HitEvent, reactForce float64) {
	force := reactForce / 2
	dx := (evt.TargetRect.X + evt.TargetRect.W/2) - (evt.AttackRect.X + evt.AttackRect.W/2)
	if dx < 0 {
		force *= -1
	}
	physics.Velocity.X += force
}

// applyBlockStagger applies stagger when stamina is depleted from blocking.
func applyBlockStagger(world *ecs.World, target entities.EntityId, comps *playerBlockComponents, evt resources.HitEvent, reactForce float64) {
	force := reactForce * (evt.Damage / comps.health.Max)
	facing := ecs.GetComponent[components.Facing](world, target)
	stagger(comps.anim, comps.physics, comps.hitbox, facing, force, false, 2.0)
}

// shieldDown exits blocking state by setting animation to Idle and popping the hitbox.
// Migrated from entity/actor/actor.go Control.ShieldDown() method.
func shieldDown(anim *components.Animation, hitbox *components.Hitbox) {
	if anim == nil || hitbox == nil {
		return
	}

	// Check if currently blocking
	isBlocking := anim.State == components.BlockTag || anim.State == components.ParryBlockTag
	if !isBlocking {
		return
	}

	// Exit blocking state using animation system helper
	animation.SetAnimationState(anim, components.IdleTag)
	// Pure ECS: Remove block/parry shield hitbox (last box added)
	if len(hitbox.Boxes) > 0 {
		hitbox.Boxes = hitbox.Boxes[:len(hitbox.Boxes)-1]
	}
}

// stagger applies stagger animation and knockback with adjustable timing.
// Migrated from entity/actor/actor.go Control.Stagger() method.
//
// Parameters:
// - force: knockback force (applied to velocityX)
// - moveBack: if true and facing left (via Facing component), reverse force direction
// - timeMult: animation speed multiplier (higher = slower stagger animation)
func stagger(anim *components.Animation, physics *components.Physics, hitbox *components.Hitbox, facing *components.Facing, force float64, moveBack bool, timeMult float64) {
	if anim == nil || physics == nil {
		return
	}

	// Exit blocking state first
	shieldDown(anim, hitbox)

	// Set stagger animation using animation system helper
	animation.SetAnimationState(anim, components.StaggerTag)

	// Adjust animation speed if needed
	if timeMult != 1 {
		animation.SetStateEffect(anim, func() func() {
			prevPlaySpeed := anim.Data.PlaySpeed
			anim.Data.PlaySpeed = float32(1.0 / timeMult)
			return func() { anim.Data.PlaySpeed = prevPlaySpeed }
		}, components.StaggerTag)
	}

	// Apply knockback force (reverse if moveBack and facing left)
	if moveBack && facing != nil && facing.FlipX {
		force *= -1
	}
	physics.Velocity.X += force
}
