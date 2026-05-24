## MODIFIED Requirements

### Requirement: Public BuildHookRegistry helper
The `internal/app` package SHALL export a `BuildHookRegistry(cfg *config.Config, bus *eventbus.Bus, knowledgeSaver toolchain.KnowledgeSaver, catalog *toolcatalog.Catalog) *toolchain.HookRegistry` function that produces the same hook registry as the runtime app builder. When `bus` is nil, EventBus hooks are omitted. When `knowledgeSaver` is nil, `KnowledgeSaveHook` is still registered (for snapshot inspection) but its `Post` method safely no-ops. When `catalog` is non-nil, `SaveableTools` is derived from catalog; otherwise falls back to `DefaultSaveableTools`. The private `buildHookRegistry` function SHALL delegate to this public helper, passing the runtime `KnowledgeSaver` from the knowledge subsystem.

#### Scenario: CLI uses BuildHookRegistry without full bootstrap
- **WHEN** the `agent hooks` CLI command loads config and calls `BuildHookRegistry(cfg, nil, nil, nil)`
- **THEN** the returned registry contains all config-derivable hooks (SecurityFilter, AccessControl, KnowledgeSaveHook)
- **AND** no database connection, crypto initialization, or event bus is required

#### Scenario: Runtime path provides KnowledgeSaver
- **WHEN** the app builder calls `buildHookRegistry` during full bootstrap
- **THEN** the `KnowledgeSaver` from the knowledge subsystem is passed through to `KnowledgeSaveHook`
- **AND** tool results for saveable tools are persisted to the knowledge store
