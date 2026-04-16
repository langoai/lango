## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Public BuildHookRegistry helper
The `internal/app` package SHALL export a `BuildHookRegistry(cfg *config.Config, bus *eventbus.Bus, knowledgeSaver toolchain.KnowledgeSaver) *toolchain.HookRegistry` function that produces the same hook registry as the runtime app builder. When `bus` is nil, EventBus hooks are omitted. When `knowledgeSaver` is nil, `KnowledgeSaveHook` is still registered (for snapshot inspection) but its `Post` method safely no-ops. The private `buildHookRegistry` function SHALL delegate to this public helper.

#### Scenario: CLI uses BuildHookRegistry without full bootstrap
- **WHEN** the `agent hooks` CLI command loads config and calls `BuildHookRegistry(cfg, nil, nil)`
- **THEN** the returned registry contains all config-derivable hooks (SecurityFilter, AccessControl, KnowledgeSaveHook)
- **AND** no database connection, crypto initialization, or event bus is required

### Requirement: Default SaveableTools allowlist
The `toolchain` package SHALL define a `DefaultSaveableTools` constant containing the default set of tool names whose results are eligible for knowledge saving. The list SHALL include only read-type tools. The `app.go` builder SHALL use this constant when constructing `KnowledgeSaveHook`.

#### Scenario: KnowledgeSaveHook uses default allowlist
- **WHEN** the app builder constructs `KnowledgeSaveHook` without explicit configuration
- **THEN** the hook's `SaveableTools` set equals `DefaultSaveableTools`
- **AND** only read-type tool names are included (no write or execute tools)
