# X402 V2 Protocol Specification

## Purpose
Define the X402 V2 payment protocol integration in Lango, using the Coinbase X402 Go SDK for automatic HTTP 402 payment handling with EIP-3009 off-chain signatures.

## Requirements

### Requirement: X402 V2 protocol integration documented
The X402 V2 payment protocol integration SHALL be documented through the Protocol Flow, Components, and Network Configuration sections below. This requirement exists to satisfy the canonical spec structure; detailed behavior contracts are in the descriptive sections that follow.

#### Scenario: Spec contains protocol documentation
- **WHEN** the X402 V2 spec.md is read
- **THEN** it SHALL include the protocol flow, components, and network configuration sections that describe the integration

### Requirement: Bounded X402 HTTP client timeout
The X402 interceptor SHALL create its wrapped HTTP client with a finite default timeout so automatic payment fetches cannot rely on an unbounded `http.Client`.

#### Scenario: X402 HTTP client has a default timeout
- **WHEN** `Interceptor.HTTPClient(ctx)` creates the wrapped payment client
- **THEN** the returned HTTP client SHALL have a non-zero timeout
- **AND** the timeout SHALL be at least 15 seconds

#### Scenario: Cached X402 HTTP client remains bounded
- **WHEN** `Interceptor.HTTPClient(ctx)` is called more than once
- **THEN** the cached client SHALL be reused
- **AND** the cached client SHALL retain the bounded timeout

## Protocol Flow
1. Agent makes HTTP request via `payment_x402_fetch` tool
2. Server returns 402 with `PAYMENT-REQUIRED` header (Base64 JSON)
3. SDK's `PaymentRoundTripper` intercepts the 402 response
4. SDK creates EIP-3009 authorization, signs with EIP-712 typed data
5. SDK retries request with `PAYMENT-SIGNATURE` header (Base64)
6. Server verifies signature, returns content with `PAYMENT-RESPONSE` header

## Components

### SignerProvider (`internal/x402/signer.go`)
- Interface: `SignerProvider` with `EvmSigner(ctx) (ClientEvmSigner, error)`
- Implementation: `LocalSignerProvider` loads key from SecretsStore under key `"wallet.privatekey"`
- Key material zeroed after signer creation

#### Scenario: Signer loads key using wallet convention
- **WHEN** `LocalSignerProvider.EvmSigner(ctx)` is called
- **THEN** the provider SHALL read the private key from SecretsStore using key name `"wallet.privatekey"`

#### Scenario: Key name matches wallet package
- **WHEN** a wallet is created via `CreateWallet()` and then x402 payment is attempted
- **THEN** the x402 signer SHALL successfully load the same key stored by the wallet package

### Config (`internal/x402/types.go`)
- `Config` struct: Enabled, ChainID, MaxAutoPayAmount
- `CAIP2Network(chainID)` helper: converts `84532` → `"eip155:84532"`

### Interceptor HTTP client wiring (`internal/x402/interceptor.go`)
- `HTTPClient(ctx)` creates the SDK-backed wrapped client with the exact EVM scheme registered

### Interceptor (`internal/x402/interceptor.go`)
- Thread-safe lazy initialization of wrapped `*http.Client`
- `BeforePaymentCreationHook` enforces spending limits
- `HTTPClient(ctx)` returns the X402-wrapped HTTP client
- `IsEnabled()` and `SignerAddress()` for callers

### payment_x402_fetch Tool (`internal/tools/payment/payment.go`)
- SafetyLevel: Dangerous (requires approval)
- Parameters: url (required), method, body, headers
- Uses interceptor's HTTPClient for automatic 402 handling
- Truncates response body at 8KB for agent context
- Records successful X402 payments for audit

### PaymentTx Schema (`internal/ent/schema/payment_tx.go`)
- `payment_method` enum: `direct_transfer` (default) | `x402_v2`
- Distinguishes explicit transfers from X402 auto-payments

## Network Configuration
- Chain ID: numeric (e.g., 84532)
- CAIP-2 format: `eip155:<chainID>` (used by SDK)
- Supported: Any EVM chain with EIP-3009 USDC contract
