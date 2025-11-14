package player

// ActionIntents captures per-frame control requests from inputs or AI systems.
// Each field represents an intent that can be held or released in a frame.
type ActionIntents struct {
	ShieldHeld    bool
	ShieldRelease bool
	ClimbHeld     bool
	ClimbDrop     bool
	ClimbRelease  bool
	Heal          bool
}
