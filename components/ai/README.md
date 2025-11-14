# AI Components - Architecture

## Overview

This directory contains AI components and the behavior tree framework for the Chronotome game engine.

## Pure Data Components

These are ECS-compliant components that hold only data (no logic):

- **`AI`** - Main AI component that holds:
  - `TargetID`: Reference to the entity being targeted
  - `BehaviorTree`: Reference to the behavior tree (framework object)
  - `Blackboard`: Key-value store for AI state sharing between nodes

- **`ApproachBehavior`** - Configuration for movement toward target:
  - Speed parameters
  - Range settings
  - Movement adjustments

- **`BackupBehavior`** - Configuration for retreat behavior:
  - Backward movement speed
  - Maximum retreat distance

- **`MeleeAttackBehavior`** - Attack parameters:
  - Damage values
  - Push/react forces
  - Animation and hitbox references
  - Cooldown timing

## Behavior Tree Framework

The behavior tree nodes (Sequence, Selector, Action, etc.) are an **intentional exception** to pure ECS principles. This is a deliberate architectural decision for the following reasons:

### Why Behavior Trees Have Methods

#### 1. Industry-Standard Pattern
Behavior trees are a proven AI pattern used across the game industry. They work as composable objects with methods, not as pure data structures.

#### 2. Stateless ECS Integration
The nodes don't store game state - they only store execution state (current child index, elapsed time). All game state comes from ECS components via the `world` parameter.

#### 3. Composability
Nodes compose into complex behaviors declaratively:

```go
behaviorTree := &Selector{
    Children: []Node{
        &Sequence{  // Attack if in range
            Children: []Node{
                &Condition{Check: isInAttackRange},
                &Action{OnTick: performAttack},
            },
        },
        &Action{OnTick: approachTarget},  // Otherwise approach
    },
}
```

#### 4. System Integration
The AI system (`systems/update/decision/ai.go`) orchestrates execution while keeping the ECS architecture clean:

```go
func TickBehaviorTree(world *ecs.World, eid entities.EntityId, ai *components.AI, dt float64) {
    ai.BehaviorTree.Tick(world, eid, dt)
}
```

#### 5. Domain-Specific Language
Behavior trees are a **DSL for AI**, not traditional components. The AI component itself remains pure data - it just holds a reference to the tree.

## Node Types

### Composites (Multi-child nodes)
- **`Sequence`** - Execute children in order until one fails or all succeed
- **`Selector`** - Execute children in order until one succeeds or all fail
- **`Parallel`** - Execute all children simultaneously with configurable success/failure policies

### Decorators (Single-child wrappers)
- **`Inverter`** - Inverts Success/Failure results
- **`Repeat`** - Repeats child N times or infinitely
- **`UntilFail`** - Repeats child until it fails
- **`UntilSuccess`** - Repeats child until it succeeds
- **`Timeout`** - Fails child if it takes too long
- **`Succeeder`** - Always returns Success
- **`Failer`** - Always returns Failure

### Leaves (Action/condition nodes)
- **`Action`** - Executes a function and returns its status
- **`Condition`** - Evaluates a boolean function (Success/Failure)
- **`AlwaysSuccess`** - Always returns Success
- **`AlwaysFailure`** - Always returns Failure
- **`AlwaysRunning`** - Always returns Running

## Debug Support

In non-release builds, the framework includes debug visualization:

- **`DebugNode`** - Wraps nodes to track execution statistics
- **`DebugNodeInfo`** - Stores node name, status, tick count, depth
- **`WrapForDebug()`** - Recursively wraps a tree for debugging
- **`CollectDebugInfo()`** - Gathers debug info from all nodes

Debug features are automatically disabled in release builds via build tags.

## Usage

See `systems/update/decision/ai.go` for how behavior trees integrate with the ECS architecture.

For examples of behavior tree construction, see the enemy prefabs in `prefabs/`.

## Architecture Philosophy

**Think of behavior trees as a domain-specific language embedded in your game engine.**

The AI component is pure ECS data. The behavior tree is a framework that operates on that data. This separation keeps the benefits of ECS (data-driven, decoupled systems) while leveraging the power of behavior trees for AI logic.

For a detailed analysis of this architecture decision, see `ai-components-analysis.md`.
