## Purpose

On-chain USDC settlement for the escrow engine. Converts DIDs to Ethereum addresses and executes USDC transfers using the agent wallet as custodian.

## Requirements

### Requirement: DID-to-Address resolution
The `ResolveAddress` function SHALL be provided via an `AddressResolver` interface with `ResolveAddress(did string) (common.Address, error)`. The `DefaultAddressResolver` SHALL dispatch by DID version: v1 DIDs (`did:lango:<hex>`) are resolved directly by decompressing the secp256k1 public key and deriving the Ethereum address. V2 DIDs (`did:lango:v2:<hash>`) are resolved via `SettlementKeyLookup` -> settlement key extraction -> address derivation. A backward-compatible package-level `ResolveAddress` function is retained for v1-only callers.

#### Scenario: v1 DID resolves directly
- **WHEN** `ResolveAddress("did:lango:<secp256k1-hex>")` is called
- **THEN** the resolver SHALL decompress the secp256k1 key and derive the Ethereum address

#### Scenario: v2 DID resolves via bundle
- **WHEN** `ResolveAddress("did:lango:v2:<hash>")` is called and the bundle is available
- **THEN** the resolver SHALL look up the IdentityBundle, extract the settlement key (secp256k1), and derive the Ethereum address

#### Scenario: v2 DID without bundle returns error
- **WHEN** `ResolveAddress("did:lango:v2:<hash>")` is called and no bundle is available
- **THEN** the resolver SHALL return an `ErrBundleNotFound` error

#### Scenario: Missing DID prefix returns error
- **WHEN** `ResolveAddress` is called with a string not prefixed with `did:lango:`
- **THEN** an `ErrInvalidDID` wrapped error is returned

#### Scenario: Invalid hex in DID returns error
- **WHEN** `ResolveAddress` is called with non-hex characters after the prefix
- **THEN** an `ErrInvalidDID` wrapped error is returned

#### Scenario: Invalid pubkey bytes returns error
- **WHEN** `ResolveAddress` is called with valid hex that is not a valid compressed pubkey
- **THEN** an `ErrInvalidDID` wrapped error is returned

### Requirement: USDC settler implements SettlementExecutor for on-chain transfers
The system SHALL provide `USDCSettler` implementing `SettlementExecutor`. `Lock` SHALL verify agent wallet USDC balance sufficiency. `Release` SHALL transfer USDC from agent wallet to seller address (resolved from DID). `Refund` SHALL transfer USDC from agent wallet to buyer address (resolved from DID).

#### Scenario: Lock verifies sufficient balance
- **WHEN** `Lock` is called and agent wallet USDC balance >= amount
- **THEN** no error is returned (balance check passes)

#### Scenario: Lock rejects insufficient balance
- **WHEN** `Lock` is called and agent wallet USDC balance < amount
- **THEN** an error indicating insufficient balance is returned

#### Scenario: Release transfers to seller
- **WHEN** `Release` is called with a valid seller DID and amount
- **THEN** a USDC transfer transaction is built, signed, submitted with retry, and confirmed

#### Scenario: Refund transfers to buyer
- **WHEN** `Refund` is called with a valid buyer DID and amount
- **THEN** a USDC transfer transaction is built, signed, submitted with retry, and confirmed

### Requirement: USDC settler uses functional options for configuration
The system SHALL support `WithReceiptTimeout`, `WithMaxRetries`, and `WithLogger` options. Default receipt timeout SHALL be 2 minutes. Default max retries SHALL be 3.

#### Scenario: Custom timeout option applied
- **WHEN** `NewUSDCSettler` is called with `WithReceiptTimeout(5 * time.Minute)`
- **THEN** the settler uses 5-minute receipt timeout

#### Scenario: Zero values ignored
- **WHEN** options with zero values are passed (e.g., `WithMaxRetries(0)`)
- **THEN** the default values are preserved
