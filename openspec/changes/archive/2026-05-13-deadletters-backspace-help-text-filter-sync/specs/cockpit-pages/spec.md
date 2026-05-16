## MODIFIED Requirements
### Requirement: Dead Letters remains registered as a cockpit degraded surface
The cockpit SHALL keep Dead Letters available as a registered page route even when the dead-letter bridge is unavailable, relying on page-level unavailable messaging instead of suppressing page registration.

#### Scenario: Backspace help describes active text-filter editing
- **WHEN** the Dead Letters help is rendered
- **THEN** the `Backspace` binding SHALL describe editing the active text filter rather than only the query field
