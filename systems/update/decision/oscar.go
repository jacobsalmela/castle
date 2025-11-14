package decision

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/systems/update/entities/animation"
)

// UpdateOscar handles Oscar NPC behavior in Pure ECS style.
//
// Pure ECS - no Control component needed:
// - Hitbox initialization done directly in system
// - Death handled by Pure ECS death animation (no Control.Die)
// - Oscar stays as solid obstacle after death (never removed)
//
// Required Components:
//   - Oscar: Entity-specific state (DeadText, death flags)
//   - Health: Hit points for damage/death detection
//   - Animation: Animation state for death frame
//   - Textbox: Dialogue display and interaction
//   - Collider: Physics state (unmovable/solid)
//   - Hitbox: Collision detection
//   - Facing: Sprite orientation
//
// (Removed: Control, ControlState)
//
// All components are passed as separate parameters to helper functions.
func UpdateOscar(world *ecs.World, _ interface{}, dt float64) {
	if world == nil || dt == 0 {
		return
	}
	for _, eid := range world.EntitiesWith((*components.Oscar)(nil)) {
		oscar := ecs.GetComponent[components.Oscar](world, eid)
		health := ecs.GetComponent[components.Health](world, eid)
		if oscar == nil || health == nil {
			continue
		}

		// Initialize hitbox if needed (Pure ECS - no Control)
		ensureOscarInit(world, eid, oscar)

		// Handle death behavior and animation
		updateOscarBehavior(world, eid, oscar, health, dt)
	}
}

func ensureOscarInit(world *ecs.World, eid entities.EntityId, oscar *components.Oscar) {
	if oscar == nil || oscar.HitboxInited {
		return
	}

	// Fetch only components needed for hitbox initialization
	anim := ecs.GetComponent[components.Animation](world, eid)
	hitbox := ecs.GetComponent[components.Hitbox](world, eid)
	facing := ecs.GetComponent[components.Facing](world, eid)

	if anim == nil || hitbox == nil || facing == nil {
		return
	}

	// Initialize hitbox with correct facing
	oscarInitHitbox(hitbox, anim, facing)
	oscar.HitboxInited = true
}

func oscarInitHitbox(hitbox *components.Hitbox, anim *components.Animation, facing *components.Facing) {
	if hitbox == nil || anim == nil || facing == nil {
		return
	}
	// Extract hurtbox with correct facing direction using Pure ECS helper
	if hurtboxRect, err := animation.ExtractSlice(anim, "hurtbox", facing.FlipX, false); err == nil {
		hitbox.Boxes = append(hitbox.Boxes, components.HitboxRect{
			X:       hurtboxRect.X,
			Y:       hurtboxRect.Y,
			W:       hurtboxRect.W,
			H:       hurtboxRect.H,
			Contact: components.Hit,
		})
	}
}

func updateOscarBehavior(world *ecs.World, eid entities.EntityId, oscar *components.Oscar, health *components.Health, dt float64) {
	// Handle death behavior once when Oscar first dies
	if health.Current <= 0 && !oscar.DeathHandled {
		handleOscarDeath(world, eid, oscar, dt)
		oscar.DeathHandled = true
	}

	// Continue death animation if dead (keeps stagger animation playing)
	if health.Current <= 0 {
		freezeOscarDeathAnimation(world, eid)
	}
}

func handleOscarDeath(world *ecs.World, eid entities.EntityId, oscar *components.Oscar, dt float64) {
	anim := ecs.GetComponent[components.Animation](world, eid)
	textbox := ecs.GetComponent[components.TextboxData](world, eid)

	// Disable gravity so dead Oscar doesn't fall through the ground
	physics := ecs.GetComponent[components.Physics](world, eid)
	if physics != nil && physics.GravityEnabled {
		physics.GravityEnabled = false
		physics.Velocity.X = 0
		physics.Velocity.Y = 0
	}

	// Trigger stagger animation
	if anim != nil && anim.Data != nil {
		// Play stagger animation using Pure ECS helper
		animation.SetAnimationState(anim, "stagger")
		anim.Data.PlaySpeed = 0 // Freeze on death frame
	}

	// Swap to dead dialogue if set
	if oscar.DeadText != "" && textbox != nil {
		// Pure ECS: inline SetText - update text and reset pagination
		textbox.Text = oscar.DeadText
		textbox.AdvanceState = 0
		textbox.AdvanceFlickerTimer = 0
		textbox.AdvanceFlicker = false
		textbox.ProcessedText = "" // Clear to trigger reprocessing
	}

	// Hide interaction indicator
	if textbox != nil {
		textbox.Indicator = false
	}
}

func freezeOscarDeathAnimation(world *ecs.World, eid entities.EntityId) {
	anim := ecs.GetComponent[components.Animation](world, eid)
	if anim != nil && anim.Data != nil {
		// Keep Oscar paused and in stagger state
		// Don't call Die(dt) repeatedly as it would fade alpha
		// Just ensure paused state is maintained
		anim.Data.PlaySpeed = 0 // Stop animation on death frame
	}
}
