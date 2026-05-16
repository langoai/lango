## ADDED Requirements
### Requirement: CLI output-flag contract guards stay executable
Repository-level CLI output-flag regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Boolean `--output` flags are rejected
- **WHEN** CLI production code reintroduces a boolean `--output` flag declaration
- **THEN** an executable repository test SHALL fail
