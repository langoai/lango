## Purpose

Capability spec for cockpit-status-page. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Feature status and metrics dashboard
StatusPage SHALL display feature flags from a `func() []types.FeatureStatus` status provider, token usage and tool execution stats from `MetricsCollector.Snapshot()`, and provider/model info from Config. StatusPage MUST NOT import or depend on `internal/app`.

#### Scenario: Feature flags display
- **WHEN** StatusPage is active and the provider returns feature statuses
- **THEN** it SHALL render each feature with enabled/disabled badge

#### Scenario: Token usage display
- **WHEN** StatusPage is active
- **THEN** it SHALL show input, output, total, and cache token counts from Snapshot

#### Scenario: Rendered status-page config labels stay plain and single-line
- **WHEN** provider or model labels contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Status page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed config text to a single line before rendering it

### Requirement: Auto-refresh via tea.Tick
StatusPage SHALL refresh metrics every 5 seconds using `tea.Tick`. The tick SHALL start on `Activate()` and stop on `Deactivate()`.

#### Scenario: Activate starts tick
- **WHEN** StatusPage.Activate() is called
- **THEN** it SHALL return a tea.Cmd that triggers the first tick

#### Scenario: Deactivate stops tick
- **WHEN** StatusPage.Deactivate() is called
- **THEN** subsequent tick callbacks SHALL not schedule new ticks

### Requirement: Graph admission metrics are surfaced on the cockpit status page
The cockpit status page SHALL surface observe-only graph admission metrics from the runtime feedback snapshot, including graph-admission counts grouped by source and validator identity, extractor dropped-unknown baselines, unmapped-source counts, and aggregate graph write-failure baselines.

#### Scenario: Status page renders graph admission metrics
- **WHEN** the cockpit status page is rendered while observe mode metrics are available
- **THEN** it SHALL display event-bus graph-admission counts grouped by supported producer source and producer group, plus a separate grouped view for the synthetic `content_saved_extractor` source label
- **AND** it SHALL display validator-source as a grouping key on graph-admission metrics rather than as a separate independent metric family
- **AND** it SHALL display extractor dropped-unknown, unmapped-source, and aggregate graph write-failure baseline counts as distinct metrics
- **AND** it SHALL preserve raw `unmapped-source` identity by grouping those counts by raw source label
- **AND** it SHALL preserve validator-source identity by grouping those counts by validator-source identifier
- **AND** it SHALL display `known`, `unknown`, and `unvalidated` triple totals for graph-admission decisions

### Requirement: Status page renders explicit unavailable messaging for missing dependencies
The cockpit Status page SHALL distinguish missing feature-status or observability dependencies from valid zero-valued data.

#### Scenario: Missing feature-status provider renders unavailable message
- **WHEN** the Status page renders with no feature-status provider
- **THEN** the Feature Status section SHALL explain that the feature status provider is not configured

#### Scenario: Missing metrics collector renders unavailable message
- **WHEN** the Status page renders with no observability metrics collector
- **THEN** the Token Usage, Tool Execution, and Graph Admission sections SHALL explain that the metrics collector is not configured
