# Spec: Security Fixes

## Purpose

Capability spec for security-fixes. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: SQL Injection prevention in dbmigrate

All SQLCipher PRAGMA statements that interpolate passphrase values must escape single quotes. Since PRAGMA doesn't support parameterized queries, an `escapePassphrase()` function must double single quotes.

**Scenarios:**
- Given passphrase `test'OR'1'='1`, when used in PRAGMA key, then it is escaped to `test''OR''1''=''1` preventing injection.

### REQ-2: Session key encryption must store actual ciphertext

`session.Manager.Create()` must store hex-encoded encrypted bytes in `PrivateKeyRef`, not discard them. `SignUserOp()` must decode the hex ciphertext and pass the key ID (not the ref) to the decrypt function.

**Scenarios:**
- Given encryption is enabled, when a session key is created, then `PrivateKeyRef` contains hex-encoded ciphertext (not a UUID).
- Given an encrypted session key, when `SignUserOp` is called, then the ciphertext is decoded and passed to the decrypt function with the correct key ID.

### REQ-3: P2P handshake must have default-deny approval

The handshaker's `ApprovalFn` must default to denying unknown peers. When `AutoApproveKnownPeers` is enabled and a reputation store is available, peers above the minimum trust score threshold are approved.

### REQ-4: ZK prover must sign challenges with wallet key

The ZK prover closure must call `wp.SignMessage(ctx, challenge)` to produce an ECDSA signature as the witness `Response`, not echo the challenge bytes.

### REQ-5: NonceCache must be lifecycle-managed

The `NonceCache` must be stored in `p2pComponents` and stopped during graceful shutdown to prevent goroutine leaks.
