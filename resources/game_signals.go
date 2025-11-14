package resources

// GameSignals is an ECS resource that holds game-level signal flags.
// This replaces global vars.SaveGame and vars.ResetGame with proper ECS resource management.
//
// Usage:
//   - Set resource: world.SetResource(&resources.GameSignals{})
//   - Request save: signals.RequestSave()
//   - Request reset: signals.RequestReset()
//   - Check and consume: if signals.ConsumeSave() { /* save game */ }
type GameSignals struct {
	saveRequested    bool
	resetRequested   bool
	pendingResetFunc func() // Deferred reset function to call after frame completes
}

// RequestSave signals that a game save should be performed.
func (g *GameSignals) RequestSave() {
	if g != nil {
		g.saveRequested = true
	}
}

// RequestReset signals that a game reset should be performed.
func (g *GameSignals) RequestReset() {
	if g != nil {
		g.resetRequested = true
	}
}

// ConsumeSave checks if a save was requested and clears the flag.
// Returns true if save was requested.
func (g *GameSignals) ConsumeSave() bool {
	if g == nil {
		return false
	}
	if g.saveRequested {
		g.saveRequested = false
		return true
	}
	return false
}

// ConsumeReset checks if a reset was requested and clears the flag.
// Returns true if reset was requested.
func (g *GameSignals) ConsumeReset() bool {
	if g == nil {
		return false
	}
	if g.resetRequested {
		g.resetRequested = false
		return true
	}
	return false
}

// IsSaveRequested checks if a save is pending without consuming the flag.
func (g *GameSignals) IsSaveRequested() bool {
	return g != nil && g.saveRequested
}

// IsResetRequested checks if a reset is pending without consuming the flag.
func (g *GameSignals) IsResetRequested() bool {
	return g != nil && g.resetRequested
}

// SetPendingReset stores a reset function to be called after the current frame.
// This prevents race conditions between Update and Draw when resetting mid-frame.
func (g *GameSignals) SetPendingReset(resetFunc func()) {
	if g != nil {
		g.pendingResetFunc = resetFunc
	}
}

// ConsumePendingReset executes and clears any pending reset function.
// Returns true if a reset was executed.
func (g *GameSignals) ConsumePendingReset() bool {
	if g == nil || g.pendingResetFunc == nil {
		return false
	}
	g.pendingResetFunc()
	g.pendingResetFunc = nil
	return true
}
