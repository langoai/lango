## MODIFIED Requirements

### Requirement: Tool lifecycle transcript items remain operator-informative
The chat transcript SHALL keep tool lifecycle rows informative enough for a live operator to understand what is running, not just that something is running.

#### Scenario: Tool approval transition updates the latest matching row
- **WHEN** multiple running tool transcript rows share the same tool name
- **AND** an approval-driven tool state transition is applied for that tool name
- **THEN** the most recent matching running row SHALL receive the new state
