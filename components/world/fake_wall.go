package world

// FakeWall is an interactive entity representing a destructible wall tile.
// Fake walls appear as normal wall tiles but crumble when hit by the player,
// triggering a chain reaction to adjacent fake walls.
type FakeWall struct {
	Opened              bool    // Whether fake wall has been crumbled
	DestroyTimer        float64 // Time until entity destruction (0 = not scheduled)
	NeighborTriggerTime float64 // Time until neighbor chain reaction triggers (0 = not scheduled)
}
