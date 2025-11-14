# Castle Game - Developer Onboarding & Simplification Guide

## Quick Overview

**Type:** 2D Action RPG built with Ebitengine (Go game engine)
**Pattern:** Pure Entity Component System (ECS)

---

## Recommended Reading Order

1. Start: `game/game.go` - Entry point
2. Core: `ecs/ecs.go` - ECS implementation
3. Organization: `components/components.go`
4. Orchestration: `systems/update.go` → `systems/update/tick/tick.go`
5. Example System: `systems/update/physics/physics.go`
6. Prefabs: `prefabs/player.go` - Entity creation
7. Detail: Specific system files based on interest

## Key Files to Know

### Game Loop Entry Points
- `game/game.go` - Game struct & NewGame()
- `game/game_update.go` - Update() → systems.Update()
- `game/game_draw.go` - Draw() → systems.Draw()

### ECS Core
- `ecs/ecs.go` - World, entity management, component storage (635 lines)
  - Key methods: `NewEntity()`, `AddComponent[T]()`, `GetComponent[T]()`
  - Query optimization: `EntitiesWith()` uses type index

### Component Central Hub
- `components/components.go` - Re-exports all 87 component types

### System Orchestrators
- `systems/update.go` - Routes to pretick → tick → posttick
- `systems/update/tick/tick.go` - Main 10-phase ECS loop
- `systems/draw/draw.go` - 7-phase rendering pipeline

---

## High-Level Game Flow

### Entry Point: `cmd/main.go`

```
STARTUP SEQUENCE
================

1. CONFIGURATION [cmd/main.go]
   └── Load config.yml with hot-reload watcher

2. ASSET INITIALIZATION [game.InitAssets]
   ├── Load fonts (bitmap/pixel)
   ├── Auto-discover sprites from assets/images/
   ├── Load Tiled maps
   └── Validate entity GID bindings

3. GAME INSTANCE [game.NewGame]
   ├── Load save data
   ├── Create ECS World with resources
   ├── Load starting map + collision geometry
   ├── Spawn entities from Tiled "entities" layer
   ├── Apply save state (player position, stats, opened doors)
   └── Setup camera tracking

4. EBITENGINE LOOP [ebiten.RunGameWithOptions]
   └── Hands control to engine for Layout/Update/Draw cycle
```

### Runtime Game Loop (per frame)

```
┌─────────────────────────────────────────────────────────┐
│  LayoutF()  [game/game_layout.go]                       │
│  ├── Calculate viewport scaling                         │
│  └── Handle DPI for pixel art                           │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  Update()   [systems/update.go]                         │
│                                                         │
│  PRETICK                                                │
│  ├── Time control (game speed, freeze)                  │
│  ├── Input polling (keyboard → Input component)         │
│  └── Map tile animation                                 │
│                                                         │
│  TICK (10 phases in systems/update/tick/tick.go)        │
│  ├── 1. Preupdate     - Time scaling                    │
│  ├── 2. Initialization - Deferred entity setup          │
│  ├── 3. Decision      - AI + player input → intents     │
│  ├── 4. Physics       - Movement + collision            │
│  ├── 5. Combat        - Damage resolution               │
│  ├── 6. State         - Animation + status updates      │
│  ├── 7. Entities      - Interactive objects             │
│  ├── 8. VFX           - Particle effects                │
│  ├── 9. UI            - HUD/textbox updates             │
│  └── 10. Cleanup      - Entity removal + transitions    │
│                                                         │
│  POSTTICK                                               │
│  ├── Lighting updates                                   │
│  └── State finalization                                 │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  Draw()     [systems/draw/draw.go]                      │
│                                                         │
│  7-PHASE RENDER PIPELINE                                │
│  ├── 1. Buffers     - Create logical screen + normal map│
│  ├── 2. World       - Map tiles + Z-sorted sprites      │
│  ├── 3. Lighting    - Apply dynamic lighting shader     │
│  ├── 4. UI          - HUD, headbars, textboxes          │
│  ├── 5. Postprocess - Transitions, screen effects       │
│  ├── 6. Viewport    - Scale to device resolution        │
│  └── 7. Debug       - Hitboxes, collision visualization │
└─────────────────────────────────────────────────────────┘
```

---

## Architecture Overview

### Directory Structure

```
castle/
├── cmd/main.go              # Entry point
├── game/                    # Ebitengine adapter (thin orchestrator)
│   ├── game.go              # Game struct, NewGame()
│   ├── game_update.go       # Update() → systems.Update()
│   ├── game_draw.go         # Draw() → systems.Draw()
│   ├── game_layout.go       # LayoutF() viewport scaling
│   ├── world.go             # InitializeWorld() 8-phase bootstrap
│   └── game_vars.go         # Entity GID → constructor bindings
│
├── ecs/                     # ECS core implementation
│   └── ecs.go               # World, components, resources (635 lines)
│
├── components/              # 87 pure data components
│   ├── spatial/             # Transform, Physics, Collider
│   ├── combat/              # Health, Stamina, Hitbox, AttackIntent
│   ├── ai/                  # AI, BehaviorTree nodes
│   ├── visual/              # Animation, Render
│   └── ...                  # enemies/, markers/, player/, ui/, world/
│
├── systems/                 # All game logic
│   ├── update.go            # Update orchestrator
│   ├── update/tick/tick.go  # 10-phase update pipeline
│   ├── update/decision/     # AI systems (42 files)
│   ├── update/physics/      # Movement + collision (13 files)
│   ├── update/combat/       # Damage resolution (13 files)
│   └── draw/                # 7-phase render pipeline (37 files)
│
├── prefabs/                 # Entity constructors (32 files)
│   ├── player.go            # NewPlayerPrefab()
│   ├── enemy_factory.go     # Shared enemy setup
│   └── *.go                 # Per-entity type constructors
│
├── resources/               # Singleton systems (14 files)
│   ├── camera.go            # Viewport + room management
│   ├── time_control.go      # Game speed + freeze
│   └── events.go            # Event queue system
│
└── pkg/                     # Utilities
    ├── config/              # Game configuration
    ├── tilemap/             # Tiled map loading
    └── bump/                # Collision library
```

### ECS Pattern

**Components** are pure data structures:
```go
type Health struct {
    Current float64
    Max     float64
    Lag     float64  // for smooth bar animation
}
```

**Systems** query and modify components:
```go
func UpdatePhysics(world *ecs.World, dt float64) {
    entities := world.EntitiesWith((*Transform)(nil), (*Physics)(nil))
    for _, eid := range entities {
        transform := ecs.GetComponent[Transform](world, eid)
        physics := ecs.GetComponent[Physics](world, eid)
        transform.Y += physics.Velocity.Y * dt
    }
}
```

**Resources** are singletons:
```go
camera := ecs.Resource[resources.Camera](world)
config := ecs.Resource[config.Config](world)
```

---

## Major Subsystems at a Glance

### Physics System 
- Location: `systems/update/physics/`
- Uses: Transform, Physics, Collider
- Features: Gravity, friction, collision detection, grounding, coyote time

### Combat System
- Location: `systems/update/combat/`
- Uses: Health, Stamina, Poise, Hitbox, AttackIntent
- Features: Damage, parry blocking, knockback, stagger

### AI System
- Location: `systems/update/decision/`
- Uses: AI, BehaviorTree, ApproachBehavior
- Features: Behavior tree-based AI for 6 simple enemies + 4 bosses

### Animation System
- Location: `systems/update/entities/animation/`
- Uses: Animation, Render
- Features: Frame callbacks, state-driven animation

### Rendering System
- Location: `systems/draw/`
- Features: Sprite composition, lighting, UI, debug overlays

---

## Data Flow Example: Player Attack

```
1. INPUT      [preupdate/input.go]
   └── ebiten.IsKeyPressed() → Input.KeyPressed[Attack] = true

2. DECISION   [decision/player.go]
   └── Input component → AttackIntent component

3. PHYSICS    [physics/]
   └── Apply movement from ActionIntents

4. COMBAT     [combat/attack_intent_consume.go]
   └── AttackIntent → AttackActive
   └── Check hitbox overlaps with bump.Space
   └── Apply damage to target Health

5. STATE      [state/]
   └── Update Stamina (consume attack cost)
   └── Set animation state to "attack"

6. ANIMATION  [entities/animation/]
   └── Advance frames, trigger attack callbacks

7. RENDER     [draw/world/]
   └── Draw current animation frame
   └── Draw damage numbers
```

---

## Recommended Reading Order

For new developers, read these files in order:

1. **Entry point:** `cmd/main.go` (87 lines)
2. **Game adapter:** `game/game.go` (thin orchestrator)
3. **ECS core:** `ecs/ecs.go` (635 lines, understand World/components/resources)
4. **Update pipeline:** `systems/update/tick/tick.go` (10 phases)
5. **Render pipeline:** `systems/draw/draw.go` (7 phases)
6. **Entity creation:** `prefabs/player.go` (see full prefab pattern)
7. **Example system:** `systems/update/physics/physics.go`


