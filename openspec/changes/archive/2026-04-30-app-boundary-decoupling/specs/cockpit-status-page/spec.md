## MODIFIED Requirements

### Requirement: Feature status and metrics dashboard
StatusPage SHALL display feature flags from a `func() []types.FeatureStatus` status provider, token usage and tool execution stats from `MetricsCollector.Snapshot()`, and provider/model info from Config. StatusPage MUST NOT import or depend on `internal/app`.

#### Scenario: Feature flags display
- **WHEN** StatusPage is active and the provider returns feature statuses
- **THEN** it SHALL render each feature with enabled/disabled badge

#### Scenario: Token usage display
- **WHEN** StatusPage is active
- **THEN** it SHALL show input, output, total, and cache token counts from Snapshot
