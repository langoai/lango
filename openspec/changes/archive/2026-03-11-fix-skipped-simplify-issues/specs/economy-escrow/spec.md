## ADDED Requirements

### Requirement: Store ListByStatus query
The escrow `Store` interface SHALL provide a `ListByStatus(status EscrowStatus) []*EscrowEntry` method that returns only escrows matching the given status.

#### Scenario: Query pending escrows
- **WHEN** `ListByStatus(StatusPending)` is called on a store containing escrows in pending, funded, and active statuses
- **THEN** the result SHALL contain only escrows with `Status == StatusPending`

#### Scenario: No matching escrows
- **WHEN** `ListByStatus(StatusDisputed)` is called on a store with no disputed escrows
- **THEN** the result SHALL be an empty (or nil) slice

### Requirement: Exported NoopSettler type
The escrow package SHALL export a `NoopSettler` struct that implements `SettlementExecutor` with no-op operations. All packages requiring a placeholder settler SHALL use `escrow.NoopSettler{}` instead of defining local noop types.

#### Scenario: NoopSettler satisfies interface
- **WHEN** `escrow.NoopSettler{}` is used as a `SettlementExecutor`
- **THEN** `Lock`, `Release`, and `Refund` SHALL return nil without performing any operations

#### Scenario: Compile-time interface check
- **WHEN** the escrow package is compiled
- **THEN** a `var _ SettlementExecutor = (*NoopSettler)(nil)` check SHALL verify interface compliance
