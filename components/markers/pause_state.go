package markers

// PauseState indicates whether an entity's update logic should pause.
// Used for hitstun, death, cutscenes, etc.
type PauseState struct {
	// Paused indicates if the entity should skip update logic this frame.
	Paused bool
}
