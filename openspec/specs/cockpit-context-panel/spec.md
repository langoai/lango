## ADDED Requirements

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
