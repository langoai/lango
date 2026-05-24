## ADDED Requirements
### Requirement: Migrated CLI prose and examples stay aligned with explicit output-format contracts
Public docs for migrated CLI surfaces SHALL not drift back to stale `--json` examples or prose when the implementation already expects `--output table|json`.

#### Scenario: Stale migrated CLI prose is rejected
- **WHEN** a migrated CLI doc reintroduces prose like `Use --json` or an example such as `lango payment balance --json`
- **THEN** an executable repository test SHALL fail
