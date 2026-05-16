## Purpose

Capability spec for cockpit-context-panel. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Toggleable right context panel
The cockpit SHALL support a right-side context panel (Ctrl+P toggle) displaying live token usage, tool execution stats, and uptime from MetricsCollector.Snapshot(). The panel SHALL NOT be a Page — it uses Start()/Stop() lifecycle managed by the cockpit toggle.

#### Scenario: Toggle context panel on
- **WHEN** user presses Ctrl+P with contextVisible=false
- **THEN** the context panel SHALL appear on the right, Start() SHALL be called, and all components SHALL receive updated WindowSizeMsg with reduced width

#### Scenario: Toggle context panel off
- **WHEN** user presses Ctrl+P with contextVisible=true
- **THEN** the context panel SHALL disappear, Stop() SHALL be called, and all components SHALL receive updated WindowSizeMsg with increased width

### Requirement: Context panel auto-refresh
The context panel SHALL refresh metrics every 5 seconds when visible. Refresh SHALL stop when hidden.

#### Scenario: Auto-refresh while visible
- **WHEN** context panel is visible
- **THEN** it SHALL call MetricsCollector.Snapshot() every 5 seconds and re-render

#### Scenario: Stop refresh when hidden
- **WHEN** Stop() is called
- **THEN** subsequent tick callbacks SHALL not schedule new ticks

### Requirement: Context panel renders token and tool metrics
The context panel SHALL display token usage (input/output/total/cache), top-5 tools by execution count, and system uptime.

#### Scenario: Token usage display
- **WHEN** context panel is visible
- **THEN** it SHALL show input, output, total, and cache token counts

#### Scenario: Tool stats display
- **WHEN** context panel is visible with tool executions recorded
- **THEN** it SHALL show up to 5 tools sorted by execution count descending

#### Scenario: System uptime display
- **WHEN** context panel is visible
- **THEN** it SHALL render uptime from the metrics snapshot

#### Scenario: Rendered context-panel labels stay plain and single-line
- **WHEN** top-tool names, runtime active-agent labels, or channel names contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the context panel SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it

### Requirement: Channel status section in context panel
The context panel SHALL display a "Channels" section showing each channel's connection status (connected/disconnected indicator), name, and message count.

#### Scenario: Connected channel displayed
- **WHEN** a channel with Connected=true and MessageCount=5 is set
- **THEN** the context panel renders a green "●" indicator, the channel name, and "5 msgs"

#### Scenario: Channel snapshot labels are replay-safe
- **WHEN** channel names contain ANSI/OSC escape sequences or embedded newlines before entering the channel snapshot
- **THEN** the channel tracker SHALL strip those control sequences
- **AND** it SHALL normalize the stored channel snapshot text to a single line before replay

#### Scenario: Disconnected channel displayed
- **WHEN** a channel with Connected=false is set
- **THEN** the context panel renders a red "○" indicator

#### Scenario: No channels configured
- **WHEN** no channel statuses are set
- **THEN** the "Channels" section is not rendered (graceful degradation)

#### Scenario: Channel statuses updated on tick
- **WHEN** the context panel tick fires and a ChannelTracker is available
- **THEN** the cockpit calls `tracker.Snapshot()` and pushes results to `SetChannelStatuses`

### Requirement: Runtime status section in context panel
The context panel SHALL display a "Runtime" section showing the active agent, delegation count, and per-turn token usage when a turn is active. The section SHALL appear between Tool Stats and Channels.

#### Scenario: Runtime section when turn is active
- **WHEN** `SetRuntimeStatus` is called with `IsRunning=true`, `ActiveAgent="operator"`, `DelegationCount=3`, `TurnTokens=1234`
- **THEN** the context panel SHALL display a "Runtime" section with a green running indicator, agent name, delegation count, and formatted token count

#### Scenario: Runtime snapshot labels are replay-safe
- **WHEN** active-agent labels contain ANSI/OSC escape sequences or embedded newlines before entering the runtime snapshot
- **THEN** the runtime tracker SHALL strip those control sequences
- **AND** it SHALL normalize the stored runtime snapshot text to a single line before replay

#### Scenario: Runtime section hidden when idle
- **WHEN** `SetRuntimeStatus` is called with `IsRunning=false`
- **THEN** the "Runtime" section SHALL NOT be rendered (graceful degradation)

#### Scenario: Runtime status refreshed on tick
- **WHEN** a contextTickMsg fires and a RuntimeTracker is available
- **THEN** the cockpit SHALL push `runtimeTracker.Snapshot()` to the context panel alongside channel statuses

#### Scenario: Zero delegations not displayed
- **WHEN** `DelegationCount=0` in the runtime status
- **THEN** the delegation line SHALL NOT be rendered

#### Scenario: Zero tokens not displayed
- **WHEN** `TurnTokens=0` in the runtime status
- **THEN** the token line SHALL NOT be rendered

### Requirement: Context panel optimized snapshot handling
The ContextPanel SHALL reuse existing slice capacity in SetChannelStatuses() instead of allocating a new slice on every call. Style variables for render methods SHALL be pre-allocated at module level. The toolCountSum SHALL be cached alongside the sortedTools dirty flag.

#### Scenario: SetChannelStatuses reuses slice capacity
- **WHEN** SetChannelStatuses is called with a status list of equal or smaller length than existing capacity
- **THEN** the existing slice SHALL be resliced and copied without new allocation

#### Scenario: Context panel setters keep cached labels replay-safe
- **WHEN** channel names or runtime active-agent labels contain ANSI/OSC escape sequences or embedded newlines before entering the context panel setters
- **THEN** the context panel SHALL strip those control sequences
- **AND** it SHALL normalize the cached setter-owned text to a single line before replay

#### Scenario: Render styles pre-allocated
- **WHEN** renderRuntimeStatus or renderChannelStatus renders status items
- **THEN** they SHALL use module-level pre-allocated style variables instead of inline lipgloss.NewStyle()

#### Scenario: Tool count sum cached with dirty flag
- **WHEN** the sortedTools dirty flag is false and toolCountSum is needed
- **THEN** the cached sum SHALL be returned without iterating the tool breakdown map

#### Scenario: Cached tool labels are replay-safe
- **WHEN** tool names contain ANSI/OSC escape sequences or embedded newlines before entering the cached `sortedTools` slice
- **THEN** the context panel SHALL strip those control sequences
- **AND** it SHALL normalize the stored cached tool text to a single line before replay

### Requirement: Context panel renders unavailable messaging when metrics are absent
The cockpit context panel SHALL distinguish an unavailable metrics collector from valid zero-valued metric data.

#### Scenario: Nil metrics collector renders unavailable messages
- **WHEN** the context panel renders with no configured metrics collector
- **THEN** the Token Usage section SHALL explain that the metrics collector is not configured
- **AND** the Tool Stats section SHALL explain that the metrics collector is not configured
- **AND** the System section SHALL explain that the metrics collector is not configured
