// Package bump provides spatial hashing and AABB collision detection.
//
// This is a collision detection library, NOT a physics engine.
// It answers: "If I move from A to B, what do I hit?"
//
// # Architecture
//
// The package is organized into focused files:
//   - bump.go: Core Space and public API
//   - types.go: Type definitions, constants, constructors
//   - geometry.go: Geometric utilities (rectangles, vectors, line intersection)
//   - collision.go: Collision detection logic
//   - response.go: Collision response functions
//
// # Quick Start
//
//	space := bump.NewSpace()
//	space.Set(player, bump.NewRect(0, 0, 16, 16))
//	goal := bump.Vec2{X: 100, Y: 100}
//	actualPos, cols := space.Move(player, goal, bump.DefaultResponseFilter)
//
// # Collision Response Types
//
//   - Touch: Stop at collision point
//   - Cross: Pass through (ghost collision)
//   - Slide: Slide along surfaces (handles slopes)
//   - RectSlide: Slide along rectangular surfaces only
//
// # Thread Safety
//
// All Space operations are thread-safe using RWMutex.
// Multiple readers can query simultaneously, writers have exclusive access.
//
// For more details, see: https://github.com/kikito/bump.lua
package bump
