## ADDED Requirements

### Requirement: Phase B cleanup stack accumulates rollback functions
`app.New()` Phase B SHALL maintain a `cleanupStack` that accumulates named cleanup functions as components are registered. Each Phase B step that creates a stoppable resource MUST push a cleanup entry.

#### Scenario: OutputStore registered with cleanup
- **WHEN** Phase B step B4b creates and registers an OutputStore
- **THEN** a cleanup entry named "output-store" is pushed that calls `outputStore.Stop()`

#### Scenario: Gateway registered with cleanup
- **WHEN** Phase B step B5 creates a Gateway
- **THEN** a cleanup entry named "gateway" is pushed that calls `gateway.Shutdown()`

### Requirement: Phase B rollback executes cleanups in reverse order on failure
When a Phase B step fails, the cleanup stack SHALL execute all accumulated cleanups in LIFO (last-in, first-out) order.

#### Scenario: Agent creation failure triggers rollback
- **WHEN** Phase B step B6 (agent creation) fails
- **THEN** cleanup stack rolls back gateway first, then output-store

#### Scenario: Empty stack rollback does not panic
- **WHEN** rollback is called on an empty cleanup stack
- **THEN** no panic occurs and the stack remains empty

### Requirement: Phase B success discards cleanup stack
When Phase B completes successfully, the cleanup stack SHALL be cleared without executing any cleanups. The lifecycle registry takes ownership of all registered components.

#### Scenario: Successful initialization clears stack
- **WHEN** all Phase B steps complete without error
- **THEN** `cleanups.clear()` is called and no cleanup functions execute
