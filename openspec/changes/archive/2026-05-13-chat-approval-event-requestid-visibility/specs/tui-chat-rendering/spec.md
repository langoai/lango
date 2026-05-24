## MODIFIED Requirements

### Requirement: Tool lifecycle transcript items remain operator-informative
The chat transcript SHALL keep tool lifecycle rows informative enough for a live operator to understand what is running, not just that something is running.

#### Scenario: Approval transcript events surface compact request IDs
- **WHEN** an approval transcript event is rendered for a request that has an ID
- **THEN** the event text SHALL include a compact request-id annotation so the operator can distinguish repeated requests for the same tool
