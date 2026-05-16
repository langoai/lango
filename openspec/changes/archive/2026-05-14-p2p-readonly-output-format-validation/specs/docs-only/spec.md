## ADDED Requirements
### Requirement: P2P read-only docs stay aligned with explicit output-format contracts
Public docs and downstream main specs for the migrated P2P read-only inspection subset SHALL not drift back to boolean `--json` output docs.

#### Scenario: Stale P2P read-only `--json` docs are rejected
- **WHEN** public docs or downstream main specs for `discover`, `pricing`, `reputation`, or `session list` reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail
