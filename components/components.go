// Package components provides all game component types organized into logical subdirectories.
// Components are pure data structures with zero methods, following the Pure ECS pattern.
//
// This file re-exports all component types from their subdirectories, allowing external
// packages to import "game/components" and access types like components.Health, components.Transform, etc.
//
// Organization:
//   - markers/    : Core markers and state flags (Team, Facing, DeathState, etc.)
//   - enemies/    : Enemy-specific behavior components (Bat, Rat, Knight, etc.)
//   - ui/         : User interface components (HUD, Textbox, Headbar, etc.)
//   - combat/     : Combat-related components (Health, Stamina, Hitbox, etc.)
//   - spatial/    : Position and physics components (Transform, Physics, Collider, etc.)
//   - visual/     : Rendering and visual effects (Render, Animation, VisualEffects, etc.)
//   - ai/         : AI behavior components (AI, ApproachBehavior, BackupBehavior, etc.)
//   - world/      : World interactive components (Door, Chest, Grave, etc.)
//   - player/     : Player-specific components (Player, Input, ActorActions)
//   - debug/      : Debug visualization components (DebugOverlay, DebugStats, etc.)
//   - vfx/        : Visual effects and particles (Projectile, etc.)
package components

import (
	"game/components/ai"
	"game/components/combat"
	"game/components/debug"
	"game/components/enemies"
	"game/components/markers"
	"game/components/player"
	"game/components/spatial"
	"game/components/ui"
	"game/components/vfx"
	"game/components/visual"
	"game/components/world"
)

// MARKERS - Core markers and state flags
type Team = markers.Team
type TeamType = markers.TeamType

const (
	TeamPlayer  = markers.TeamPlayer
	TeamEnemy   = markers.TeamEnemy
	TeamNeutral = markers.TeamNeutral
)

type Facing = markers.Facing
type DeathState = markers.DeathState
type Stagger = markers.Stagger
type PauseState = markers.PauseState

// ENEMIES - Enemy-specific behavior components
type Bat = enemies.Bat
type Rat = enemies.Rat
type Crawler = enemies.Crawler
type Ent = enemies.Ent
type Ghoul = enemies.Ghoul
type Knight = enemies.Knight
type Skeleman = enemies.Skeleman
type Oscar = enemies.Oscar
type Gram = enemies.Gram
type Ferragus = enemies.Ferragus
type Acedian = enemies.Acedian

// UI - User interface components
type HUDData = ui.HUDData
type Textbox = ui.Textbox
type TextboxData = ui.TextboxData
type HeadbarData = ui.HeadbarData
type ViewPort = ui.ViewPort
type DamageNumber = ui.DamageNumber

// COMBAT - Combat-related components
type Health = combat.Health
type Poise = combat.Poise
type Stamina = combat.Stamina
type Hitbox = combat.Hitbox
type HitboxRect = combat.HitboxRect
type ContactType = combat.ContactType

const (
	Hit        = combat.Hit
	Block      = combat.Block
	ParryBlock = combat.ParryBlock
)

type AttackIntent = combat.AttackIntent
type AttackActive = combat.AttackActive
type AttackModifier = combat.AttackModifier
type AttackMultiplier = combat.AttackMultiplier
type Healing = combat.Healing
type Experience = combat.Experience
type HeadHealthTimer = combat.HeadHealthTimer

// SPATIAL - Position and physics components
type Transform = spatial.Transform
type Physics = spatial.Physics
type Collider = spatial.Collider
type DetectionRange = spatial.DetectionRange

// VISUAL - Rendering and visual effects
type Render = visual.Render
type Animation = visual.Animation
type AnimationStateEffect = visual.AnimationStateEffect
type AnimationSliceCallback = visual.AnimationSliceCallback
type UberColor = visual.UberColor
type VisualEffects = visual.VisualEffects
type Flake = visual.Flake
type Smoke = visual.Smoke
type Debris = visual.Debris

// AI - AI behavior components
type AI = ai.AI
type ApproachBehavior = ai.ApproachBehavior
type BackupBehavior = ai.BackupBehavior
type MeleeAttackBehavior = ai.MeleeAttackBehavior

// WORLD - World interactive components
type Door = world.Door
type StartDoor = world.StartDoor
type Chest = world.Chest
type Grave = world.Grave
type Spike = world.Spike
type FakeWall = world.FakeWall
type Object = world.Object

// PLAYER - Player-specific components
type Player = player.Player
type Input = player.Input
type InputKey = player.InputKey

const (
	InputKeyRight  = player.InputKeyRight
	InputKeyLeft   = player.InputKeyLeft
	InputKeyUp     = player.InputKeyUp
	InputKeyDown   = player.InputKeyDown
	InputKeyJump   = player.InputKeyJump
	InputKeyAction = player.InputKeyAction
	InputKeyGuard  = player.InputKeyGuard
	InputKeyHeal   = player.InputKeyHeal
	InputKeyDash   = player.InputKeyDash
)

type ActionIntents = player.ActionIntents

// DEBUG - Debug visualization components
type DebugOverlay = debug.DebugOverlay
type DebugTextLabel = debug.DebugTextLabel
type DebugVisual = debug.DebugVisual
type DebugVisualType = debug.DebugVisualType

const (
	DebugVisualRect   = debug.DebugVisualRect
	DebugVisualLine   = debug.DebugVisualLine
	DebugVisualCircle = debug.DebugVisualCircle
	DebugVisualText   = debug.DebugVisualText
)

// VFX - Visual effects and particles
type Projectile = vfx.Projectile
type LightEmitter = vfx.LightEmitter

// HELPER FUNCTIONS (Re-exported from subpackages)
var NewPhysics = spatial.NewPhysics
var NewPhysicsStatic = spatial.NewPhysicsStatic

// ANIMATION CONSTANTS (Re-exported from visual package)

// Animation State Tags
const (
	IdleTag       = visual.IdleTag
	WalkTag       = visual.WalkTag
	AttackTag     = visual.AttackTag
	BlockTag      = visual.BlockTag
	ParryBlockTag = visual.ParryBlockTag
	StaggerTag    = visual.StaggerTag
	ClimbTag      = visual.ClimbTag
	ConsumeTag    = visual.ConsumeTag
)

// Animation Slice Names
const (
	HurtboxSliceName = visual.HurtboxSliceName
	HitboxSliceName  = visual.HitboxSliceName
	BlockSliceName   = visual.BlockSliceName
	HealSliceName    = visual.HealSliceName
)

// Animation rendering constants
var (
	WhiteScalerColor     = visual.WhiteScalerColor
	FillNormalMaskColorM = visual.FillNormalMaskColorM
	NormalMaskColor      = visual.NormalMaskColor
)

// DEBUG FUNCTIONS (Re-exported from debug package)
var (
	SetDebugRect          = debug.SetDebugRect
	SetDebugRectWithLabel = debug.SetDebugRectWithLabel
	SetDebugRectWithColor = debug.SetDebugRectWithColor
)

// TYPE ALIASES FOR COMPATIBILITY
type Vec2 = spatial.Vec2
