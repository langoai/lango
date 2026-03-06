## ADDED Requirements

### Requirement: DID-to-Address resolver converts DID to Ethereum address
The system SHALL provide a `ResolveAddress(did string) (common.Address, error)` function that parses `did:lango:<hex-compressed-pubkey>`, hex-decodes the suffix, decompresses the secp256k1 public key via `crypto.DecompressPubkey`, and derives the Ethereum address via `crypto.PubkeyToAddress`.

#### Scenario: Valid DID resolves to address
- **WHEN** `ResolveAddress` is called with a valid `did:lango:<33-byte-hex-compressed-pubkey>`
- **THEN** the correct Ethereum address is returned

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
