## ADDED Requirements

### Requirement: README agent-inspection completeness guard stays executable
Repository-level regressions that drop implemented `lango agent` inspection command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented agent-inspection quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango agent status`, `lango agent list`, `lango agent tools`, and `lango agent hooks` commands
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries

### Requirement: README graph completeness guard stays executable
Repository-level regressions that drop implemented `lango graph` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented graph quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, and `import` commands
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
