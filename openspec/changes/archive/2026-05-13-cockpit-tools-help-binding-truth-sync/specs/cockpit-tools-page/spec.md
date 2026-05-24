## ADDED Requirements

### Requirement: Tools page help bindings match supported navigation
The cockpit Tools page SHALL expose only the key bindings it actually handles.

#### Scenario: Tools page help lists only vertical navigation bindings
- **WHEN** the Tools page help is rendered
- **THEN** it SHALL expose up/down navigation bindings
- **AND** it SHALL NOT advertise an `Esc back` binding unless the page implements that behavior
