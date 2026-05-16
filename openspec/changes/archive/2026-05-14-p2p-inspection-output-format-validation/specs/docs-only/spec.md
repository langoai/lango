## ADDED Requirements
### Requirement: P2P inspection docs stay aligned with explicit output-format contracts
Public docs and main specs for the migrated P2P inspection subset SHALL not drift back to boolean `--json` output docs.

#### Scenario: Stale P2P inspection `--json` docs are rejected
- **WHEN** public docs or main specs for `lango p2p status`, `peers`, or `identity` reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail
