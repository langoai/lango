## MODIFIED Requirements

### Requirement: Hub client contract method invocation
The `HubClient` SHALL use `writeMethod` and `readMethod` helper methods to eliminate boilerplate in contract call methods. Each public method (Deposit, SubmitWork, Release, Refund, Dispute, ResolveDispute) SHALL delegate to these helpers.

#### Scenario: writeMethod wraps contract write call
- **WHEN** `Deposit(ctx, dealID)` is called
- **THEN** it delegates to `writeMethod(ctx, MethodDeposit, dealID)` which handles request construction and error wrapping

#### Scenario: readMethod wraps contract read call
- **WHEN** `NextDealID(ctx)` is called
- **THEN** it delegates to `readMethod(ctx, MethodNextDealID)` which handles request construction and error wrapping

#### Scenario: Error messages use method name
- **WHEN** a contract call fails
- **THEN** the error is wrapped as `"<methodName>: <underlying error>"` (e.g., `"deposit: rpc down"`)

### Requirement: Contract method name constants
The `hub` package SHALL define exported constants for all contract method names (V1 and V2). All client code SHALL reference these constants instead of string literals.

#### Scenario: Method constant usage
- **WHEN** `HubClient.Deposit` constructs a contract call
- **THEN** it uses `MethodDeposit` constant ("deposit") instead of a string literal

## ADDED Requirements

### Requirement: Transaction type constants
The escrow package SHALL define a `TransactionType` string type with constants `TxDeposit`, `TxRelease`, and `TxRefund`.

#### Scenario: Transaction type usage
- **WHEN** escrow store records a transaction
- **THEN** it uses `TxDeposit`/`TxRelease`/`TxRefund` constants instead of string literals
