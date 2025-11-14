package config

// Lighting holds all lighting system configuration including light source types.
type Lighting struct {
	Enabled           bool                   `yaml:"enabled" line_comment:"Enable dynamic lighting shader (toggle with L key)"`
	AmbientBrightness float32                `yaml:"ambient_brightness" line_comment:"Minimum light level 0.0-1.0 (0=pitch black, 0.5=upstream, 1.0=fully lit)"`
	FlickerAmount     float32                `yaml:"flicker_amount" line_comment:"Global flicker intensity 0.0-1.0 (0=steady, 0.1=subtle, 0.4=candle)"`
	FlickerSpeed      float32                `yaml:"flicker_speed" line_comment:"Global torch sway speed in Hz (1.5=slow, 4.0=medium, 10.0=fast)"`
	Resolution        float32                `yaml:"resolution" line_comment:"Posterization bands (4=heavy, 8=upstream, 32=smooth)"`
	ResolutionOffset  float32                `yaml:"resolution_offset" line_comment:"Band positioning (0.8=upstream, 1.9=tuned)"`
	DitherIntensity   float32                `yaml:"dither_intensity" line_comment:"Grain texture (0=off, 0.01=subtle, 0.04=upstream)"`
	Sources           map[string]LightSource `yaml:"sources" head_comment:"Light source type definitions"`
}

// LightSource defines a type of light with specific properties.
type LightSource struct {
	Radius    float64 `yaml:"radius" line_comment:"Light emission radius in pixels"`
	Intensity float64 `yaml:"intensity" line_comment:"Brightness multiplier (1.0=normal, 2.0=twice as bright)"`
}

var (
	defaultLightingEnabled           = false // Lighting enabled by default
	defaultLightingAmbientBrightness = 0.8   // Bright ambient level (0.0=pitch black, 1.0=fully lit)
	defaultLightingFlickerAmount     = 0.01  // Very subtle flicker
	defaultLightingFlickerSpeed      = 2.0   // Slow torch sway
	defaultLightingResolution        = 7.0   // Slightly more posterization bands
	defaultLightingResolutionOffset  = 1.8   // Band positioning for desired look
	defaultLightingDitherIntensity   = 0.09  // Strong grain texture
)

// NewDefaultLighting returns Lighting configuration with sensible defaults.
func NewDefaultLighting() Lighting {
	return Lighting{
		Enabled:           defaultLightingEnabled,
		AmbientBrightness: float32(defaultLightingAmbientBrightness),
		FlickerAmount:     float32(defaultLightingFlickerAmount),
		FlickerSpeed:      float32(defaultLightingFlickerSpeed),
		Resolution:        float32(defaultLightingResolution),
		ResolutionOffset:  float32(defaultLightingResolutionOffset),
		DitherIntensity:   float32(defaultLightingDitherIntensity),
		Sources: map[string]LightSource{
			"torch": {
				Radius:    64.0, // Moderate reach
				Intensity: 1.0,  // Normal brightness
			},
			"torch_dim": {
				Radius:    32.0, // Shorter reach
				Intensity: 0.6,  // Dimmer
			},
			"torch_bright": {
				Radius:    128.0, // Very long reach
				Intensity: 3.5,   // Very bright
			},
			"brazier": {
				Radius:    96.0, // Large reach
				Intensity: 2.0,  // Very bright
			},
		},
	}
}

// GetLightSource retrieves a light source by name, returning default torch if not found.
func (l *Lighting) GetLightSource(name string) LightSource {
	if source, ok := l.Sources[name]; ok {
		return source
	}
	// Fallback to default torch
	return LightSource{
		Radius:    64.0,
		Intensity: 1.0,
	}
}
