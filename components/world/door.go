package world

// Door is a Pure ECS component for interactive door entities.
type Door struct {
	Opened         bool    // Whether door is currently open
	OpensFromRight bool    // Direction check for opening (left vs right)
	Height         float64 // Door height in pixels (needed for image compositing)
	NeedsInit      bool    // Whether door needs initialization by system (Pure ECS)
}
