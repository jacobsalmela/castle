package config

type Actor struct {
	AttackPushForce    float64 `yaml:"attackPushForce" line_comment:"Force applied to an enemy when attacked"`
	ReactForce         float64 `yaml:"reactForce" line_comment:"Force applied to the actor when reacting to an attack"`
	MaxXDiv            int     `yaml:"maxXDiv" line_comment:"Divisor for maximum horizontal speed"`
	MaxXRecoverRateDiv int     `yaml:"maxXRecoverRateDiv" line_comment:"Divisor for maximum horizontal speed during stamina recovery"`
}

// Actor configuration defaults
var (
	defaultAttackPushForce    = 100.0
	defaultReactForce         = 50.0
	defaultMaxXDiv            = 2
	defaultMaxXRecoverRateDiv = 3
)
