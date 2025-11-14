package config

type Stats struct {
	Health                float64 `yaml:"defaultHealth" line_comment:"Default health points for an actor"`
	Stamina               float64 `yaml:"defaultStamina" line_comment:"Default stamina points for an actor"`
	Poise                 float64 `yaml:"defaultPoise" line_comment:"Default poise points for an actor"`
	Heal                  int     `yaml:"defaultHeal" line_comment:"Default number of heals an actor can perform"`
	HealAmount            float64 `yaml:"defaultHealAmount" line_comment:"Default amount of health restored per heal"`
	AttackMultPerHeal     float64 `yaml:"attackMultPerHeal" line_comment:"Attack multiplier per heal used"`
	RecoverRate           float64 `yaml:"defaultRecoverRate" line_comment:"Default stamina recovery rate per second"`
	RecoverSeconds        float64 `yaml:"defaultRecoverSeconds" line_comment:"Seconds after last stamina use before recovery starts"`
	HeadHealthShowSeconds float64 `yaml:"headHealthShowSeconds" line_comment:"Seconds to show health above head after taking damage"`
	PlayerInvulnerable    bool    `yaml:"playerInvulnerable" line_comment:"Make player invulnerable to damage (debug/testing)"`
}

// Stats configuration defaults
var (
	defaultHealth                = 1200.0 // was 100
	defaultStamina               = 80.0
	defaultPoise                 = 30.0
	defaultHeal                  = 2
	defaultHealAmount            = 20.0
	defaultAttackMultPerHeal     = 0.2
	defaultRecoverRate           = 26.0
	defaultRecoverSeconds        = 3.0
	defaultHeadHealthShowSeconds = 60.0
	defaultPlayerInvulnerable    = false
)
