## ADDED Requirements

### Requirement: Built-in teammate blocked-state durability
Built-in teammate approval-blocked state SHALL be durably reconstructible through the RunLedger mirror while the live runtime continues using the control-plane projection.

#### Scenario: Hard-cut audit closure is recorded by cross-reference
- **WHEN** the transient-state durability change completes
- **THEN** the archived hard-cut `approval-blocked conditions` follow-up SHALL be closed by cross-reference in the new change
- **AND** the archived `recovery states` follow-up SHALL remain open until a production writer exists
