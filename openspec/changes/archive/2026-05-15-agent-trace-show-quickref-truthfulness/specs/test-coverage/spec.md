## ADDED Requirements

### Requirement: Agent-diagnostics quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango agent` diagnostics command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented agent-diagnostics quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango agent trace list`, `lango agent trace show <trace-id>`, `lango agent graph <session>`, and `lango agent trace metrics` commands
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries
