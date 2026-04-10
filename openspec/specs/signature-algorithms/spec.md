## Purpose

Capability spec for signature-algorithms. Defines reusable signature algorithm types, constants, and verification functions.

## Requirements

### Requirement: SignatureScheme type

The system SHALL provide a `SignatureScheme` struct in `internal/security` containing `ID` (algorithm identifier string), `Verify` (verification function), `SignatureSize`, and `PublicKeySize`. This type is a canonical algorithm descriptor, not a registry — actual dispatch is handled by consumer-specific injected verifier maps.

#### Scenario: SignatureScheme metadata is accurate
- **WHEN** `Secp256k1Keccak256Scheme` is accessed
- **THEN** `ID` SHALL be `"secp256k1-keccak256"`, `SignatureSize` SHALL be 65, `PublicKeySize` SHALL be 33

#### Scenario: Ed25519 metadata is accurate
- **WHEN** `Ed25519Scheme` is accessed
- **THEN** `ID` SHALL be `"ed25519"`, `SignatureSize` SHALL be 64, `PublicKeySize` SHALL be 32

### Requirement: Algorithm constants

The system SHALL define canonical algorithm identifier constants: `AlgorithmSecp256k1Keccak256 = "secp256k1-keccak256"` and `AlgorithmEd25519 = "ed25519"` in `internal/security`. These are the single source of truth for algorithm identifiers across the codebase.

### Requirement: Secp256k1Keccak256 verification

The `VerifySecp256k1Keccak256(publicKey, message, signature []byte) error` function SHALL hash the message with Keccak256, recover the public key from the 65-byte ECDSA signature, and compare with the claimed compressed public key.

#### Scenario: Valid secp256k1 signature accepted
- **WHEN** a valid secp256k1+keccak256 signature is verified with the correct public key
- **THEN** the function SHALL return nil

#### Scenario: Mismatched public key rejected
- **WHEN** the recovered public key does not match the claimed key
- **THEN** the function SHALL return an error

### Requirement: Ed25519 verification

The `VerifyEd25519(publicKey, message, signature []byte) error` function SHALL verify a 64-byte Ed25519 signature against a 32-byte public key using `crypto/ed25519.Verify`. Ed25519 is registered as a framework verification algorithm; it is not wired into production identity flows in Phase 2.

#### Scenario: Valid Ed25519 signature accepted
- **WHEN** a valid Ed25519 signature is verified with the correct public key
- **THEN** the function SHALL return nil

#### Scenario: Invalid Ed25519 signature rejected
- **WHEN** the signature does not match the message and key
- **THEN** the function SHALL return an error
