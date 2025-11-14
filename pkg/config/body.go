package config

type Body struct {
	Gravity            float64 `yaml:"gravity" line_comment:"Gravity affecting the body"`
	MaxX               float64 `yaml:"maxX" line_comment:"Maximum horizontal speed" yaml2:"maxX" line_comment2:"Maximum horizontal speed"`
	MaxY               float64 `yaml:"maxY" line_comment:"Maximum vertical speed" yaml2:"maxY" line_comment2:"Maximum vertical speed"`
	GroundFriction     float64 `yaml:"groundFriction" line_comment:"Friction when on the ground"`
	AirFriction        float64 `yaml:"airFriction" line_comment:"Friction when in the air"`
	CollisionStiffness float64 `yaml:"collisionStiffness" line_comment:"Stiffness of collisions with the environment"`
	FrictionEpsilon    float64 `yaml:"frictionEpsilon" line_comment:"Minimum speed before friction is applied"`
	CoyoteTimeSeconds  float64 `yaml:"coyoteTimeSeconds" line_comment:"Time after leaving a platform when a jump is still allowed"`
}

// Body configuration defaults
var (
	defaultGravity            = 300.0
	defaultMaxX               = 20.0
	defaultMaxY               = 200.0
	defaultGroundFriction     = 8.0
	defaultAirFriction        = 2.0 // Tuned empirically during gameplay
	defaultCollisionStiffness = 1.0
	defaultFrictionEpsilon    = 0.05
	defaultCoyoteTimeSeconds  = 0.1
)
