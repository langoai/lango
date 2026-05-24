## ADDED Requirements
### Requirement: Migrated CLI docs drift guards stay executable
Repository-level docs guards SHALL continue to cover stale prose and downstream spec references for migrated output-format contracts, not just flag tables.

#### Scenario: Stale migrated prose and downstream spec references are rejected
- **WHEN** a guarded migrated CLI doc or main spec reintroduces stale `Use --json`, `support --json`, or `--json` example wording
- **THEN** an executable repository test SHALL fail
