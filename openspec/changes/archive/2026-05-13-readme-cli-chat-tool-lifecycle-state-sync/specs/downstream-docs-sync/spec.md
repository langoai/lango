## MODIFIED Requirements

### Requirement: First-touch public docs mirror current CLI and cockpit operator surfaces
The repository SHALL keep first-touch public docs aligned with the runtime behavior and discoverability of the current CLI and cockpit entry surfaces.

#### Scenario: README mentions approval-wait and canceled tool lifecycle states
- **WHEN** the README describes cockpit or chat transcript tool lifecycle behavior
- **THEN** it SHALL mention that compact param previews persist through approval waits and canceled states, not only success/error

#### Scenario: CLI overview mentions approval-wait and canceled tool lifecycle states
- **WHEN** the CLI core overview describes cockpit chat tool lifecycle visibility
- **THEN** it SHALL mention the same approval-wait and canceled state coverage
