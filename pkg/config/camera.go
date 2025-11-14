package config

type Camera struct {
	TransitionDuration float64 `yaml:"transitionDuration" line_comment:"Duration of room transition animation in seconds" head_comment:"0.8 = 800 milliseconds (current default, ~48 frames at 60fps) 1.2 = 1200 milliseconds (slower, ~72 frames)`
	DamperStrength     float64 `yaml:"damperStrength" line_comment:"Camera follow smoothing (0=instant, higher=smoother)"`
}

// Camera configuration defaults
var (
	defaultTransitionDuration = 0.8 // Current default in camera.go
	defaultDamperStrength     = 0.1 // Current default in camera.go
)
