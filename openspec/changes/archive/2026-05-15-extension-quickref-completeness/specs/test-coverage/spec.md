## ADDED Requirements

### Requirement: Extension quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango extension` command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented extension quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango extension inspect <source>`, `install <source>`, `list`, and `remove <name>` commands
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries
