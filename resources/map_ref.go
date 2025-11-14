package resources

import "game/pkg/tilemap"

// MapRef is an ECS resource that holds a reference to the current tilemap.
// This replaces the global currentMap variable with an idiomatic ECS pattern.
type MapRef struct {
	Map *tilemap.Map
}
