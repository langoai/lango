## MODIFIED Requirements
### Requirement: Public cockpit docs describe Sessions page behavior
Public cockpit documentation SHALL describe the Sessions page using the current runtime contract.

#### Scenario: README describes Sessions ordering and empty-state split
- **WHEN** a user reads the cockpit shortcut table in `README.md`
- **THEN** the Sessions row SHALL describe the page as a newest-first session summary list
- **AND** it SHALL mention the page's explicit unavailable or empty-state messaging
