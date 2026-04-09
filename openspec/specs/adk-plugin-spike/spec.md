## Purpose

Capability spec for adk-plugin-spike. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Plugin config pass-through to ADK runner
The system SHALL provide an option to pass `plugin.Config` (ADK v0.6.0 `plugin/plugin.go:26`) through to `runner.Config.PluginConfig` when creating an ADK runner in `internal/adk/agent.go`.

#### Scenario: Runner created with plugin config
- **WHEN** an agent is created with ADK plugin options specified
- **THEN** the runner.Config.PluginConfig field SHALL contain the provided plugins
- **AND** the ADK runner SHALL invoke plugin callbacks during its lifecycle

#### Scenario: Runner created without plugin config (backward compatible)
- **WHEN** an agent is created without ADK plugin options
- **THEN** the runner.Config.PluginConfig SHALL be empty (zero value)
- **AND** the runner SHALL function identically to current behavior

### Requirement: Parity gap analysis for toolchain middleware
The spike SHALL produce a parity gap table comparing each existing toolchain middleware (`internal/toolchain/mw_*.go`) against ADK plugin callback capabilities.

#### Scenario: Per-tool vs agent-level granularity documented
- **WHEN** the parity analysis is performed
- **THEN** the analysis SHALL document that ADK `BeforeToolCallback`/`AfterToolCallback` are agent-level (apply to all tools) while Lango's middleware chain supports per-tool application
- **AND** each middleware SHALL be classified as "movable to ADK callback" or "must remain in toolchain"

#### Scenario: SecurityFilterHook mapping evaluated
- **WHEN** the `BeforeToolCallback` mapping is evaluated for `SecurityFilterHook`
- **THEN** the analysis SHALL verify that returning a non-nil result from `BeforeToolCallback` blocks tool execution
- **AND** the analysis SHALL document whether the hook's tool-name-based filtering can be expressed within the callback

#### Scenario: Learning observation mapping evaluated
- **WHEN** the `AfterToolCallback` mapping is evaluated for `WithLearning` middleware
- **THEN** the analysis SHALL verify that `AfterToolCallback` receives tool name, args, result, and error
- **AND** the analysis SHALL document whether learning engine access can be provided through callback context
