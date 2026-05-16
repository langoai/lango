## ADDED Requirements
### Requirement: Migrated payment CLI stays covered by output-contract guards
Once the payment CLI family has migrated to `--output table|json`, the executable migrated-family code and docs guards SHALL continue to cover it.

#### Scenario: Payment family remains inside migrated-family guard scope
- **WHEN** the repository test suite scans migrated CLI families for stale `--json` regressions
- **THEN** `internal/cli/payment`, payment CLI docs, and the payment CLI main spec SHALL be included in that scan
