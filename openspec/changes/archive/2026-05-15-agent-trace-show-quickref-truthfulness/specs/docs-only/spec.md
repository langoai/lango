## ADDED Requirements

### Requirement: Public quick references include implemented agent-diagnostics commands
The public quick-reference docs SHALL include the implemented `lango agent` diagnostics commands that are already present in the CLI index and core command docs.

#### Scenario: Implemented agent-diagnostics commands stay discoverable
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL include the implemented `lango agent trace list`, `lango agent trace show <trace-id>`, `lango agent graph <session>`, and `lango agent trace metrics` command entries
