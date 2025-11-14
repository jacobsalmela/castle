# Enemy Prefab Construction Guide

This guide explains the **minimum required components** to create a new enemy in Castle's Pure ECS architecture.

## Table of Contents

- [Quick Start Checklist](#quick-start-checklist)
- [Component Categories](#component-categories)
- [Minimum Required Components](#minimum-required-components)
- [Optional Enhancement Components](#optional-enhancement-components)
- [Code Template](#code-template)
- [Examples by Enemy Type](#examples-by-enemy-type)
- [Common Patterns](#common-patterns)

---

## Quick Start Checklist

✅ **Every enemy MUST have these 16 core components:**

1. `Transform` - Position and size
2. `Animation` - Visual rendering
3. `Hitbox` - Damage reception (empty, initialized by system)
4. `Health` - Hit points
5. `Poise` - Knockback resistance
6. `Experience` - XP reward on death
7. `HeadHealthTimer` - Healthbar visibility
8. `AI` - Behavior brain
9. `DetectionRange` - Target detection
10. `ApproachBehavior` - Movement toward target
11. `Physics` - Movement and forces
12. `Collider` - Collision detection
13. `Team` - Combat allegiance
14. `Facing` - Sprite direction
15. `VisualEffects` - Damage flash
16. `DeathState` - Death animation
17. `[EnemyType]` - Enemy-specific component (e.g., `Rat`, `Bat`, `Skeleman`)
18. `EnemyInit` - One-time initialization tracker

✅ **Ground enemies need:**
- `BackupBehavior` - Movement away from target

✅ **Melee enemies need:**
- `MeleeAttackBehavior` - Attack configuration

---

## Component Categories

### 🗺️ Spatial (1 component)
Position and dimensions in world space.

### 🎨 Visual (1 component)
Sprite rendering and animation.

### ⚔️ Combat (6 components)
Health, damage, death, and visual feedback.

### 🤖 AI & Detection (3 components)
Target detection and behavior logic.

### 🏃 Movement (2-3 components)
Physics, approach, and backup behaviors.

### 🔧 Configuration (4 components)
Team, facing, collision, and entity-specific state.

### ⚙️ Initialization (1 component)
One-time setup tracking.

---

## Minimum Required Components

### 1. Transform (Spatial)
**Purpose:** Entity position and collision dimensions  
**Location:** `components/transform.go`

```go
transform := &components.Transform{
    X: x,        // World X position
    Y: y,        // World Y position
    W: width,    // Collision width (pixels)
    H: height,   // Collision height (pixels)
}
world.AddComponent(eid, transform)
```

**Notes:**
- Use constants for width/height (e.g., `ratWidth = 11`)
- Position is center-based, not top-left
- Collision box, not sprite dimensions

---

### 2. Animation (Visual)
**Purpose:** Sprite rendering with animation  
**Location:** Use `NewAnimationComponent()` helper in `prefabs/animation.go`

```go
anim := NewAnimationComponent(AnimationConfig{
    FilesName:  "rat",           // Base animation filename
    OX:         -13,             // Sprite offset X (facing right)
    OY:         -17,             // Sprite offset Y (vertical)
    OXFlip:     21,              // Sprite offset X (facing left)
    Layer:      5,               // Render layer (5 = standard entity)
    FSMInitial: "Idle",          // Starting animation state
    FSMTransitions: map[string]string{
        "Attack": "Idle",        // Animation → next state
    },
})
world.AddComponent(eid, anim)
```

**Notes:**
- Animation files in `assets/` folder (e.g., `rat.json`, `rat.aseprite`)
- Layer 5 is standard for entities
- FSM transitions optional (only if enemy uses state machine)

---

### 3. Hitbox (Combat - Empty Initialization)
**Purpose:** Damage reception area  
**Location:** `components/hitbox.go`

```go
hitbox := &components.Hitbox{}  // Empty - system will initialize
world.AddComponent(eid, hitbox)
```

**Notes:**
- Leave empty - `systems/update/initialization/` will populate using `Facing`
- Initialized via `extractSlice()` for complex enemies or simple collision rect
- Must be present before initialization system runs

---

### 4. Health (Combat)
**Purpose:** Hit points and death tracking  
**Location:** `components/health.go`

```go
health := &components.Health{
    Current: maxHealth,  // Current HP
    Max:     maxHealth,  // Maximum HP
    Lag:     maxHealth,  // Lag HP (for smooth bar animation)
}
world.AddComponent(eid, health)
```

**Notes:**
- Define `maxHealth` as constant (e.g., `ratHealth = 25`)
- `Lag` should start equal to `Current` for new entities
- Use `health.IsDead()` method to check death state

---

### 5. Poise (Combat)
**Purpose:** Knockback resistance  
**Location:** `components/poise.go`

```go
poise := &components.Poise{
    Current: maxPoise,   // Current poise
    Max:     maxPoise,   // Maximum poise
    Lag:     maxPoise,   // Lag poise (for smooth bar animation)
}
world.AddComponent(eid, poise)
```

**Notes:**
- Higher poise = harder to stagger/knockback
- Small enemies: 10-15, Medium: 20-30, Large: 40+
- Regenerates over time in combat system

---

### 6. Experience (Combat)
**Purpose:** XP reward when defeated  
**Location:** `components/experience.go`

```go
experience := &components.Experience{Points: xpAmount}
world.AddComponent(eid, experience)
```

**Notes:**
- Define as constant (e.g., `ratExp = 15`)
- Awarded to player on enemy death via flake system
- Small enemies: 10-20, Medium: 20-40, Large: 50+

---

### 7. HeadHealthTimer (Combat)
**Purpose:** Controls overhead healthbar visibility  
**Location:** `components/head_health_timer.go`

```go
headTimer := &components.HeadHealthTimer{}
world.AddComponent(eid, headTimer)
```

**Notes:**
- Shows healthbar above enemy when damaged
- Timer duration configured in `config.Cfg.Stats.HeadHealthShowSeconds`
- System manages visibility automatically

---

### 8. AI (AI & Detection)
**Purpose:** Behavior brain and target tracking  
**Location:** `components/ai.go`

```go
brain := &components.AI{TargetID: 0}  // 0 = no target yet
world.AddComponent(eid, brain)
```

**Notes:**
- Managed by `systems/update/decision/ai_*.go` systems
- `TargetID` set by target detection system
- State machine managed through `ActionQueue` field

---

### 9. DetectionRange (AI & Detection)
**Purpose:** Target detection configuration  
**Location:** `components/detection_range.go`

```go
detectionRange := components.NewDetectionRange(
    60.0,     // FrontDistance: pixels ahead
    0.0,      // BackDistance: pixels behind (0 = no back vision)
    16.0,     // UpDistance: pixels above
    16.0,     // DownDistance: pixels below
    "player", // TeamFilter: which teams to detect
)
world.AddComponent(eid, detectionRange)
```

**Notes:**
- **Ground enemies:** Forward-focused (back=0, down=up)
- **Flying enemies:** Look down (down >> up, front=back)
- **Smart enemies:** 360° vision (front=back, up=down)
- Use entity dimensions as base: `1.5 * width`, `1.5 * height`

---

### 10. ApproachBehavior (Movement)
**Purpose:** Configuration for moving toward target  
**Location:** `components/approach_behavior.go`

```go
approachBehavior := components.NewApproachBehavior(
    60.0, // Speed: Movement acceleration
    40.0, // MaxSpeed: Max velocity when approaching
    20.0, // MinRange: Stop within this distance of target
    0.0,  // RangeAdjustment: Additional padding (usually 0)
)
world.AddComponent(eid, approachBehavior)
```

**Notes:**
- Used by `systems/update/decision/approach.go`
- Speed values: Slow=40-60, Medium=60-100, Fast=100+
- MinRange = attack range (usually 20-30 pixels)
- MaxSpeed should match `Physics.MaxVelocity.X`

---

### 11. Physics (Movement)
**Purpose:** Movement, forces, gravity  
**Location:** `components/physics.go`

```go
physics := spatial.NewPhysics()
physics.Weight = 0.6              // Body weight (0.5-1.0)
physics.MaxVelocity.X = 40.0      // Max horizontal speed
physics.GravityEnabled = true     // Ground enemy
physics.FrictionEnabled = true    // Apply ground friction
world.AddComponent(eid, physics)
```

**Notes:**
- **Ground enemies:** `GravityEnabled=true`, `FrictionEnabled=true`
- **Flying enemies:** Both false, set `MaxVelocity.Y` as well
- Weight affects knockback (lighter = more knockback)
- MaxVelocity.X should match ApproachBehavior.MaxSpeed

---

### 12. Collider (Movement)
**Purpose:** Collision detection and physics  
**Location:** `components/collider.go`

```go
collider := components.NewCollider()
collider.Tags = []string{"enemy", "body"}
collider.QueryTags = []string{"player", "body", "map", "solid"}
collider.Solid = true  // Blocks other solid entities
world.AddComponent(eid, collider)
```

**Notes:**
- `Tags`: How this entity is identified in collisions
- `QueryTags`: What this entity collides with
- `Solid=true`: Standard for enemies (can't overlap solid objects)
- Width/Height auto-populated from Transform

---

### 13. Team (Configuration)
**Purpose:** Combat allegiance  
**Location:** `components/team.go`

```go
world.AddComponent(eid, &components.Team{Type: components.TeamEnemy})
```

**Notes:**
- All enemies: `TeamEnemy`
- Player: `TeamPlayer`
- Used for friendly fire prevention and targeting

---

### 14. Facing (Configuration)
**Purpose:** Sprite orientation  
**Location:** `components/facing.go`

```go
facing := &components.Facing{FlipX: flipX}  // true = left, false = right
world.AddComponent(eid, facing)
```

**Notes:**
- Used by animation system for sprite flipping
- Used by hitbox initialization for directional collision
- AI systems update this based on movement direction

---

### 15. VisualEffects (Combat)
**Purpose:** Damage flash effect  
**Location:** `components/visual_effects.go`

```go
// White flash (default)
visualEffects := components.NewVisualEffects(0.8)
world.AddComponent(eid, visualEffects)

// Custom color flash
visualEffects := components.NewVisualEffectsWithColor(0.8, 222, 0, 0)  // Red
world.AddComponent(eid, visualEffects)
```

**Notes:**
- Duration: 0.8-1.5 seconds typical
- Triggered by damage in combat system
- Custom colors for special enemies (e.g., skeleman = red)

---

### 16. DeathState (Combat)
**Purpose:** Death animation and fade-out  
**Location:** `components/death_state.go`

```go
deathState := components.NewDeathState(1.0)  // 1 second fade
world.AddComponent(eid, deathState)
```

**Notes:**
- Duration: 1.0 seconds standard
- Managed by `systems/update/state/death.go`
- Entity removed after fade completes

---

### 17. Enemy-Specific Component (Configuration)
**Purpose:** Entity-type-specific state  
**Location:** `components/[enemy_type].go`

```go
// Example: Rat
rat := &components.Rat{
    RemovalTarget: eid,
    Paused:        false,
}
world.AddComponent(eid, rat)

// Example: Bat
bat := &components.Bat{
    RemovalTarget: eid,
    Paused:        false,
}
world.AddComponent(eid, bat)

// Example: Ghoul
ghoul := &components.Ghoul{
    Rocks:         0,
    Poacher:       false,
    ThrowCooldown: 0,
    Paused:        false,
}
world.AddComponent(eid, ghoul)
```

**Notes:**
- Every enemy type needs its own component
- Minimum: `RemovalTarget` (EntityId) and `Paused` (bool)
- Add enemy-specific fields as needed (e.g., Ghoul.Rocks)
- Used by enemy-specific AI systems

---

### 18. EnemyInit (Initialization)
**Purpose:** Tracks one-time initialization state  
**Location:** `components/enemy_init.go`

```go
enemyInit := components.NewEnemyInit(0.8)  // Flash duration
world.AddComponent(eid, enemyInit)
```

**Notes:**
- Prevents re-initialization on every frame
- Flash duration should match VisualEffects duration
- Removed by initialization system after setup completes

---

## Optional Enhancement Components

### BackupBehavior (Ground Enemies)
**Purpose:** Configuration for moving away from target  
**Required for:** Ground-based melee enemies

```go
backupBehavior := components.NewBackupBehavior(
    60.0, // Speed: Movement acceleration away
    30.0, // MaxRange: Backup until beyond this distance
)
world.AddComponent(eid, backupBehavior)
```

**When to use:**
- Ground enemies that need retreat behavior
- NOT needed for flying enemies (bat)
- Used by `systems/update/decision/backup.go`

---

### MeleeAttackBehavior (Melee Enemies)
**Purpose:** Melee attack configuration  
**Required for:** Enemies with melee attacks

```go
meleeAttackBehavior := components.NewMeleeAttackBehavior(
    18.0,     // Damage: Hit point damage
    10.0,     // PushForce: Knockback force on target
    10.0,     // ReactForce: Recoil force on attacker
    "Attack", // AnimationTag: Animation state name
)
world.AddComponent(eid, meleeAttackBehavior)
```

**When to use:**
- Enemies with melee attacks (skeleman, crawler, ghoul)
- NOT needed for ranged-only enemies
- Used by `systems/update/combat/attack_*.go` systems

---

## Code Template

```go
package prefabs

import (
    "game/components"
    "game/ecs"
    "game/entities"
)

const (
    // Visual properties
    myEnemyAnimFile = "myenemy"
    myEnemyWidth    = 12
    myEnemyHeight   = 10
    myEnemyOffsetX  = -10
    myEnemyOffsetY  = -8
    myEnemyOffsetFlip = 18

    // Physics properties
    myEnemyWeight = 0.7

    // Combat stats
    myEnemyHealth = 50
    myEnemyPoise  = 20
    myEnemyExp    = 25

    // Visual effect durations
    myEnemyFlashDuration = 0.8
    myEnemyDieDuration   = 1.0
)

// NewMyEnemyPrefab constructs a MyEnemy entity.
//
// [Describe enemy behavior here]
//
// Parameters:
//   - world: ECS world instance (required)
//   - x, y: Spawn position in world coordinates
//   - w, h: Unused (kept for tilemap compatibility)
//   - flipX: Initial facing direction (true = left, false = right)
//
// Returns: EntityId of the created entity, or 0 if world is nil
func NewMyEnemyPrefab(world *ecs.World, x, y, w, h float64, flipX bool) entities.EntityId {
    if world == nil {
        return 0
    }

    eid := world.NewEntity()

    // === SPATIAL COMPONENT ===
    transform := &components.Transform{X: x, Y: y, W: myEnemyWidth, H: myEnemyHeight}
    world.AddComponent(eid, transform)

    // === VISUAL COMPONENT ===
    anim := NewAnimationComponent(AnimationConfig{
        FilesName:  myEnemyAnimFile,
        OX:         myEnemyOffsetX,
        OY:         myEnemyOffsetY,
        OXFlip:     myEnemyOffsetFlip,
        Layer:      5,
        FSMInitial: "Idle",
        FSMTransitions: map[string]string{
            "Attack": "Idle",
        },
    })
    world.AddComponent(eid, anim)

    // === COMBAT COMPONENTS ===
    hitbox := &components.Hitbox{}
    world.AddComponent(eid, hitbox)

    health := &components.Health{Current: myEnemyHealth, Max: myEnemyHealth, Lag: myEnemyHealth}
    world.AddComponent(eid, health)

    poise := &components.Poise{Current: myEnemyPoise, Max: myEnemyPoise, Lag: myEnemyPoise}
    world.AddComponent(eid, poise)

    experience := &components.Experience{Points: myEnemyExp}
    world.AddComponent(eid, experience)

    headTimer := &components.HeadHealthTimer{}
    world.AddComponent(eid, headTimer)

    // === AI COMPONENTS ===
    brain := &components.AI{}
    world.AddComponent(eid, brain)

    detectionRange := components.NewDetectionRange(
        18.0,     // FrontDistance (1.5 * width)
        18.0,     // BackDistance (360° vision)
        15.0,     // UpDistance (1.5 * height)
        15.0,     // DownDistance
        "player",
    )
    world.AddComponent(eid, detectionRange)

    // === MOVEMENT COMPONENTS ===
    approachBehavior := components.NewApproachBehavior(80.0, 50.0, 20.0, 0.0)
    world.AddComponent(eid, approachBehavior)

    backupBehavior := components.NewBackupBehavior(80.0, 30.0)
    world.AddComponent(eid, backupBehavior)

    physics := spatial.NewPhysics()
    physics.Weight = myEnemyWeight
    physics.MaxVelocity.X = 50.0
    physics.GravityEnabled = true
    physics.FrictionEnabled = true
    world.AddComponent(eid, physics)

    collider := components.NewCollider()
    collider.Tags = []string{"enemy", "body"}
    collider.QueryTags = []string{"player", "body", "map", "solid"}
    collider.Solid = true
    world.AddComponent(eid, collider)

    // === CONFIGURATION COMPONENTS ===
    world.AddComponent(eid, &components.Team{Type: components.TeamEnemy})

    facing := &components.Facing{FlipX: flipX}
    world.AddComponent(eid, facing)

    visualEffects := components.NewVisualEffects(myEnemyFlashDuration)
    world.AddComponent(eid, visualEffects)

    deathState := components.NewDeathState(myEnemyDieDuration)
    world.AddComponent(eid, deathState)

    // === BEHAVIOR COMPONENT ===
    myEnemy := &components.MyEnemy{
        RemovalTarget: eid,
        Paused:        false,
    }
    world.AddComponent(eid, myEnemy)

    // === INITIALIZATION COMPONENT ===
    enemyInit := components.NewEnemyInit(myEnemyFlashDuration)
    world.AddComponent(eid, enemyInit)

    return eid
}
```

---

## Examples by Enemy Type

### Ground Melee Enemy (Rat, Crawler, Skeleman)
- ✅ All 18 core components
- ✅ BackupBehavior (retreat from player)
- ✅ MeleeAttackBehavior (attack config)
- ✅ Physics: Gravity + Friction enabled
- ✅ DetectionRange: Forward-focused or 360°

### Flying Enemy (Bat)
- ✅ All 18 core components
- ❌ No BackupBehavior (uses custom stalk behavior)
- ❌ No MeleeAttackBehavior (uses dive attack)
- ✅ Physics: Gravity + Friction disabled
- ✅ Physics: MaxVelocity.Y set for vertical movement
- ✅ DetectionRange: Looks down (DownDistance >> UpDistance)

### Ranged Enemy (Ghoul Poacher)
- ✅ All 18 core components
- ✅ BackupBehavior (kite player)
- ✅ MeleeAttackBehavior (fallback melee)
- ✅ Custom component fields (Rocks, Poacher, ThrowCooldown)
- ✅ Helper functions (SetGhoulRocks, SetGhoulPoacher)

---

## Common Patterns

### Detection Range Guidelines

**Ground Melee (Rat, Crawler):**
```go
detectionRange := components.NewDetectionRange(
    60.0,  // Front: 5-6x entity width
    0.0,   // Back: No back vision
    16.0,  // Up/Down: Similar (ground-focused)
    16.0,
    "player",
)
```

**Flying (Bat):**
```go
detectionRange := components.NewDetectionRange(
    7.5,   // Front: Small horizontal range
    7.5,   // Back: Equal (looks in all directions)
    0.0,   // Up: Doesn't look up
    80.0,  // Down: Large (detects ground targets)
    "player",
)
```

**Smart Melee (Skeleman, Ghoul):**
```go
detectionRange := components.NewDetectionRange(
    12.0,  // Front: 1.5x width
    12.0,  // Back: 360° vision
    18.0,  // Up: 1.5x height
    18.0,  // Down: Equal all directions
    "player",
)
```

---

### Physics Configuration

**Ground Enemy:**
```go
physics := spatial.NewPhysics()
physics.Weight = 0.6-0.9          // Lighter = more knockback
physics.MaxVelocity.X = 40-80     // Match ApproachBehavior.MaxSpeed
physics.GravityEnabled = true
physics.FrictionEnabled = true
```

**Flying Enemy:**
```go
physics := spatial.NewPhysics()
physics.GravityEnabled = false
physics.FrictionEnabled = false
physics.MaxVelocity.X = 40.0
physics.MaxVelocity.Y = 40.0      // Allow vertical movement
```

---

### Combat Stats

**Small Enemy (Rat):**
- Health: 25-50
- Poise: 10-15
- Experience: 10-20

**Medium Enemy (Crawler, Skeleman):**
- Health: 70-110
- Poise: 20-30
- Experience: 20-30

**Large Enemy (Boss):**
- Health: 200+
- Poise: 40+
- Experience: 50+

---

## Next Steps

1. **Create enemy-specific component** in `components/[enemy_type].go`
2. **Create prefab constructor** using template above
3. **Create AI behavior** in `systems/update/decision/ai_[enemy_type].go`
4. **Create animations** in Aseprite, export to `assets/`
5. **Test spawning** via tilemap or debug command
6. **Tune constants** (health, speed, detection) for balance

---

## Related Documentation

- **Pure ECS Pattern:** `docs/PURE_ECS_PATTERN.md`
- **System Phases:** See `.github/copilot-instructions.md` Phase Flow
- **Component Reference:** `components/README.md` (if exists)
- **Flake Example:** `prefabs/flake.go` (gold standard implementation)

---

## Questions?

If you need help:
1. Check existing enemy prefabs for reference (rat.go, bat.go, skeleman.go)
2. Review component files in `components/` to understand fields
3. Look at AI systems in `systems/update/decision/` for behavior patterns
4. Test with minimal setup first, add complexity incrementally
