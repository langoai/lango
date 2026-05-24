## MODIFIED Requirements

### Requirement: Tool lifecycle transcript items remain operator-informative
The chat transcript SHALL keep tool lifecycle rows informative enough for a live operator to understand what is running, not just that something is running.

#### Scenario: Approval transcript events surface compact summary previews
- **WHEN** an approval transcript event is rendered for a request that has a summary
- **THEN** the event text SHALL include a compact request summary preview in addition to the request-id annotation
