# session-modes Specification

## Purpose
TBD - created by archiving change ux-capability-concierge. Update Purpose after archive.
## Requirements
### Requirement: SessionMode type and built-in modes
The system SHALL define a `SessionMode` type containing `Name`, `Tools` (list of tool names or `@category` references), `Skills` (list of skill names), and `SystemHint` (free-form prompt addition). The system SHALL ship three built-in modes: `code-review`, `research`, and `debug`. User config MAY define additional modes that merge with built-ins by name.

#### Scenario: Built-in code-review mode available
- **WHEN** the system starts with default config
- **THEN** the built-in `code-review` mode SHALL be available for selection

#### Scenario: User-defined mode overrides built-in
- **WHEN** user config defines a mode named `code-review`
- **THEN** the user-defined mode SHALL take precedence over the built-in

#### Scenario: Mode references tool category
- **WHEN** a mode's `Tools` list contains `@exec`
- **THEN** all tools in the `exec` category SHALL be included in the mode's allowlist

### Requirement: Session mode persistence
A `Session` SHALL have an optional `Mode` field (string, zero value = no mode). The mode SHALL persist across turn boundaries within the session. Changing the mode via `/mode` or `--mode` SHALL update the session record.

#### Scenario: Mode persists across turns
- **WHEN** a user sets mode `code-review` and sends a second turn
- **THEN** the second turn SHALL execute with `code-review` mode active

#### Scenario: Session without mode
- **WHEN** a session has no mode set
- **THEN** no mode-based filtering or enforcement SHALL occur (legacy behavior)

### Requirement: Mode resolution from context
The system SHALL provide `session.ModeNameFromContext(ctx) string` and `session.WithModeName(ctx, name string) context.Context` following the existing `SessionKeyFromContext` pattern. The turn runner SHALL set the mode name on the context before invoking the executor. Consumers that need the full `SessionMode` definition SHALL look it up via `config.LookupMode(name)` (this avoids import cycles between `session` and `config`).

#### Scenario: Mode name set on context at turn start
- **WHEN** a turn starts for a session with mode `research`
- **THEN** the executor's context SHALL carry the mode name `"research"` via `ModeNameFromContext`

#### Scenario: Missing mode returns empty string
- **WHEN** `ModeNameFromContext` is called on a context without a mode
- **THEN** it SHALL return `""`

### Requirement: Tool catalog filtering by mode
`Catalog.ListVisibleToolsForMode(mode SessionMode) []ToolSchema` SHALL return only tools whose names appear in the mode's resolved allowlist (after expanding `@category` references). When `mode.Name == ""`, it SHALL behave identically to `ListVisibleTools("")`.

#### Scenario: Mode filters tool list
- **WHEN** `ListVisibleToolsForMode` is called with a mode containing `["builtin_search", "@exec"]`
- **THEN** only `builtin_search` and tools in the `exec` category SHALL be returned

#### Scenario: Empty mode returns full visible set
- **WHEN** `ListVisibleToolsForMode` is called with a zero-value SessionMode
- **THEN** the result SHALL equal `ListVisibleTools("")`

### Requirement: Dynamic tool catalog section in GenerateContent
The `ContextAwareModelAdapter` SHALL accept a `*toolcatalog.Catalog` via `WithCatalog(c *Catalog)`. During `GenerateContent()`, after Phase 2 budget measurement, the adapter SHALL generate a tool catalog prompt section using `Catalog.ListVisibleToolsForMode(currentMode)` and append it to the per-turn prompt. The static tool catalog section SHALL NOT be included in `basePrompt`.

#### Scenario: Tool catalog generated per turn
- **WHEN** `GenerateContent` executes with a catalog wired
- **THEN** the resulting prompt SHALL include a tool catalog section describing the mode-filtered tool set

#### Scenario: basePrompt has no tool catalog
- **WHEN** `ContextAwareModelAdapter` is constructed via `NewContextAwareModelAdapter`
- **THEN** `basePrompt` SHALL NOT contain a `## Tools` or equivalent section sourced from the global catalog

### Requirement: Mode allowlist enforcement middleware
The `toolchain` package SHALL provide `WithModeAllowlist(modeResolver)` middleware. Before invoking the wrapped tool handler, it SHALL resolve the session mode from context. If the mode is non-empty and the current tool name is NOT in the mode's allowlist, the middleware SHALL return an error with message `"tool <name> not available in current mode: <mode>"` without calling the handler.

#### Scenario: Mode-blocked tool returns error
- **WHEN** a tool not in the active mode's allowlist is invoked
- **THEN** the middleware SHALL return an error containing `"not available in current mode"`
- **AND** the underlying handler SHALL NOT be called

#### Scenario: Allowed tool proceeds to handler
- **WHEN** a tool in the active mode's allowlist is invoked
- **THEN** the middleware SHALL call the underlying handler with the original params

#### Scenario: No active mode passes through
- **WHEN** `ModeFromContext` returns no mode
- **THEN** the middleware SHALL call the underlying handler without filtering

### Requirement: Mode system hint injection
When a session has an active mode, `GenerateContent` SHALL include the mode's `SystemHint` in the per-turn prompt as an additional guidance section. An empty `SystemHint` SHALL be omitted.

#### Scenario: SystemHint appended to prompt
- **WHEN** a session runs under a mode with `SystemHint: "Focus on code review."`
- **THEN** the resulting system prompt SHALL contain `"Focus on code review."`

### Requirement: Mode change event
Changing a session's mode SHALL publish a `ModeChangedEvent{SessionKey, OldMode, NewMode}` to the eventbus. The TUI and channel adapters SHALL subscribe and render the change in their native format.

#### Scenario: /mode publishes event
- **WHEN** the user runs `/mode research` in the TUI
- **THEN** a `ModeChangedEvent` SHALL be published with `OldMode` (prior value) and `NewMode="research"`

#### Scenario: TUI renders mode change
- **WHEN** `ModeChangedEvent` is received
- **THEN** the chat view SHALL append a system status entry indicating the new mode

