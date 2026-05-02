## ADDED Requirements

### Requirement: Approval-blocked replacement stays durable
When a built-in teammate run remains approval-blocked but its blocked metadata changes, the durable mirror SHALL record the replacement and refresh the latest snapshot values.

#### Scenario: Approval-blocked metadata changes while the run stays blocked
- **WHEN** a built-in teammate run remains `blocked_waiting_approval`
- **AND** either `blocked_reason` or `grant_request_id` changes
- **THEN** RunLedger SHALL append a fresh approval-block journal event
- **AND** the cached durable snapshot SHALL retain the latest blocked condition, blocked reason, and grant request ID
