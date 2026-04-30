## MODIFIED Requirements

### Requirement: StatusCollector aggregation
The system SHALL provide a `StatusCollector` in the app layer that collects `FeatureStatus` from wiring functions. It SHALL expose `All()` to list all statuses and `SilentDisabledCount()` to count features that are disabled with a non-empty reason. CLI and TUI packages SHALL consume feature statuses through `[]types.FeatureStatus` or provider functions and MUST NOT import `internal/app` to access the collector.

#### Scenario: Silent disabled count
- **WHEN** StatusCollector has 3 features: knowledge (enabled), embedding (disabled, reason="no provider"), graph (disabled, reason="")
- **THEN** `SilentDisabledCount()` returns 1 (only embedding has a reason)

#### Scenario: CLI/TUI feature status consumption
- **WHEN** CLI or TUI code needs feature statuses
- **THEN** the status data is provided as `[]types.FeatureStatus` or `func() []types.FeatureStatus`
- **AND** the CLI or TUI package does not import `internal/app`
