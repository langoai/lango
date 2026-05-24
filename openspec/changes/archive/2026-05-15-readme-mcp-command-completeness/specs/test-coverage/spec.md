## ADDED Requirements
### Requirement: README MCP completeness guard stays executable
Repository-level regressions that drop implemented `lango mcp` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README MCP entries are rejected
- **WHEN** the implemented `lango mcp` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
