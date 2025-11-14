package config

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Cfg holds the active configuration. Initialize to sensible defaults so
// other packages can safely read configuration values during package
// initialization. The real config may be loaded and overwrite this in main.
//
// DEPRECATED: Direct access to global config creates tight coupling and makes testing difficult.
// For runtime code, prefer accessing config via ECS resource:
//
//	cfg := ecs.Resource[config.Config](world)
//
// This global is still used by:
// - Initialization code (before ECS world exists) - acceptable temporary usage
// - Package-level variables (assets, lighting) - needs lazy initialization refactor
// - Draw systems - planned for future refactoring
// - Debug logging in animation helpers (37 call sites) - low priority
var Cfg = NewDefaultConfig()

// Config holds all game settings that can be serialized to YAML.
type Config struct {
	Screen       Screen       `yaml:"screen" head_comment:"Screen settings"`
	Lighting     Lighting     `yaml:"lighting" head_comment:"Lighting system settings"`
	Camera       Camera       `yaml:"camera" head_comment:"Camera settings"`
	Hud          Hud          `yaml:"hud" line_comment:"HUD settings"`
	Textbox      Textbox      `yaml:"textbox" line_comment:"Textbox settings"`
	World        World        `yaml:"world" head_comment:"In-game world settings"`
	Actor        Actor        `yaml:"actor" head_comment:"Actor settings"`
	Stats        Stats        `yaml:"stats" head_comment:"Stats settings"`
	Body         Body         `yaml:"body" head_comment:"Body settings"`
	Input        Input        `yaml:"input" head_comment:"Input key bindings"`
	Entities     Entities     `yaml:"entities" head_comment:"Entity GID mappings from Tiled tileset"`
	EnemyBalance EnemyBalance `yaml:"enemy_balance" head_comment:"Enemy balance configuration"`
	Debug        bool         `yaml:"debug" head_comment:"Enable debug mode" line_comment:"true or false"`
	DebugConsole bool         `yaml:"debug_console" line_comment:"Enable verbose console logging"`
}

type World struct {
	StartingMap string `yaml:"starting_map" line_comment:"Path to initial map file (relative to maps directory)"`
}

// NewDefaultConfig returns a Config populated with default values.
func NewDefaultConfig() *Config {
	return &Config{
		Screen: Screen{
			Width:                 screenWidth,
			Height:                screenHeight,
			Lighting:              lighting,
			LightRadius:           defaultLightRadius,
			AmbientBrightness:     float32(defaultAmbientLight),
			LightFlickerAmount:    float32(defaultFlickerAmt),
			LightFlickerSpeed:     float32(defaultFlickerSpeed),
			LightResolution:       float32(defaultLightResolution),
			LightResolutionOffset: float32(defaultLightResOffset),
			LightDitherIntensity:  float32(defaultLightDitherIntens),
			HighDpi:               true,
			DpiScale:              0, // 0 = auto-detect
			BackgroundColor:       defaultBackgroundColor,
			scale:                 scale,
		},
		Lighting: NewDefaultLighting(),
		Camera: Camera{
			TransitionDuration: defaultTransitionDuration,
			DamperStrength:     defaultDamperStrength,
		},
		Textbox: Textbox{
			BoxX:       defaultBoxX,
			BoxY:       defaultBoxY,
			BoxMarginY: defaultBoxMarginY,
			BoxMinY:    defaultBoxMinY,
			BoxMaxY:    defaultBoxMaxY,
			BoxW:       defaultBoxW,
			BoxH:       defaultBoxH,
			LineWidth:  defaultLineWidth,
			LineHeight: defaultLineHeight,
			MaxLines:   defaultMaxLines,
		},
		Hud: Hud{
			HudIconsX:    defaultHudIconsX,
			BarEndX1:     defaultBarEndX1,
			BarEndX2:     defaultBarEndX2,
			BarH:         defaultBarH,
			BarMiddleH:   defaultBarMiddleH,
			MiddleBarX1:  defaultMiddleBarX1,
			MiddleBarX2:  defaultMiddleBarX2,
			InnerBarH:    defaultInnerBarH,
			EnemyBarW:    defaultEnemyBarW,
			MaxTextWidth: defaultMaxTextWidth,
		},
		Body: Body{
			Gravity:            defaultGravity,
			MaxX:               defaultMaxX,
			MaxY:               defaultMaxY,
			GroundFriction:     defaultGroundFriction,
			AirFriction:        defaultAirFriction,
			CollisionStiffness: defaultCollisionStiffness,
			FrictionEpsilon:    defaultFrictionEpsilon,
			CoyoteTimeSeconds:  defaultCoyoteTimeSeconds,
		},
		World: World{
			StartingMap: "intro/playground_imp.tmx",
		},
		Actor: Actor{
			AttackPushForce:    defaultAttackPushForce,
			ReactForce:         defaultReactForce,
			MaxXDiv:            defaultMaxXDiv,
			MaxXRecoverRateDiv: defaultMaxXRecoverRateDiv,
		},
		Stats: Stats{
			Health:                defaultHealth,
			Stamina:               defaultStamina,
			Poise:                 defaultPoise,
			Heal:                  defaultHeal,
			HealAmount:            defaultHealAmount,
			AttackMultPerHeal:     defaultAttackMultPerHeal,
			RecoverRate:           defaultRecoverRate,
			RecoverSeconds:        defaultRecoverSeconds,
			HeadHealthShowSeconds: defaultHeadHealthShowSeconds,
			PlayerInvulnerable:    defaultPlayerInvulnerable,
		},
		Input:        NewDefaultInput(),
		EnemyBalance: NewDefaultEnemyBalance(),
		Entities:     NewDefaultEntities(),
		Debug:        true,
	}
}

// LoadConfig reads a YAML config from the given path.
// If the file doesn't exist or fails to parse, returns defaults and the error.
func LoadConfig(path string) (cfg *Config, err error) {
	cfg = NewDefaultConfig()
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			// Missing file: will save defaults below.
			if saveErr := cfg.Save("config.yml"); saveErr != nil {
				return cfg, fmt.Errorf("creating default config file: %w", saveErr)
			}
		} else {
			return cfg, readErr
		}
	}
	if unmarshalErr := yaml.Unmarshal(data, cfg); unmarshalErr != nil {
		return cfg, unmarshalErr
	}
	return cfg, nil
}

// Save writes the Config as YAML to the given path.
func (c *Config) Save(path string) error {
	// Open (or create) the file for writing:
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot open config file for writing: %w", err)
	}
	defer f.Close()

	// Write the YAML (with comments) into that file:
	if err := c.save(f); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Save writes the Config to w as YAML with comments.
func (c *Config) save(w io.Writer) error {
	node, err := c.ToYAMLNode(c)
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(node)
}

// ToYAMLNode builds a yaml.Node tree with comments from struct tags.
func (c *Config) ToYAMLNode(value interface{}) (*yaml.Node, error) {
	v := reflect.ValueOf(value)
	// unwrap pointer
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		return c.ToYAMLNode(v.Elem().Interface())
	}

	t := v.Type()
	switch t.Kind() {
	case reflect.Struct:
		return c.structNode(v)
	case reflect.Map:
		// build a map node
		mapNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		iter := v.MapRange()
		for iter.Next() {
			// key node
			keyNode, err := c.ToYAMLNode(iter.Key().Interface())
			if err != nil {
				return nil, err
			}
			// value node
			valNode, err := c.ToYAMLNode(iter.Value().Interface())
			if err != nil {
				return nil, err
			}
			mapNode.Content = append(mapNode.Content, keyNode, valNode)
		}
		return mapNode, nil
	case reflect.Slice, reflect.Array:
		// build a sequence node
		seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for i := 0; i < v.Len(); i++ {
			elemNode, err := c.ToYAMLNode(v.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			seq.Content = append(seq.Content, elemNode)
		}
		return seq, nil
	default:
		// scalar types: preserve numeric and boolean types
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: strconv.FormatInt(v.Int(), 10),
			}, nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: strconv.FormatUint(v.Uint(), 10),
			}, nil
		case reflect.Float32, reflect.Float64:
			return &yaml.Node{
				Kind: yaml.ScalarNode,
				// Tag:   "!!float",
				Value: strconv.FormatFloat(v.Float(), 'g', -1, 64),
			}, nil
		case reflect.Bool:
			return &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!bool",
				Value: strconv.FormatBool(v.Bool()),
			}, nil
		default:
			// fallback to string
			return &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: fmt.Sprint(value),
			}, nil
		}
	}
}

func (c *Config) structNode(v reflect.Value) (*yaml.Node, error) {
	t := v.Type()
	mapNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		// skip unexported
		if !v.Field(i).CanInterface() {
			continue
		}
		// yaml key name
		tag := field.Tag.Get("yaml")
		parts := strings.Split(tag, ",")
		keyName := parts[0]
		if keyName == "" {
			keyName = strings.ToLower(field.Name)
		}

		// key node (with comments)
		keyNode := &yaml.Node{
			Kind:        yaml.ScalarNode,
			Tag:         "!!str",
			Value:       keyName,
			HeadComment: field.Tag.Get("head_comment"),
			LineComment: field.Tag.Get("line_comment"),
			FootComment: field.Tag.Get("foot_comment"),
		}

		// value node (possibly nested)
		val, err := c.ToYAMLNode(v.Field(i).Interface())
		if err != nil {
			return nil, err
		}

		mapNode.Content = append(mapNode.Content, keyNode, val)
	}

	return mapNode, nil
}
