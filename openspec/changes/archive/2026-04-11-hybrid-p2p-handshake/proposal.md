## Why

The current handshake protocol (v1.1) authenticates peers via signatures but does not perform key exchange. Session tokens are HMAC-SHA256 over random bytes with no session encryption key derived. A future quantum computer could break the classical Diffie-Hellman or ECDH assumptions underlying libp2p's Noise transport. Adding a hybrid post-quantum KEM (X25519 + ML-KEM-768) to the handshake establishes quantum-resistant forward secrecy: even if long-term identity keys are compromised, past session keys cannot be recovered.

## What Changes

- Add hybrid KEM (X25519-MLKEM768 via `cloudflare/circl`) to the handshake protocol as v1.2
- Challenge and ChallengeResponse structs gain optional KEM fields (`kemPublicKey`, `kemAlgorithm`, `kemCiphertext`) with `omitempty` for backward compatibility
- Derive a 32-byte session encryption key via HKDF-SHA256 from the hybrid shared secret, anchored to the signing identity DIDs
- Store session key in `Session.EncryptionKey` (in-memory only, never serialized) with `KEMUsed` flag
- Bind KEM fields to handshake signatures (transcript binding) to prevent active tampering
- Add `PreferredProtocols()` helper for initiator protocol fallback (v1.2 → v1.1 → v1.0)
- Add `EnablePQHandshake` config field (default false, opt-in)
- Zero `EncryptionKey` at all session exit points including `Create()` overwrite
- Promote `cloudflare/circl` from indirect to direct dependency

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `p2p-handshake` — KEM fields in Challenge/ChallengeResponse, protocol v1.2, transcript binding, session key derivation, `PreferredProtocols()` helper
- `settings-p2p` — `EnablePQHandshake` config field
- `cli-security-status` — PQ KEM status display in `lango security status`

## Impact

- **Code**: `internal/security/` (new KEM wrapper + session key derivation), `internal/p2p/handshake/` (protocol v1.2, verify.go canonical payload extension, session.go key storage), `internal/config/types_p2p.go`, `internal/app/wiring_p2p.go`, `internal/app/tools_p2p.go` (initiator protocol selection), `internal/cli/security/status.go`
- **Dependencies**: `cloudflare/circl v1.6.1` promoted from indirect to direct
- **Wire format**: Backward compatible — new `omitempty` JSON fields ignored by v1.1 peers
- **No message encryption**: Session key is derived and stored but not used for payload encryption (Phase 4b scope)
