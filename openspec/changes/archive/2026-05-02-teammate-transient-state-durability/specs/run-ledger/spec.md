## ADDED Requirements

### Requirement: Teammate approval-blocked durability mirror
The system SHALL durably mirror built-in teammate approval-blocked state into RunLedger. The durable mirror SHALL cover `runtime_condition`, `blocked_reason`, and `grant_request_id` for approval-blocked teammate runs. This mirror uses best-effort semantics: live projection writes remain authoritative for runtime continuity, while journal plus snapshot state provide durable reconstruction.

#### Scenario: Approval-blocked teammate state is reconstructible
- **WHEN** a built-in teammate run enters `blocked_waiting_approval`
- **THEN** RunLedger SHALL append a durable approval-block journal event
- **AND** the RunLedger snapshot SHALL retain the latest blocked condition, blocked reason, and grant request ID

#### Scenario: Approval unblock clears durable blocked state
- **WHEN** a built-in teammate run leaves approval-blocked state
- **THEN** RunLedger SHALL append a durable approval-unblocked journal event
- **AND** the latest durable blocked snapshot fields SHALL be cleared

#### Scenario: Mirror failure does not fail-close runtime
- **WHEN** the durable mirror write fails
- **THEN** the live control-plane projection write SHALL still succeed
- **AND** the failure SHALL be observable through logs and metrics

#### Scenario: RunLedger disabled skips mirror silently
- **WHEN** RunLedger or write-through mirroring is disabled
- **THEN** approval-blocked mirroring SHALL be skipped
- **AND** the live control-plane projection SHALL remain the only state source
