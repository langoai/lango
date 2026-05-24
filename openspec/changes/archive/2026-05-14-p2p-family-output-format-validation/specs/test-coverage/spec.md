## ADDED Requirements
### Requirement: P2P family output-contract guard stays executable
Repository-level regressions that reintroduce boolean `--json` flags or stale output-contract docs across the migrated P2P operator family SHALL be enforced by executable tests.

#### Scenario: P2P family production code rejects boolean `--json` regressions
- **WHEN** the remaining migrated P2P operator files reintroduce a boolean `--json` flag declaration
- **THEN** an executable repository test SHALL fail

#### Scenario: P2P family docs reject stale `--json` regressions
- **WHEN** public docs or main specs for the migrated P2P operator family reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail
