## ADDED Requirements
### Requirement: Migrated agent inspection subset stays covered by code-level output-contract guards
Once `lango agent status`, `list`, `tools`, and `hooks` have migrated to `--output table|json`, an executable repository test SHALL continue to cover those specific files for stale boolean `--json` regressions.

#### Scenario: Agent inspection subset remains inside code-level guard scope
- **WHEN** the repository test suite scans migrated CLI code for stale `--json` regressions
- **THEN** the `status.go`, `list.go`, `catalog.go`, and `hooks.go` files under `internal/cli/agent` SHALL be included in that scan
