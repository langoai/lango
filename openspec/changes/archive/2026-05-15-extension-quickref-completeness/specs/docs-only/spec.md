## ADDED Requirements

### Requirement: Public quick references include implemented extension commands
The public quick-reference docs SHALL include the implemented `lango extension` command family that is already present in README extension-pack docs and the wired root CLI.

#### Scenario: Implemented extension commands stay discoverable
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL include the implemented `lango extension inspect <source>`, `install <source>`, `list`, and `remove <name>` command entries
