package physics

import (
	"game/components"
	"game/ecs"
	"game/entities"
	"game/pkg/config"
)

// getConfig retrieves config from ECS resources with fallback to default.
// This replaces the global config.Cfg pattern with resource-based access.
func getConfig(world *ecs.World) *config.Config {
	cfg := ecs.Resource[config.Config](world)
	if cfg == nil {
		cfg = config.NewDefaultConfig()
		world.SetResource(cfg)
	}
	return cfg
}

// IsProjectileSpawning checks if a projectile is in spawn grace period.
// Prevents physics from running on newly created projectiles before initialization completes.
func IsProjectileSpawning(world *ecs.World, eid entities.EntityId) bool {
	projectile := ecs.GetComponent[components.Projectile](world, eid)
	return projectile != nil && projectile.SpawnGrace > 0
}
