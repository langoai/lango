## ADDED Requirements

### Requirement: Feature status and metrics dashboard
StatusPage SHALL display feature flags from `FeatureStatuses.All()`, token usage and tool execution stats from `MetricsCollector.Snapshot()`, and provider/model info from Config.

#### Scenario: Feature flags display
- **WHEN** StatusPage is active
- **THEN** it SHALL render each feature with enabled/disabled badge

#### Scenario: Token usage display
- **WHEN** StatusPage is active
- **THEN** it SHALL show input, output, total, and cache token counts from Snapshot

### Requirement: Auto-refresh via tea.Tick
StatusPage SHALL refresh metrics every 5 seconds using `tea.Tick`. The tick SHALL start on `Activate()` and stop on `Deactivate()`.

#### Scenario: Activate starts tick
- **WHEN** StatusPage.Activate() is called
- **THEN** it SHALL return a tea.Cmd that triggers the first tick

#### Scenario: Deactivate stops tick
- **WHEN** StatusPage.Deactivate() is called
- **THEN** subsequent tick callbacks SHALL not schedule new ticks
