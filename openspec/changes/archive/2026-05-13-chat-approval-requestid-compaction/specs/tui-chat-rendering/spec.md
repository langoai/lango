## MODIFIED Requirements

### Requirement: Tool lifecycle transcript items remain operator-informative
The chat transcript SHALL keep tool lifecycle rows informative enough for a live operator to understand what is running, not just that something is running.

#### Scenario: Approval transcript request IDs are compacted
- **WHEN** an approval transcript event is rendered for a long request ID
- **THEN** the visible request-id annotation SHALL be compacted instead of rendering the full raw request ID
