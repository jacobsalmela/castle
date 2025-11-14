package config

type Textbox struct {
	BoxX       float64 `yaml:"boxX" line_comment:"X position of the textbox"`
	BoxY       float64 `yaml:"boxY" line_comment:"Y position of the textbox"`
	BoxMarginY float64 `yaml:"boxMarginY" line_comment:"Vertical margin of the textbox"`
	BoxMinY    float64 `yaml:"boxMinY" line_comment:"Minimum Y position of the textbox"`
	BoxMaxY    float64 `yaml:"boxMaxY" line_comment:"Maximum Y position of the textbox"`
	BoxW       float64 `yaml:"boxW" line_comment:"Width of the textbox"`
	BoxH       float64 `yaml:"boxH" line_comment:"Height of the textbox"`
	LineWidth  float64 `yaml:"lineWidth" line_comment:"Width of each line in the textbox"`
	LineHeight float64 `yaml:"lineHeight" line_comment:"Height of each line in the textbox"`
	MaxLines   float64 `yaml:"maxLines" line_comment:"Maximum number of lines in the textbox"`
	// Font settings
	FontSize          int    `yaml:"fontSize" line_comment:"Font size in device pixels for textbox text"`
	IndicatorFontSize int    `yaml:"indicatorFontSize" line_comment:"Font size in device pixels for dismissed indicator"`
	FontName          string `yaml:"fontName" line_comment:"Font name (reserved for future use)"`

	// Spacing
	PaddingX float64 `yaml:"paddingX" line_comment:"Horizontal padding inside textbox"`
	PaddingY float64 `yaml:"paddingY" line_comment:"Vertical padding inside textbox"`

	// Colors
	BackgroundColor ColorConfig `yaml:"backgroundColor" line_comment:"Background color of textbox (RGBA)"`
	TextColor       ColorConfig `yaml:"textColor" line_comment:"Text color (RGBA)"`

	// Indicator settings
	IndicatorText  string      `yaml:"indicatorText" line_comment:"Character(s) shown when textbox is dismissed"`
	IndicatorColor ColorConfig `yaml:"indicatorColor" line_comment:"Color of dismissed indicator (RGBA)"`
}

// Textbox configuration defaults
var (
	defaultBoxX       = 6.0
	defaultBoxY       = 3.0
	defaultBoxMarginY = 5.0
	defaultBoxMaxY    = 25.0
	defaultBoxMinY    = 96.0 - defaultBoxH - defaultBoxMarginY
	defaultBoxW       = screenWidth - defaultBoxX*2
	defaultBoxH       = 3.0
	defaultLineWidth  = (defaultBoxW - 8)
	defaultLineHeight = 6.0 + 1.0
	defaultMaxLines   = 4.0

	// Font defaults
	defaultFontSize          = 12
	defaultIndicatorFontSize = 36
	defaultFontName          = "default"

	// Spacing defaults
	defaultPaddingX = 8.0
	defaultPaddingY = 4.0

	// Color defaults
	defaultBackgroundColorTextbox = ColorConfig{R: 168, G: 167, B: 168, A: 255}
	defaultTextColor              = ColorConfig{R: 255, G: 255, B: 255, A: 255}

	// Indicator defaults
	defaultIndicatorText  = "i"
	defaultIndicatorColor = ColorConfig{R: 255, G: 255, B: 255, A: 255}
)
