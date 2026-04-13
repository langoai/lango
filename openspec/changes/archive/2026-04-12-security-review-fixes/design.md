# Design: Security Review Fixes

## Approach

Targeted hardening fixes from 10 rounds of Codex code review. No new features — purely defensive corrections to identity binding, credential management, and trust boundaries.

## Key Decisions

### D1: Auth-before-cache in handshake
Bundle cache and alias registration happen only after signature verification + approval. Prevents forged bundles from poisoning `BundleResolver`/`DIDAlias`.

### D2: v2 DID requires bundle + signing key binding
v2 handshakes must include bundle. `ComputeDIDv2(bundle) == SenderDID` AND `bytes.Equal(PublicKey, Bundle.SigningKey.PublicKey)` enforced. Prevents bundle replay attacks.

### D3: LegacyDID not trusted without Proofs.Legacy
`bundle.LegacyDID` is a self-reported field. Without `Proofs.Legacy` (Phase 3 deferred), it cannot be used for session lookup, auto-approve, or gossip card DID matching. Known UX limitation: v1→v2 first handshake requires manual approval.

### D4: Keyring always updated on passphrase change
`change-passphrase` and `recovery restore` always attempt keyring Set (interactive terminal). Prevents stale keyring entries from breaking headless bootstrap.

### D5: Provenance uses wallet v1 DID
Provenance export uses `DIDFromPublicKey(wallet.PublicKey())` — v1 DID that `VerifyMessageSignature` can validate. Avoids v2 DID + secp256k1 signer mismatch.

### D6: ZK verifier pinned at deployment
`LangoZKEscrow.zkVerifier` is `immutable` — set in constructor, never caller-supplied. Prevents mock verifier attack.
