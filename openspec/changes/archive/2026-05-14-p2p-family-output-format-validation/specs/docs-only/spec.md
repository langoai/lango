## ADDED Requirements
### Requirement: P2P family docs stay aligned with explicit output-format contracts
Public docs and downstream main specs for the migrated P2P operator family SHALL not drift back to boolean `--json` output docs.

#### Scenario: Stale P2P family `--json` docs are rejected
- **WHEN** public docs or downstream main specs for the migrated P2P operator family reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail
