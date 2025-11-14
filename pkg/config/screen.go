package config

import "image/color"

type Screen struct {
	Width                 float64     `yaml:"width" line_comment:"Width of the game screen in pixels"`
	Height                float64     `yaml:"height" line_comment:"Height of the game screen in pixels"`
	Lighting              bool        `yaml:"lighting" line_comment:"Enable dynamic lighting shader (toggle with L key)"`
	LightRadius           float64     `yaml:"light_radius" line_comment:"Light emission radius in pixels (16=upstream tight, 32=wider)"`
	AmbientBrightness     float32     `yaml:"ambient_brightness" line_comment:"Minimum light level 0.0-1.0 (0=pitch black, 0.5=upstream, 1.0=fully lit)"`
	LightFlickerAmount    float32     `yaml:"light_flicker_amount" line_comment:"Flicker intensity 0.0-1.0 (0=steady, 0.1=subtle, 0.4=candle)"`
	LightFlickerSpeed     float32     `yaml:"light_flicker_speed" line_comment:"Torch sway speed in Hz (1.5=slow, 4.0=medium, 10.0=fast)"`
	LightResolution       float32     `yaml:"light_resolution" line_comment:"Posterization bands (4=heavy, 8=upstream, 32=smooth)"`
	LightResolutionOffset float32     `yaml:"light_resolution_offset" line_comment:"Band positioning (0.8=upstream, 1.9=tuned)"`
	LightDitherIntensity  float32     `yaml:"light_dither_intensity" line_comment:"Grain texture (0=off, 0.01=subtle, 0.04=upstream)"`
	HighDpi               bool        `yaml:"high_dpi" line_comment:"Enable high-DPI rendering on supported displays"`
	DpiScale              float64     `yaml:"dpi_scale" line_comment:"Manual DPI scale override (0 = auto-detect, 2.0 = Retina)"`
	BackgroundColor       ColorConfig `yaml:"background_color" head_comment:"Background color (RGBA)"`
	scale                 float64
}

type ColorConfig struct {
	R uint8 `yaml:"r" line_comment:"Red component (0-255)"`
	G uint8 `yaml:"g" line_comment:"Green component (0-255)"`
	B uint8 `yaml:"b" line_comment:"Blue component (0-255)"`
	A uint8 `yaml:"a" line_comment:"Alpha component (0-255)"`
}

// ToColor converts ColorConfig to color.RGBA
func (c ColorConfig) ToColor() color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

var (
	screenWidth              = 160.0
	screenHeight             = 96.0
	scale                    = 1.0
	lighting                 = true // Lighting enabled by default
	defaultLightRadius       = 16.0 // Upstream default (tight torch radius)
	defaultAmbientLight      = 0.5  // Upstream ambient level (0.0=pitch black, 1.0=fully lit)
	defaultFlickerAmt        = 0.1  // Subtle flicker (0.4 = strong candle-like)
	defaultFlickerSpeed      = 4.0  // Medium-fast torch sway
	defaultLightResolution   = 8.0  // Upstream posterization bands
	defaultLightResOffset    = 1.9  // Band positioning for desired look
	defaultLightDitherIntens = 0.01 // Subtle grain texture

	// Default background color (dark teal)
	defaultBackgroundColor = ColorConfig{R: 50, G: 60, B: 57, A: 255}
)
