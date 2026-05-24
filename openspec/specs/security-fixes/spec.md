# Spec: Security Fixes

## Purpose

Capability spec for security-fixes. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: SQL Injection prevention in dbmigrate
All SQLCipher PRAGMA statements that interpolate passphrase values SHALL escape single quotes. Since PRAGMA does not support parameterized queries, an `escapePassphrase()` function SHALL double single quotes.

#### Scenario: Embedded single quotes are escaped in PRAGMA strings
- **WHEN** passphrase `test'OR'1'='1` is used in a PRAGMA key statement
- **THEN** it SHALL be escaped to `test''OR''1''=''1`
- **AND** SQL injection through the PRAGMA string SHALL be prevented

### REQ-2: Session key encryption must store actual ciphertext
`session.Manager.Create()` SHALL store hex-encoded encrypted bytes in `PrivateKeyRef` rather than discarding them. `SignUserOp()` SHALL decode the hex ciphertext and pass the key ID (not the ref) to the decrypt function.

#### Scenario: Session creation stores encrypted ciphertext reference
- **WHEN** encryption is enabled and a session key is created
- **THEN** `PrivateKeyRef` SHALL contain the hex-encoded ciphertext rather than a UUID placeholder

#### Scenario: SignUserOp decrypts ciphertext using the key ID
- **WHEN** `SignUserOp` is called for an encrypted session key
- **THEN** it SHALL decode the ciphertext
- **AND** pass the ciphertext and correct key ID to the decrypt function

### REQ-3: P2P handshake must have default-deny approval
The handshaker's `ApprovalFn` SHALL default to denying unknown peers. When `AutoApproveKnownPeers` is enabled and a reputation store is available, peers above the minimum trust score threshold SHALL be approved.

#### Scenario: Unknown peer is denied by default
- **WHEN** a peer has no qualifying reputation signal
- **THEN** the handshake approval path SHALL deny that peer by default

#### Scenario: Trusted peer can be auto-approved
- **WHEN** `AutoApproveKnownPeers` is enabled and a peer exceeds the minimum trust threshold
- **THEN** the handshake approval path SHALL auto-approve that peer

### REQ-4: ZK prover must sign challenges with wallet key
The ZK prover closure SHALL call `wp.SignMessage(ctx, challenge)` to produce an ECDSA signature as the witness `Response`, rather than echoing the challenge bytes.

#### Scenario: ZK prover returns a wallet signature
- **WHEN** the prover is asked to answer a challenge
- **THEN** it SHALL sign the challenge with the wallet key
- **AND** return the signature as the witness response

### REQ-5: NonceCache must be lifecycle-managed
The `NonceCache` SHALL be stored in `p2pComponents` and stopped during graceful shutdown to prevent goroutine leaks.

#### Scenario: NonceCache stops during shutdown
- **WHEN** the application shuts down gracefully
- **THEN** the stored `NonceCache` SHALL be stopped as part of lifecycle cleanup
