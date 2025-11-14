package ai

// BackupBehavior defines how an entity retreats from its target.
// Example: Ghoul backs up when player gets too close
type BackupBehavior struct {
	Speed    float64 // Movement acceleration away from target (pixels/sec²)
	MaxRange float64 // Backup until beyond this distance from target (pixels)
}
