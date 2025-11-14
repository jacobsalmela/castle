package world

import "time"

// StartDoor is a temporary visual barrier at spawn points that self-destructs.
// The lifecycle is handled by systems/update/entities/start_door_lifecycle.go.
type StartDoor struct {
	SpawnTime    time.Time // When door was spawned
	DestroyDelay float64   // Delay before destruction in seconds
}
