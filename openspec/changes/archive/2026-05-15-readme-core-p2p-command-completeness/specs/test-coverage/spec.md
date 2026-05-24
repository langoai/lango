## ADDED Requirements
### Requirement: README core P2P completeness guard stays executable
Repository-level regressions that drop implemented core `p2p` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README core P2P entries are rejected
- **WHEN** the implemented core `p2p` CLI families remain shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
