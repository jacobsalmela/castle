# Behavior Tree Nodes

This directory contains factory functions that create behavior tree nodes for enemy AI. These nodes are used by the decision system and provide game-specific behaviors for conditions, actions, and composite nodes.

## Architecture Notes

### Component Mutations

Behavior tree action nodes directly mutate components (`Physics.Velocity`, `Facing.FlipX`, `Health.Current`, etc.). This is intentional and safe because:

1. **Execution Order**: The decision system runs BEFORE Physics/Combat systems
   - Decision phase sets intents (velocity changes, attack triggers)
   - Physics phase processes movement
   - Combat phase processes damage

2. **Single-Threaded**: Behavior trees execute sequentially per entity
   - No concurrent access to components
   - No race conditions

3. **Alternative Considered**: Intent components (`MoveIntent`, `AttackIntent`)
   - Rejected: Adds complexity without clear benefit
   - Current approach is simpler and performs better

### Thread Safety

**Closure-Captured State**: Some actions use closure variables (e.g., `elapsed` in `Wait()` from utility.go)
- ✅ Safe: Single-threaded execution
- ⚠️ Warning: If parallel behavior trees are added, use component-based state instead

### Code Organization

The node files are organized by function:

- **conditions.go** - Condition nodes for checking entity state (health, range, animation, etc.)
- **combat.go** - Combat action nodes (attacks, blocking, damage application)
- **movement.go** - Movement action nodes (approach, backup, face target, jump)
- **utility.go** - Utility action nodes (wait, idle, callbacks, animation)
- **choice.go** - Weighted random selection composites (RandomSelector, WeightedSequence)
- **helpers/** - Shared utility functions used by nodes
  - **transform.go** - Transform calculations (distance, direction, center)
  - **weighted.go** - Weighted random selection algorithms

### Helper Functions

The `helpers/` package provides reusable utility functions to reduce code duplication:

#### Transform Helpers

- `GetCenter(t *Transform) (x, y float64)` - Returns the center point of a transform
- `CalculateDistance(t1, t2 *Transform) float64` - Returns Euclidean distance between two transforms
- `CalculateDirection(t1, t2 *Transform) (dx, dy float64)` - Returns direction vector from t1 to t2
- `IsTargetOnRight(t1, t2 *Transform) bool` - Returns true if t2 is to the right of t1

#### Weighted Selection Helpers

- `SelectWeightedIndex(weights []float64) int` - Returns a random index based on weights
- `ShuffleWeightedIndices(weights []float64) []int` - Returns indices shuffled by weight priority

These helpers eliminate duplicate distance/center calculations that previously appeared 10+ times across the codebase.

### Future Considerations

If parallel system execution is added:
1. Replace closure state with components (e.g., `WaitTimer` component)
2. Consider intent components for deferred mutations
3. Add mutex protection for shared state

## Node Types

### Conditions

Conditions check entity state and return true/false:

- `HasTarget()` - Check if AI has valid target
- `IsInRange(minDist, maxDist)` - Distance-based condition
- `IsGrounded()` / `IsNotGrounded()` - Physics state checks
- `HasHealth(minHealth)` / `HealthBelow(threshold)` - Health checks
- `IsAnimationState(stateName)` - Animation state check
- `HasStamina(minStamina)` - Stamina check
- `BlackboardCheck(key)` / `BlackboardExists(key)` - Blackboard queries

### Actions

Actions perform behaviors and return Success/Running/Failure:

#### Combat Actions
- `PlayAnimation(animationState)` - Play animation and wait for completion
- `PlayAnimationOnce(animationState)` - Fire-and-forget animation
- `MeleeAttack(attackAnimationName)` - Face target, attack, wait for animation
- `Block(blockAnimationName, duration)` - Blocking state with timer
- `ApplyDamageToTarget(damage)` - Direct damage application
- `SetBlackboardValue(key, value)` / `ClearBlackboardValue(key)` - Blackboard mutations

#### Movement Actions
- `ApproachTarget()` - Move toward target using ApproachBehavior component
- `BackupFromTarget()` - Move away from target using BackupBehavior component
- `FaceTarget()` - Orient toward target
- `MoveToPosition(blackboardKey, speed, threshold)` - Move to specific position from blackboard
- `Jump(jumpForce)` - Apply jump force
- `StopMovement()` - Zero velocity

#### Utility Actions
- `Wait(duration)` - Timed delay
- `Idle()` - No-op
- `SetAnimationState(stateName)` - Set animation without waiting
- `RunCallback(callback)` - Execute custom logic
- `RunCallbackEveryFrame(duration, callback)` - Repeated callback
- `Fail()` / `Succeed()` - Testing/forcing nodes
- `RandomSuccess(probability)` - Probabilistic outcome

### Composites

Composite nodes control behavior tree flow:

- `RandomSelector` - Randomly selects ONE child based on weights, executes only that child
- `WeightedSequence` - Shuffles children by weight, then executes all in sequence

These use the weighted selection helpers for consistent random behavior.

## Usage Example

```go
// Create a simple enemy AI behavior tree
tree := &ai.Sequence{
    Children: []ai.Node{
        nodes.HasTarget(),           // Condition: check for target
        nodes.IsInRange(0, 50),      // Condition: within attack range
        nodes.FaceTarget(),          // Action: face the target
        nodes.MeleeAttack("Attack"), // Action: perform attack
    },
}

// Add to AI component
aiComp := &components.AI{
    BehaviorTree: tree,
    TargetID: playerEid,
}
world.AddComponent(enemyEid, aiComp)
```

## Testing

Unit tests are provided for helper functions:
- `helpers/transform_test.go` - Tests for transform calculations
- `helpers/weighted_test.go` - Tests for weighted selection (including statistical distribution tests)

Run tests with:
```bash
go test ./systems/update/decision/nodes/helpers/...
```
