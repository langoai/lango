## MODIFIED Requirements
### Requirement: Tools page help bindings match supported navigation
The cockpit Tools page SHALL expose only the key bindings it actually handles.

#### Scenario: Tools page help lists vertical navigation only when another category exists
- **WHEN** the Tools page help is rendered with two or more categories
- **THEN** it SHALL advertise `↑/k` and `↓/j`
- **AND** it SHALL NOT advertise an `Esc back` binding unless the page implements that behavior

#### Scenario: Tools page hides navigation help with fewer than two categories
- **WHEN** the Tools page help is rendered with zero or one category
- **THEN** it SHALL omit vertical navigation bindings
