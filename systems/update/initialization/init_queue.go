package initialization

import "game/entities"

type initQueueState interface {
	DrainInitQueue() []entities.EntityId
	RegisterActive(entities.EntityId) bool
}

func RunInitQueue(state initQueueState) {
	if state == nil {
		return
	}

	entityIDs := compactEntityIDs(state.DrainInitQueue())
	if len(entityIDs) == 0 {
		return
	}

	registerActiveEntities(state, entityIDs)
}

func compactEntityIDs(raw []entities.EntityId) []entities.EntityId {
	if len(raw) == 0 {
		return nil
	}
	filtered := raw[:0]
	for _, eid := range raw {
		if eid == 0 {
			continue
		}
		filtered = append(filtered, eid)
	}
	return filtered
}

func registerActiveEntities(state initQueueState, entityIDs []entities.EntityId) {
	for _, eid := range entityIDs {
		state.RegisterActive(eid)
	}
}
