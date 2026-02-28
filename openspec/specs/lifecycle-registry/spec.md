## Purpose

Component lifecycle management with priority-ordered startup, reverse-order shutdown, and automatic rollback on failure.

## Requirements

### Requirement: Component lifecycle interface
The system SHALL provide a `Component` interface with `Name()`, `Start(ctx, wg)`, and `Stop(ctx)` methods for managing application component lifecycles.

#### Scenario: Component implements interface
- **WHEN** a struct implements Name(), Start(context.Context, *sync.WaitGroup) error, and Stop(context.Context) error
- **THEN** it SHALL be usable as a lifecycle Component

### Requirement: Priority-ordered startup
The Registry SHALL start components in ascending priority order (lower number = earlier start).

#### Scenario: Components with different priorities start in order
- **WHEN** components are registered at PriorityInfra(100), PriorityBuffer(300), PriorityNetwork(400)
- **THEN** they SHALL start in order: Infra, Buffer, Network

#### Scenario: Same-priority preserves registration order
- **WHEN** multiple components are registered at the same priority
- **THEN** they SHALL start in the order they were registered (stable sort)

### Requirement: Reverse-order shutdown
The Registry SHALL stop started components in reverse startup order.

#### Scenario: Reverse stop order
- **WHEN** StopAll is called after A, B, C started in that order
- **THEN** they SHALL stop in order: C, B, A

### Requirement: Rollback on startup failure
If a component fails to start, the Registry SHALL stop all already-started components in reverse order.

#### Scenario: Third component fails to start
- **WHEN** A and B start successfully, then C fails
- **THEN** B and A SHALL be stopped in that order, and StartAll SHALL return C's error

### Requirement: Component adapters
The system SHALL provide adapters for common component signatures: SimpleComponent (Start(wg)/Stop()), FuncComponent (arbitrary functions), and ErrorComponent (Start(ctx) error/Stop()).

#### Scenario: SimpleComponent wraps buffer-style components
- **WHEN** a buffer with Start(*sync.WaitGroup) and Stop() is wrapped in SimpleComponent
- **THEN** it SHALL be usable as a lifecycle Component
