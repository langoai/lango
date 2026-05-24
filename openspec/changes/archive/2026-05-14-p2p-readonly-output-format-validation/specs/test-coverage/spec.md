## ADDED Requirements
### Requirement: P2P read-only output-contract guards stay executable
Repository-level regressions that reintroduce boolean `--json` flags or stale output-contract docs for the migrated P2P read-only inspection subset SHALL be enforced by executable tests.

#### Scenario: P2P read-only production code rejects boolean `--json` regressions
- **WHEN** `internal/cli/p2p/discover.go`, `pricing.go`, `reputation.go`, or `session.go` reintroduces a boolean `--json` flag declaration
- **THEN** an executable repository test SHALL fail

#### Scenario: P2P read-only docs reject stale `--json` regressions
- **WHEN** public docs or main specs for the migrated P2P read-only inspection subset reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail
