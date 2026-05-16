## ADDED Requirements
### Requirement: Migrated CLI JSON-flag guards stay executable
Repository-level regressions that reintroduce boolean `--json` flags into CLI families already migrated to explicit output-format contracts SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Migrated command families reject boolean `--json` regressions
- **WHEN** production CLI code under migrated families reintroduces a boolean `--json` flag declaration
- **THEN** an executable repository test SHALL fail
