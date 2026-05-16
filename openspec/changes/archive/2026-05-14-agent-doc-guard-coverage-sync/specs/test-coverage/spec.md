## ADDED Requirements
### Requirement: Migrated agent inspection docs stay covered by output-contract guards
Once the agent inspection subset has migrated to `--output table|json`, the executable migrated-family docs guard SHALL continue to cover it.

#### Scenario: Agent inspection subset remains inside migrated-family docs guard scope
- **WHEN** the repository test suite scans migrated CLI docs for stale `--json` regressions
- **THEN** public docs and the main spec for `lango agent status`, `list`, `tools`, and `hooks` SHALL be included in that scan
