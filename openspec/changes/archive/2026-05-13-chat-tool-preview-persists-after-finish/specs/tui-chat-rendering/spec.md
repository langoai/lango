## MODIFIED Requirements

### Requirement: Tool lifecycle transcript items remain operator-informative
The chat transcript SHALL keep tool lifecycle rows informative enough for a live operator to understand what is running, not just that something is running.

#### Scenario: Completed tool row keeps compact param preview
- **WHEN** a tool transcript row transitions from running to success or error
- **AND** the invocation had request params
- **THEN** the row SHALL continue showing the compact param preview alongside the completion state
