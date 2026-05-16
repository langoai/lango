## MODIFIED Requirements

### Requirement: Tool lifecycle transcript items remain operator-informative
The chat transcript SHALL keep tool lifecycle rows informative enough for a live operator to understand what is running, not just that something is running.

#### Scenario: Approval transcript events remain width-safe
- **WHEN** an approval transcript event is rendered on a narrow terminal
- **THEN** the event row SHALL truncate to the available width instead of overflowing the transcript layout
