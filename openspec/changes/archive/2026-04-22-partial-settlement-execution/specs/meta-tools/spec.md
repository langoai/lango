## ADDED Requirements

### Requirement: Partial settlement execution meta tool
The meta tools surface SHALL provide an `execute_partial_settlement` tool that executes a direct partial settlement from canonical settlement progression state.

#### Scenario: Partial settlement execution tool available
- **WHEN** the meta tools are built with a receipts store and partial-settlement runtime
- **THEN** `execute_partial_settlement` SHALL be available

#### Scenario: Partial settlement execution tool executes direct partial settlement
- **WHEN** `execute_partial_settlement` is invoked with `transaction_receipt_id`
- **THEN** it SHALL evaluate the request through the partial settlement execution service
- **AND** it SHALL return canonical transaction-level execution result including settlement progression state, executed amount, remaining amount, and runtime reference
