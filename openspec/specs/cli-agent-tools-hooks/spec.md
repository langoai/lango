# CLI Agent Tools & Hooks

## Purpose
Provides CLI commands for listing registered agent tools and displaying hook configuration with runtime registry snapshots.

## Requirements

### Requirement: Agent tools command
The system SHALL provide a `lango agent tools [--json]` command that lists all registered tools in the agent's tool catalog. The command SHALL use cfgLoader to load configuration and enumerate tools by name and description.

#### Scenario: List tools in text format
- **WHEN** user runs `lango agent tools`
- **THEN** system displays a table with NAME and DESCRIPTION columns for each registered tool

#### Scenario: List tools in JSON format
- **WHEN** user runs `lango agent tools --json`
- **THEN** system outputs a JSON array of tool objects with name and description fields

### Requirement: Agent hooks command
The system SHALL provide a `lango agent hooks [--json]` command that displays the current hook configuration including enabled hooks, blocked commands, and active hook types. The command SHALL use cfgLoader for configuration. Additionally, the command SHALL build a `HookRegistry` from the loaded config via a public helper (`BuildHookRegistry`) and display the registry snapshot: each registered pre/post hook's name, priority, and wirable status. For `KnowledgeSaveHook`, the output SHALL include the active `SaveableTools` list. Existing config-only output fields SHALL remain unchanged for backward compatibility.

#### Scenario: Hooks enabled
- **WHEN** user runs `lango agent hooks` with hooks.enabled set to true
- **THEN** system displays which hook types are active (securityFilter, accessControl, eventPublishing, knowledgeSave) and any blocked command patterns
- **AND** system displays a "Registered Hooks" section listing each pre-hook and post-hook with name and priority
- **AND** for KnowledgeSaveHook, the output includes the active saveable tools list

#### Scenario: Hooks disabled
- **WHEN** user runs `lango agent hooks` with hooks.enabled set to false
- **THEN** system displays "Hooks are disabled"
- **AND** system still displays the registry snapshot (SecurityFilter is always registered regardless of the enabled flag)

#### Scenario: Hooks in JSON format
- **WHEN** user runs `lango agent hooks --json`
- **THEN** system outputs a JSON object with fields: enabled, securityFilter, accessControl, eventPublishing, knowledgeSave, blockedCommands
- **AND** the JSON object includes a `registry` field containing `preHooks` and `postHooks` arrays, each entry having `name`, `priority`, and `wirable` fields
- **AND** hooks with extended details (e.g. KnowledgeSaveHook) include a `details` object with hook-specific information

#### Scenario: EventBus hook not wirable in CLI mode
- **WHEN** user runs `lango agent hooks` and eventPublishing is enabled in config
- **THEN** the registry output includes an EventBus placeholder entry with `wirable: false` and a `reason` field indicating it requires a running event bus
- **AND** the placeholder's `phase` is `pre+post` since EventBus registers in both phases at runtime

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

### Requirement: Default SaveableTools allowlist
The `toolchain` package SHALL define a `DefaultSaveableTools` constant as a fallback for CLI mode. At runtime, the `app.go` builder SHALL derive the saveable tools list from the tool catalog using `Catalog.SaveableToolNames()`, which filters by `ToolCapability.KnowledgeSaveable()`. The `BuildHookRegistry` function SHALL accept an optional catalog parameter and prefer catalog-derived tools when available.

#### Scenario: KnowledgeSaveHook uses catalog-derived list at runtime
- **WHEN** the app builder constructs `KnowledgeSaveHook` with a non-nil catalog
- **THEN** the hook's `SaveableTools` set equals the catalog's `SaveableToolNames()` result
- **AND** the list includes all tools where `ReadOnly == true` or `Activity ∈ {read, query}`

#### Scenario: KnowledgeSaveHook falls back to constant in CLI mode
- **WHEN** `BuildHookRegistry` is called with a nil catalog (CLI snapshot mode)
- **THEN** the hook's `SaveableTools` set equals `DefaultSaveableTools`

#### Scenario: CLI hooks output indicates source
- **WHEN** user runs `lango agent hooks`
- **THEN** the KnowledgeSaveHook details indicate whether the saveable tools list is "catalog-derived" or "fallback constant"

### Requirement: Agent command group registration
The `agent tools` and `agent hooks` subcommands SHALL be registered under the existing `lango agent` command group in `cmd/lango/main.go`.

#### Scenario: Agent help lists new subcommands
- **WHEN** user runs `lango agent --help`
- **THEN** the help output includes tools and hooks in the available subcommands list
