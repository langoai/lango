## ADDED Requirements

### Requirement: Executor wrapper injection point
The application wiring SHALL support injecting a `turnrunner.Executor` wrapper between agent creation and TurnRunner construction. In structured mode, `initAgentRuntime()` SHALL wrap the inner executor with `CoordinatingExecutor`. In classic mode, the inner executor SHALL be used directly.

#### Scenario: Structured mode injects wrapper
- **WHEN** `agent.orchestration.mode` is `"structured"` and `app.New()` completes Phase B
- **THEN** the TurnRunner SHALL receive a `CoordinatingExecutor` as its executor

#### Scenario: LocalChat mode supports structured
- **WHEN** app is created with `WithLocalChat()` and `agent.orchestration.mode` is `"structured"`
- **THEN** `CoordinatingExecutor` SHALL still wrap the executor (it is not a lifecycle component, so `SetMaxPriority` does not affect it)

### Requirement: RetentionCleaner lifecycle registration
The `RetentionCleaner` SHALL be registered as a lifecycle component at `PriorityCore` when `observability.traceStore` config is present. It SHALL be started during `registry.StartAll()` and stopped during `registry.StopAll()`.

#### Scenario: Cleaner starts with app
- **WHEN** `observability.traceStore.maxAge` is configured
- **THEN** `RetentionCleaner` SHALL be registered at `PriorityCore` and started with the app

#### Scenario: Cleaner stops gracefully
- **WHEN** app shutdown is initiated
- **THEN** `RetentionCleaner` SHALL stop its ticker goroutine within the context timeout
