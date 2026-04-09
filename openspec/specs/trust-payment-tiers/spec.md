## Purpose

Capability spec for trust-payment-tiers. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Post-pay for high-trust peers
The payment gate SHALL grant post-pay status to peers whose reputation score is strictly greater than the configured `PostPayMinScore` threshold (default: 0.8). Post-pay means the tool executes first and settlement occurs asynchronously afterward.

#### Scenario: High-trust peer gets post-pay
- **WHEN** a peer with trust score 0.9 invokes a paid tool
- **THEN** the gate returns `StatusPostPayApproved` with a non-empty `SettlementID`

#### Scenario: Score exactly at threshold stays prepay
- **WHEN** a peer with trust score exactly 0.8 invokes a paid tool
- **THEN** the gate returns `StatusPaymentRequired` (prepay path)

### Requirement: Fallback to prepay on reputation error
The payment gate SHALL fall back to prepay mode when the reputation lookup returns an error, rather than denying the request.

#### Scenario: Reputation service unavailable
- **WHEN** the reputation function returns an error
- **THEN** the gate proceeds with standard prepay logic (requires EIP-3009 authorization)

### Requirement: Nil reputation function defaults to prepay
The payment gate SHALL use prepay for all peers when no reputation function is configured.

#### Scenario: No reputation function wired
- **WHEN** `ReputationFunc` is nil on the gate
- **THEN** all paid tool invocations require upfront EIP-3009 payment authorization

### Requirement: Free tools ignore reputation
The payment gate SHALL return `StatusFree` for free tools regardless of the peer's reputation score.

#### Scenario: Free tool with high-trust peer
- **WHEN** a peer with trust score 1.0 invokes a free tool
- **THEN** the gate returns `StatusFree` without consulting reputation

### Requirement: Deferred payment ledger tracks post-pay obligations
The gate SHALL maintain an in-memory deferred ledger that records post-pay obligations. Each entry tracks peer DID, tool name, price, creation time, and settlement status.

#### Scenario: Ledger records post-pay entry
- **WHEN** a post-pay is approved
- **THEN** a new `DeferredEntry` is added to the ledger with `Settled=false`

#### Scenario: Ledger concurrent access safety
- **WHEN** multiple goroutines add and settle entries concurrently
- **THEN** no data races occur and all entries are correctly tracked

### Requirement: Configurable trust threshold
The `PostPayMinScore` threshold SHALL be configurable via `P2PPricingConfig.TrustThresholds.PostPayMinScore`. When not set, it defaults to 0.8.

#### Scenario: Custom threshold lowers post-pay barrier
- **WHEN** `PostPayMinScore` is configured as 0.6 and a peer has score 0.7
- **THEN** the gate returns `StatusPostPayApproved`
