## ADDED Requirements

### Requirement: p2p_invoke_paid tool registration
The system SHALL register a `p2p_invoke_paid` agent tool when P2P is enabled, wallet is available, spending limiter is configured, and USDC contract is resolvable for the configured chain ID. The tool SHALL have safety level `dangerous`.

#### Scenario: Tool registered with valid payment components
- **WHEN** the application initializes with `p2p.enabled=true`, a wallet provider, a spending limiter, and a valid chain ID
- **THEN** `buildP2PPaidInvokeTool` SHALL return a tool slice containing the `p2p_invoke_paid` tool

#### Scenario: Tool not registered without wallet
- **WHEN** the application initializes with `paymentComponents` having nil wallet or nil limiter
- **THEN** `buildP2PPaidInvokeTool` SHALL return nil

#### Scenario: Tool not registered for unsupported chain
- **WHEN** `contracts.LookupUSDC(chainID)` returns an error for the configured chain ID
- **THEN** `buildP2PPaidInvokeTool` SHALL log a warning and return nil

### Requirement: Session verification before invocation
The tool SHALL verify an active session exists for the specified peer DID before proceeding with any price query or invocation.

#### Scenario: No active session
- **WHEN** `p2p_invoke_paid` is called with a `peer_did` that has no active session
- **THEN** the tool SHALL return an error containing "no active session for peer"

#### Scenario: Missing required parameters
- **WHEN** `p2p_invoke_paid` is called without `peer_did` or `tool_name`
- **THEN** the tool SHALL return an error containing "peer_did and tool_name are required"

### Requirement: Automatic price query and free tool routing
The tool SHALL query the remote peer's price for the specified tool and invoke it directly (without payment) if the tool is free.

#### Scenario: Free tool invocation
- **WHEN** the remote peer reports `isFree=true` for the requested tool
- **THEN** the tool SHALL call `InvokeTool()` (not `InvokeToolPaid()`) and return the result with `paid=false`

### Requirement: Spending limit enforcement
The tool SHALL check the payment amount against the `SpendingLimiter` before signing any EIP-3009 authorization.

#### Scenario: Amount exceeds per-transaction limit
- **WHEN** the tool price exceeds the configured per-transaction spending limit
- **THEN** the tool SHALL return an error from `limiter.Check()` without signing any authorization

#### Scenario: Amount exceeds daily limit
- **WHEN** the tool price plus today's spending exceeds the daily spending limit
- **THEN** the tool SHALL return an error from `limiter.Check()` without signing any authorization

### Requirement: Auto-approval threshold check
The tool SHALL use `SpendingLimiter.IsAutoApprovable()` to determine whether the payment can proceed automatically. If not auto-approvable, it SHALL return an `approval_required` status instead of failing.

#### Scenario: Amount below auto-approve threshold
- **WHEN** the tool price is at or below the auto-approve threshold and within spending limits
- **THEN** the tool SHALL proceed with EIP-3009 signing and invocation

#### Scenario: Amount above auto-approve threshold
- **WHEN** the tool price exceeds the auto-approve threshold
- **THEN** the tool SHALL return a result with `status=approval_required`, the tool name, price, currency, and a descriptive message

### Requirement: EIP-3009 authorization signing
The tool SHALL create and sign an EIP-3009 `transferWithAuthorization` using the buyer's wallet, the seller's address from the price quote, and the canonical USDC contract for the chain.

#### Scenario: Successful authorization signing
- **WHEN** the auto-approval check passes
- **THEN** the tool SHALL call `eip3009.NewUnsigned()` with the buyer address, seller address, amount, and a 10-minute deadline, then sign it with `eip3009.Sign()`

### Requirement: Authorization serialization compatibility
The `authToMap()` helper SHALL serialize an `eip3009.Authorization` into a `map[string]interface{}` format that is wire-compatible with `paygate.parseAuthorization()`.

#### Scenario: Field format matches paygate expectations
- **WHEN** `authToMap()` serializes an authorization
- **THEN** the output SHALL contain: `from` and `to` as hex address strings, `value`/`validAfter`/`validBefore` as decimal strings, `nonce`/`r`/`s` as `0x`-prefixed hex strings of 32 bytes, and `v` as a `float64`

### Requirement: Paid invocation and response handling
The tool SHALL call `InvokeToolPaid()` with the serialized authorization and handle the response based on its status.

#### Scenario: Successful paid invocation
- **WHEN** the remote peer returns `ResponseStatusOK`
- **THEN** the tool SHALL record the spending via `limiter.Record()` and return a result with `status=ok`, `paid=true`, the price, currency, and the remote tool result

#### Scenario: Payment rejected by seller
- **WHEN** the remote peer returns `ResponseStatusPaymentRequired`
- **THEN** the tool SHALL return a result with `status=payment_required` and a descriptive message

#### Scenario: Remote error
- **WHEN** the remote peer returns any other error status
- **THEN** the tool SHALL return an error with the remote error message
