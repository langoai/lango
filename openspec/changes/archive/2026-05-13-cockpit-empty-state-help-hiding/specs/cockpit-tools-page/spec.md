## MODIFIED Requirements
### Requirement: Tools page help bindings match supported navigation
The cockpit Tools page SHALL expose only the key bindings it actually handles.

#### Scenario: Tools page help lists only vertical navigation bindings when categories exist
- **WHEN** the Tools page help is rendered with at least one category available
- **THEN** it SHALL advertise `↑/k` and `↓/j`
- **AND** it SHALL NOT advertise an `Esc back` binding unless the page implements that behavior

#### Scenario: Tools page hides navigation help without categories
- **WHEN** the Tools page help is rendered with no configured catalog or no registered categories
- **THEN** it SHALL omit vertical navigation bindings
