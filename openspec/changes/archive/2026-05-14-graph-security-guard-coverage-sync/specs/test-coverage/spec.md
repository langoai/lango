## ADDED Requirements
### Requirement: Migrated graph and security CLI stay covered by output-contract guards
Once the graph and security inspection CLI families have migrated to `--output table|json`, the executable migrated-family code and docs guards SHALL continue to cover them.

#### Scenario: Graph and security families remain inside migrated-family guard scope
- **WHEN** the repository test suite scans migrated CLI families for stale `--json` regressions
- **THEN** `internal/cli/graph`, `internal/cli/security`, their public docs, and their main specs SHALL be included in that scan
