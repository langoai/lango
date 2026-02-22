## ADDED Requirements

### Requirement: p2p_pay Tool for Peer-to-Peer USDC Payment

The system SHALL expose a `p2p_pay` agent tool (safety level: `Dangerous`) that sends a USDC payment on the Base blockchain to a connected peer identified by their DID. The tool SHALL require `peer_did` and `amount` parameters and MAY accept an optional `memo`. The tool SHALL NOT be available if the payment service is not initialized.

#### Scenario: Successful payment to connected peer
- **WHEN** `p2p_pay` is called with a valid `peer_did` and `amount` for a peer with an active session
- **THEN** the tool SHALL submit a USDC transfer and return a receipt containing `txHash`, `from`, `to`, `peerDID`, `amount`, `currency`, `chainId`, `memo`, and `timestamp`

#### Scenario: Payment rejected when no active session
- **WHEN** `p2p_pay` is called with a `peer_did` for which no active session exists in the `SessionStore`
- **THEN** the tool SHALL return an error containing "no active session for peer" and SHALL NOT submit any transaction

#### Scenario: Missing required parameters rejected
- **WHEN** `p2p_pay` is called without `peer_did` or without `amount`
- **THEN** the tool SHALL return an error containing "peer_did and amount are required"

#### Scenario: Tool unavailable without payment service
- **WHEN** the application is initialized with `payment.enabled=false`
- **THEN** `buildP2PPaymentTool` SHALL return nil and `p2p_pay` SHALL NOT be registered with the agent

---

### Requirement: Recipient Address Derivation from DID

The `p2p_pay` tool SHALL derive the recipient's Ethereum wallet address from their DID by parsing the DID using `identity.ParseDID`, extracting the 33-byte compressed secp256k1 public key, and using the first 20 bytes as the Ethereum address (formatted as `0x<hex>`). An invalid or unparseable DID MUST cause the tool to return an error before any payment is attempted.

#### Scenario: Valid DID yields deterministic Ethereum address
- **WHEN** `p2p_pay` is called with `peer_did="did:lango:<33-byte-pubkey-hex>"`
- **THEN** the payment SHALL be sent to `0x<first-20-bytes-of-pubkey>` as the `To` address

#### Scenario: Unparseable DID returns error
- **WHEN** `p2p_pay` is called with `peer_did="invalid"` (no `did:lango:` prefix)
- **THEN** the tool SHALL return an error containing "parse peer DID"

---

### Requirement: P2P Requirement for Payment Feature

The P2P subsystem SHALL require `payment.enabled=true` at configuration validation time. If a user configures `p2p.enabled=true` without `payment.enabled=true`, the configuration loader MUST reject the configuration with an error containing "p2p requires payment.enabled (wallet needed for identity)". This enforces that a wallet is always present for DID derivation when P2P is active.

#### Scenario: P2P with payment enabled accepted
- **WHEN** the configuration has `p2p.enabled=true` and `payment.enabled=true`
- **THEN** configuration validation SHALL succeed

#### Scenario: P2P without payment rejected
- **WHEN** the configuration has `p2p.enabled=true` and `payment.enabled=false`
- **THEN** configuration validation SHALL return an error containing "p2p requires payment.enabled"

---

### Requirement: Default Payment Memo

When the `memo` parameter is not provided or is an empty string, the `p2p_pay` tool SHALL use `"P2P payment"` as the default memo value in the `PaymentRequest.Purpose` field.

#### Scenario: Empty memo defaults to "P2P payment"
- **WHEN** `p2p_pay` is called without a `memo` parameter
- **THEN** the `PaymentRequest.Purpose` field SHALL be `"P2P payment"`

#### Scenario: Provided memo is used as-is
- **WHEN** `p2p_pay` is called with `memo="service fee for code review"`
- **THEN** the `PaymentRequest.Purpose` field SHALL be `"service fee for code review"`

---

### Requirement: Spending Limit Enforcement on P2P Payments

P2P payments SHALL be subject to the same `SpendingLimiter` constraints as all other USDC transfers. The `payment.Service.Send` method SHALL check per-transaction and daily spending limits before submitting the transaction. If the payment would exceed any limit, `Send` SHALL return an error and no transaction SHALL be submitted.

#### Scenario: Payment within limits succeeds
- **WHEN** the requested amount is within both per-transaction and daily remaining limits
- **THEN** the payment SHALL be submitted and a receipt returned

#### Scenario: Payment exceeding per-transaction limit rejected
- **WHEN** the requested amount exceeds `maxPerTx`
- **THEN** `payment.Service.Send` SHALL return an error containing "exceeds per-transaction limit" and `p2p_pay` SHALL propagate it
