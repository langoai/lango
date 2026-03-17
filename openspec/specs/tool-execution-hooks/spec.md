## ADDED Requirements

### Requirement: Hook interfaces
The `toolchain` package SHALL define `PreToolHook` and `PostToolHook` interfaces. PreToolHook SHALL have `PreExecute(ctx HookContext) (PreHookResult, error)`. PostToolHook SHALL have `PostExecute(ctx HookContext, result string, err error) error`.

#### Scenario: PreToolHook blocks execution
- **WHEN** a PreToolHook returns PreHookResult with Action=Block
- **THEN** the tool SHALL NOT execute and the block message SHALL be returned to the caller

#### Scenario: PostToolHook receives result
- **WHEN** a tool execution completes
- **THEN** all registered PostToolHooks SHALL receive the execution result and any error

### Requirement: PreHookResult actions
PreHookResult SHALL support three actions: Continue (proceed with execution), Block (prevent execution with message), and Modify (change input parameters before execution).

#### Scenario: Continue action
- **WHEN** PreHookResult has Action=Continue
- **THEN** the tool SHALL execute normally with original parameters

#### Scenario: Modify action
- **WHEN** PreHookResult has Action=Modify and ModifiedInput is set
- **THEN** the tool SHALL execute with the modified input parameters

### Requirement: HookRegistry with priority ordering
The `HookRegistry` SHALL maintain hooks ordered by priority (lower number = earlier execution). Hooks SHALL be registered with a name and priority.

#### Scenario: Priority ordering
- **WHEN** hooks with priorities 50, 10, and 100 are registered
- **THEN** they SHALL execute in order: 10, 50, 100

### Requirement: WithHooks middleware bridge
The package SHALL provide a `WithHooks(registry)` function that returns a `Middleware`. This middleware SHALL execute PreHooks before tool execution and PostHooks after, integrating with the existing Chain/ChainAll infrastructure.

#### Scenario: Middleware integration
- **WHEN** WithHooks middleware is applied via ChainAll
- **THEN** PreHooks SHALL execute before each tool and PostHooks after each tool

### Requirement: SecurityFilterHook blocks dangerous command patterns
The SecurityFilterHook (priority 10) SHALL include a set of default blocked patterns that are always active regardless of user configuration. Default patterns SHALL include catastrophic operations: `rm -rf /`, `mkfs.`, `dd if=/dev/zero`, fork bomb, `> /dev/sda`, `chmod -R 777 /`, `dd if=/dev/random`, `mv / `. User-configured patterns SHALL be merged with defaults, with case-insensitive deduplication. All patterns SHALL be pre-lowercased at construction time to avoid repeated lowercasing in the Pre() hot path.

#### Scenario: Default pattern blocks rm -rf
- **WHEN** agent executes `rm -rf /` via exec tool
- **THEN** SecurityFilterHook blocks it with reason "command matches blocked pattern: rm -rf /"

#### Scenario: User patterns merged with defaults
- **WHEN** SecurityFilterHook is constructed with user pattern "DROP TABLE"
- **THEN** both default patterns and "DROP TABLE" are active

#### Scenario: Duplicate patterns deduplicated
- **WHEN** user configures "rm -rf /" which is already a default
- **THEN** the pattern appears only once in the merged list

### Requirement: SecurityFilterHook always registered
The SecurityFilterHook SHALL be registered unconditionally in the tool hook pipeline, not gated by `cfg.Hooks.Enabled` or `cfg.Hooks.SecurityFilter`. Other hooks (AccessControl, EventPublishing) remain config-gated.

#### Scenario: Security hook active without config
- **WHEN** hooks.enabled is false and hooks.securityFilter is false
- **THEN** SecurityFilterHook is still active with default patterns

### Requirement: AgentAccessControlHook
A built-in AgentAccessControlHook (priority 20) SHALL enforce per-agent tool access control lists, blocking tools not in the agent's allowed set.

#### Scenario: Unauthorized tool blocked
- **WHEN** an agent attempts to use a tool not in its ACL
- **THEN** AgentAccessControlHook SHALL block the execution

### Requirement: EventBusHook
A built-in EventBusHook (priority 50) SHALL publish tool execution events to the EventBus after each tool execution.

#### Scenario: Tool event published
- **WHEN** a tool execution completes
- **THEN** EventBusHook SHALL publish a ToolExecutedEvent with tool name, agent name, duration, and success status

### Requirement: KnowledgeSaveHook
A built-in KnowledgeSaveHook (priority 100) SHALL automatically save significant tool results to the knowledge store.

#### Scenario: Result saved to knowledge
- **WHEN** a tool execution returns a result exceeding the minimum significance threshold
- **THEN** KnowledgeSaveHook SHALL save the result to the knowledge store
