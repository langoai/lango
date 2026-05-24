## MODIFIED Requirements

### Requirement: First-touch public docs mirror current CLI and cockpit operator surfaces
The repository SHALL keep first-touch public docs aligned with the runtime behavior and discoverability of the current CLI and cockpit entry surfaces.

#### Scenario: README mentions compact request-id annotations on approval events
- **WHEN** the README describes chat transcript approval visibility
- **THEN** it SHALL mention that approval transcript events can carry compact request-id annotations

#### Scenario: CLI overview mentions compact request-id annotations on approval events
- **WHEN** the CLI core overview describes cockpit chat transcript behavior
- **THEN** it SHALL mention the same approval-event traceability improvement
