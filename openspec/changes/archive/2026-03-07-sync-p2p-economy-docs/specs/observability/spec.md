## ADDED Requirements

### Requirement: Observability feature documentation page
The documentation site SHALL include a `docs/features/observability.md` page documenting the observability system including metrics collector, token tracking, health checks, audit logging, and gateway endpoints, with experimental warning, architecture mermaid diagram, and configuration reference.

#### Scenario: Observability feature docs page exists
- **WHEN** the documentation site is built
- **THEN** `docs/features/observability.md` SHALL exist with sections for metrics, token tracking, health checks, audit logging, and API endpoints

### Requirement: Metrics CLI documentation page
The documentation site SHALL include a `docs/cli/metrics.md` page documenting `lango metrics`, `lango metrics sessions`, `lango metrics tools`, `lango metrics agents`, and `lango metrics history` commands with flags tables and example output following the `docs/cli/payment.md` pattern.

#### Scenario: Metrics CLI docs page exists
- **WHEN** the documentation site is built
- **THEN** `docs/cli/metrics.md` SHALL exist with sections for all 5 metrics subcommands

#### Scenario: Persistent flags documented
- **WHEN** a user reads the metrics CLI reference
- **THEN** `--output` (table|json) and `--addr` (default http://localhost:18789) persistent flags SHALL be documented
