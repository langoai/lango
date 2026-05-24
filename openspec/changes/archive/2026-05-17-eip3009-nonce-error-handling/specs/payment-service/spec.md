## MODIFIED Requirements

### Requirement: EIP-3009 signing correctness

The EIP-3009 Sign function SHALL use SignTransaction (raw signing without additional hashing) instead of SignMessage (which applies keccak256) because TypedDataHash already returns a keccak256 digest. Unsigned EIP-3009 authorization construction SHALL require a full 32-byte cryptographically random nonce and SHALL return an error if nonce generation fails.

#### Scenario: EIP-3009 signature validity
- **WHEN** an EIP-3009 authorization is signed
- **THEN** the signature SHALL be verifiable by Verify() which uses crypto.Ecrecover on the TypedDataHash digest

#### Scenario: WalletSigner interface
- **WHEN** a wallet is used for EIP-3009 signing
- **THEN** it SHALL implement both SignTransaction (raw) and SignMessage (hashed) methods

#### Scenario: Nonce entropy failure is reported
- **WHEN** unsigned EIP-3009 authorization construction cannot read 32 random nonce bytes
- **THEN** construction SHALL return a non-nil error
- **AND** it SHALL NOT return an authorization with a zero or partial nonce
