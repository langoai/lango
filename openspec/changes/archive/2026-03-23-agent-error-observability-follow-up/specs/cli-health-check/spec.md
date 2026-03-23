## ADDED Requirements

### Requirement: Doctor recent-failure output includes cause metadata
The doctor command SHALL show classified cause metadata for recent failed multi-agent traces.

#### Scenario: Multi-agent check shows classified failures
- **WHEN** recent failed traces exist
- **THEN** the `Multi-Agent` doctor check SHALL include `trace_id`, `outcome`, `error_code`, `cause_class`, and `summary`

#### Scenario: Doctor JSON preserves cause metadata
- **WHEN** `lango doctor --json` reports recent failed multi-agent traces
- **THEN** the same classified fields SHALL be present in machine-readable output
