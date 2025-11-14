package spatial

// DetectionRange defines vision/detection area for finding targets.
// The detection rect is built relative to the entity's Transform position
// and Facing direction. Distances are measured in pixels from entity bounds.
type DetectionRange struct {
	// Directional distances from entity center (in pixels)
	FrontDistance float64 // Distance in facing direction
	BackDistance  float64 // Distance behind entity (0 = no back vision)
	UpDistance    float64 // Vertical distance above entity
	DownDistance  float64 // Vertical distance below entity

	// Detection behavior flags
	RequiresLineOfSight bool   // Whether detection blocked by walls (future feature)
	DetectsStealthed    bool   // Whether entity detects stealth (future feature)
	TeamFilter          string // Which teams to detect: "player", "enemy", "all"
}
