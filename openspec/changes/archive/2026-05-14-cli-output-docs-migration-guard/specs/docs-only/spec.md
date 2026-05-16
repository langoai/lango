## ADDED Requirements
### Requirement: Migrated CLI docs stay aligned with explicit output-format contracts
Public docs and main specs for CLI families already migrated from `--json` toggles to `--output table|json` SHALL not drift back to the old flag shape.

#### Scenario: Stale migrated CLI `--json` docs are rejected
- **WHEN** public docs or main specs for migrated command families reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail
