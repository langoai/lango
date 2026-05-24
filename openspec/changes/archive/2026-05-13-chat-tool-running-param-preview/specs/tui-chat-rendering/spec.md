## MODIFIED Requirements

### Requirement: Tool lifecycle transcript items remain operator-informative
The chat transcript SHALL keep tool lifecycle rows informative enough for a live operator to understand what is running, not just that something is running.

#### Scenario: Running tool row preserves compact param preview
- **WHEN** a tool invocation begins with request params
- **THEN** the running transcript row SHALL include a compact param preview
- **AND** that preview SHALL use deterministic param ordering
