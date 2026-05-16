## ADDED Requirements

### Requirement: README includes implemented provenance commands
The README quick reference SHALL include the implemented `lango provenance` command family that is already present in the public CLI index and dedicated provenance docs.

#### Scenario: Implemented provenance commands stay discoverable
- **WHEN** a maintainer updates `README.md`
- **THEN** it SHALL include the implemented `lango provenance status`
- **AND** it SHALL include `lango provenance checkpoint list`, `lango provenance checkpoint create`, and `lango provenance checkpoint show <id>`
- **AND** it SHALL include `lango provenance session tree` and `lango provenance session list`
- **AND** it SHALL include `lango provenance attribution show <session>` and `lango provenance attribution report`
- **AND** it SHALL include `lango provenance bundle export` and `lango provenance bundle import`
