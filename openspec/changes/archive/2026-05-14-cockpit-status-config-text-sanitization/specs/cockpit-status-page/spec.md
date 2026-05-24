## MODIFIED Requirements

### Requirement: Feature status and metrics dashboard
StatusPage SHALL display feature flags from a `func() []types.FeatureStatus` status provider, token usage and tool execution stats from `MetricsCollector.Snapshot()`, and provider/model info from Config. StatusPage MUST NOT import or depend on `internal/app`.

#### Scenario: Rendered status-page config labels stay plain and single-line
- **WHEN** provider or model labels contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Status page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed config text to a single line before rendering it
