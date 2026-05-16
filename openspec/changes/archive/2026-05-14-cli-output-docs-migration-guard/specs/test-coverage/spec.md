## ADDED Requirements
### Requirement: Migrated CLI docs output-contract guards stay executable
Repository-level docs regressions that reintroduce stale `--json` UX for migrated CLI families SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Migrated command-family docs reject stale `--json` regressions
- **WHEN** public docs or main specs for migrated command families reintroduce boolean `--json` flag tables or `--json` usage examples
- **THEN** an executable repository test SHALL fail
