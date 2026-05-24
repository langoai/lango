## MODIFIED Requirements

### Requirement: First-touch public docs mirror current CLI and cockpit operator surfaces
The repository SHALL keep first-touch public docs aligned with the runtime behavior and discoverability of the current CLI and cockpit entry surfaces.

#### Scenario: README mentions compact param previews in tool lifecycle rows
- **WHEN** the README describes cockpit or chat transcript behavior
- **THEN** it SHALL mention that tool lifecycle rows can carry compact param previews through running and completed states

#### Scenario: CLI overview mentions compact param previews in tool lifecycle rows
- **WHEN** the CLI core overview describes cockpit chat visibility
- **THEN** it SHALL mention the same compact param preview behavior for tool lifecycle rows
