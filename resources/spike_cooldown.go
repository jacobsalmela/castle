package resources

import "game/entities"

// SpikeCooldown tracks per-entity cooldowns for spike damage.
// This is a global resource to prevent spam damage from spikes while maintaining
// deterministic behavior (uses world time, not system time).
type SpikeCooldown struct {
	// Map entity ID to last damage time (in world time)
	Cooldowns map[entities.EntityId]float64

	// Cooldown duration in seconds
	Duration float64
}

// NewSpikeCooldown creates a new spike cooldown tracker with default settings.
func NewSpikeCooldown() *SpikeCooldown {
	return &SpikeCooldown{
		Cooldowns: make(map[entities.EntityId]float64),
		Duration:  2.0, // 2 seconds cooldown
	}
}

// CanDamage checks if enough time has passed to damage entity again.
func (sc *SpikeCooldown) CanDamage(eid entities.EntityId, currentTime float64) bool {
	if sc == nil {
		return true
	}
	lastTime, exists := sc.Cooldowns[eid]
	if !exists {
		return true
	}
	return currentTime-lastTime >= sc.Duration
}

// RecordDamage marks entity as damaged at current time.
func (sc *SpikeCooldown) RecordDamage(eid entities.EntityId, currentTime float64) {
	if sc == nil {
		return
	}
	sc.Cooldowns[eid] = currentTime
}

// CleanupExpired removes old cooldown entries to prevent map growth.
// Removes entries that are more than 2x the cooldown duration old.
func (sc *SpikeCooldown) CleanupExpired(currentTime float64) {
	if sc == nil {
		return
	}
	for eid, lastTime := range sc.Cooldowns {
		if currentTime-lastTime > sc.Duration*2 {
			delete(sc.Cooldowns, eid)
		}
	}
}
