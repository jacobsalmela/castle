package game

import (
	"game/components"
	"game/ecs"
	"game/pkg/config"
	"game/pkg/tilemap"
	"game/systems"
	"log"
)

// ╔════════════════════════════════════════════════════════════════════════════════╗
// ║                     EBITENGINE FRAME STEP 2: UPDATE                            ║
// ║                                                                                ║
// ║  Update() is called by Ebitengine to run game logic.                           ║
// ║  This is called at TPS rate (60 times per second by default).                  ║
// ║                                                                                ║
// ║  Flow:                                                                         ║
// ║  Game.Update()                                                                 ║
// ║    └─ systems.Update()           (systems/update.go)                           ║
// ║                                                                                ║
// ║  Game.Update() is a thin orchestrator - all logic lives in systems/update/*    ║
// ║                                                                                ║
// ║  Next step: Draw() in game_draw.go                                             ║
// ╚════════════════════════════════════════════════════════════════════════════════╝

func (g *Game) Update() error {
	// Guard against nil world during reset transitions
	if g.world == nil {
		return nil
	}

	// Delegate all update logic to systems package (passing world and callbacks explicitly)
	return systems.Update(g.world, g.Reset, g.Save)
}

// loadEntitiesECSOnly loads entities from the Tiled map into the ECS world only.
func loadEntitiesECSOnly(worldMap *tilemap.Map, world *ecs.World) {
	if worldMap == nil || world == nil {
		return
	}

	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		return
	}

	tilemap.VisitEntityObjects(worldMap, "entities", func(obj tilemap.EntityObject) {
		// Skip Player: spawned and managed separately; avoid noisy warnings.
		if obj.TileID == cfg.Entities.Player || obj.Name == "Player" {
			return
		}

		// Skip marker objects (no GID/TileID assigned in Tiled)
		// These are spawn points, regions, etc. - not actual entities
		if obj.TileID == 0 {
			return
		}

		// Look up ECS constructor by TileID
		ctor := GetEntityConstructor(obj.TileID)
		if ctor == nil {
			log.Printf("Warning: ECS load: TileID %d (%s) not bound in ecsEntityBinds, skipping", obj.TileID, obj.Name)
			return
		}

		// Call the ECS prefab constructor directly - returns entities.EntityId
		eid := ctor(world, obj.X, obj.Y, obj.W, obj.H, obj.Props)
		if eid == 0 {
			log.Printf("Warning: ECS load: constructor for %s (ID %d) returned zero EntityId, skipping", obj.Name, obj.TileID)
			return
		}

		// Get the Transform component to adjust position
		// Tiled objects have origin at bottom-left, but our entities use top-left
		transform := ecs.GetComponent[components.Transform](world, eid)
		if transform != nil {
			// Adjust Y position: obj.Y is bottom of Tiled object, need to move to top
			transform.Y = obj.Y + obj.H - transform.H
		}

		// Register entity with its Tiled object ID for save/load system
		world.QueueInitWithID(eid, obj.ID)
	})
}
