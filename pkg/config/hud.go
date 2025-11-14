package config

type Hud struct {
	HudIconsX    int `yaml:"hudIconsX" line_comment:"X position of the HUD icons"`
	BarEndX1     int `yaml:"barEndX1" line_comment:"X position of the start of the health/stamina/poise bars"`
	BarEndX2     int `yaml:"barEndX2" line_comment:"X position of the end of the health/stamina/poise bars"`
	BarH         int `yaml:"barH" line_comment:"Height of the health/stamina/poise bars"`
	BarMiddleH   int `yaml:"barMiddleH" line_comment:"Height of the middle part of the health/stamina/poise bars"`
	MiddleBarX1  int `yaml:"middleBarX1" line_comment:"X position of the start of the middle part of the health/stamina/poise bars"`
	MiddleBarX2  int `yaml:"middleBarX2" line_comment:"X position of the end of the middle part of the health/stamina/poise bars"`
	InnerBarH    int `yaml:"innerBarH" line_comment:"Height of the inner part of the health/stamina/poise bars"`
	EnemyBarW    int `yaml:"enemyBarW" line_comment:"Width of the enemy health/stamina/poise bars"`
	MaxTextWidth int `yaml:"maxTextWidth" line_comment:"Maximum width of text in the HUD"`
}

// HUD configuration defaults
var (
	defaultHudIconsX    = 7
	defaultBarEndX1     = 8
	defaultBarEndX2     = 12
	defaultBarH         = 7
	defaultBarMiddleH   = defaultBarH - 2
	defaultMiddleBarX1  = 7
	defaultMiddleBarX2  = 8
	defaultInnerBarH    = 3
	defaultEnemyBarW    = 10
	defaultMaxTextWidth = 50
)
