package combat

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/bump"
	"game/resources"
	"game/systems/update/physics"
)

func init() {
	// Phase 4 complete: All boss enemies migrated to Pure ECS.
	// ResolveHitboxAreaBridge removed - no longer needed.
	// All entities now call ResolveHitboxArea directly via attack systems.
}

type resolvedHit struct {
	targetID entities.EntityId
	hitbox   *components.Hitbox
	rect     bump.Rect
	contact  components.ContactType
}

type targetAccumulator struct {
	hits          map[*components.Hitbox]resolvedHit
	contactedSet  map[*components.Hitbox]struct{}
	contactedList *[]*components.Hitbox
}

// ResolveHitboxArea resolves hitbox collision detection using components.
// It returns the strongest contact type along with every contacted hitbox (excluding filtered hitboxes).
func ResolveHitboxArea(world *ecs.World, attacker entities.EntityId, area bump.Rect, damage float64, filterOut []*components.Hitbox) (components.ContactType, []*components.Hitbox) {
	if world == nil {
		return components.Hit, nil
	}

	attackerHitbox := ecs.GetComponent[components.Hitbox](world, attacker)
	if attackerHitbox == nil {
		return components.Hit, nil
	}

	originRect, ok := ownerRect(world, attacker)
	if !ok {
		return components.Hit, nil
	}
	attackRect := bump.Rect{X: originRect.X + area.X, Y: originRect.Y + area.Y, W: area.W, H: area.H}

	filterSet := makeFilterSet(filterOut)
	hits, contacted, maxContact := collectHitContacts(world, attacker, attackRect, filterSet)
	maxContact = maxContactWithMap(world, attackRect, maxContact)
	enqueueHitEvents(world, attacker, attackRect, hits, damage)

	return maxContact, contacted
}

func makeFilterSet(filterOut []*components.Hitbox) map[*components.Hitbox]struct{} {
	if len(filterOut) == 0 {
		return nil
	}
	out := make(map[*components.Hitbox]struct{}, len(filterOut))
	for _, hitbox := range filterOut {
		if hitbox == nil {
			continue
		}
		out[hitbox] = struct{}{}
	}
	return out
}

func collectHitContacts(world *ecs.World, attacker entities.EntityId, attackRect bump.Rect, filterSet map[*components.Hitbox]struct{}) ([]resolvedHit, []*components.Hitbox, components.ContactType) {
	hits := make(map[*components.Hitbox]resolvedHit)
	contactedSet := map[*components.Hitbox]struct{}{}
	contactedList := []*components.Hitbox{}
	acc := targetAccumulator{hits: hits, contactedSet: contactedSet, contactedList: &contactedList}
	maxContact := components.Hit

	entitiesChecked := 0
	for _, eid := range world.EntitiesWith((*components.Hitbox)(nil)) {
		entitiesChecked++
		target := ecs.GetComponent[components.Hitbox](world, eid)
		if skipTarget(world, eid, attacker, target, filterSet) {
			continue
		}
		if len(target.Boxes) == 0 {
			continue
		}
		origin, ok := ownerRect(world, eid)
		if !ok {
			continue
		}
		maxContact = accumulateTargetHits(eid, target, origin, attackRect, &acc, maxContact)
	}

	collected := make([]resolvedHit, 0, len(hits))
	for _, info := range hits {
		collected = append(collected, info)
	}

	return collected, contactedList, maxContact
}

func skipTarget(world *ecs.World, eid, attacker entities.EntityId, target *components.Hitbox, filterSet map[*components.Hitbox]struct{}) bool {
	if eid == attacker || target == nil {
		return true
	}
	if filterSet != nil {
		if _, skip := filterSet[target]; skip {
			return true
		}
	}

	// Skip friendly fire - entities on the same team don't damage each other
	// This prevents: player hitting player, enemy hitting enemy, neutral hitting neutral
	// But allows: player hitting enemy, player hitting neutral (doors/chests), enemy hitting player, etc.
	attackerTeam := ecs.GetComponent[components.Team](world, attacker)
	targetTeam := ecs.GetComponent[components.Team](world, eid)

	if attackerTeam != nil && targetTeam != nil {
		// Same team = skip (no friendly fire)
		if attackerTeam.Type == targetTeam.Type {
			return true
		}
		// Different teams = allow (player can hit enemies, enemies can hit player, anyone can hit neutral)
	}
	// If either entity has no team component, allow the interaction (e.g., spikes have no team)

	// Skip dead entities
	if health := ecs.GetComponent[components.Health](world, eid); health != nil && health.Current <= 0 {
		return true
	}

	return false

}

func accumulateTargetHits(eid entities.EntityId, target *components.Hitbox, origin bump.Rect, attackRect bump.Rect, acc *targetAccumulator, current components.ContactType) components.ContactType {
	boxCount := len(target.Boxes)
	overlapsFound := 0

	for i := 0; i < boxCount; i++ {
		box := target.Boxes[i]
		rect := bump.Rect{
			X: origin.X + box.X,
			Y: origin.Y + box.Y,
			W: box.W,
			H: box.H,
		}

		// Debug logging removed - Pure ECS migration
		// Collision checks can be visualized with debug_collision.go
		overlaps := bump.Overlaps(attackRect, rect)

		if !overlaps {
			continue
		}

		overlapsFound++
		contactType := box.ResolveContact()
		info := acc.hits[target]
		if info.hitbox == nil {
			info = resolvedHit{targetID: eid, hitbox: target, rect: rect, contact: contactType}
		}
		if contactType > info.contact || info.rect == (bump.Rect{}) {
			info.contact = contactType
			info.rect = rect
		}
		info.targetID = eid
		info.hitbox = target
		acc.hits[target] = info
		if _, seen := acc.contactedSet[target]; !seen {
			acc.contactedSet[target] = struct{}{}
			*acc.contactedList = append(*acc.contactedList, target)
		}
		if contactType > current {
			current = contactType
		}
	}

	return current
}

func maxContactWithMap(world *ecs.World, rect bump.Rect, current components.ContactType) components.ContactType {
	collision := physics.GetCollisionSpace(world)
	if collision == nil {
		return current
	}
	for _, col := range collision.Query(rect, nil, "map") {
		if collision.Has(col.Other, "slope") {
			continue
		}
		if components.Block > current {
			current = components.Block
		}
		break
	}
	return current
}

func enqueueHitEvents(world *ecs.World, attacker entities.EntityId, attackRect bump.Rect, hits []resolvedHit, damage float64) {
	if len(hits) == 0 {
		return
	}

	queue := ecs.Resource[resources.EventQueue](world)
	if queue == nil {
		queue = resources.NewEventQueue()
		world.SetResource(queue)
	}

	// Also record hitbox events for debug visualization (red fills on damaged entities)
	hitboxQueue := ecs.Resource[resources.HitboxEventQueue](world)

	for _, hit := range hits {
		queue.PushHit(resources.HitEvent{
			Attacker:   attacker,
			Target:     hit.targetID,
			Damage:     damage,
			Contact:    int(hit.contact),
			AttackRect: attackRect,
			TargetRect: hit.rect,
		})

		// Record victim's hurtbox for debug overlay (shows where damage was taken)
		if hitboxQueue != nil {
			hitboxQueue.Record(hit.rect.X, hit.rect.Y, hit.rect.W, hit.rect.H, "damage")
		}
	}
}

func ownerRect(world *ecs.World, eid entities.EntityId) (bump.Rect, bool) {
	// Get rect from entity's Transform component
	if transform := ecs.GetComponent[components.Transform](world, eid); transform != nil {
		return bump.Rect{X: transform.X, Y: transform.Y, W: transform.W, H: transform.H}, true
	}
	return bump.Rect{}, false
}
