## Context

In Phase 0, wallet dependency removal and provenance verifier injection were completed. Currently, the signature algorithm in handshake/identity is hardcoded to secp256k1+keccak256. A double-hash bug exists in challenge signatures.

## Goals / Non-Goals

**Goals:**
- Framework for making signature algorithms swappable
- Register Ed25519 as the second algorithm to validate the framework
- Fix challenge double-hash bug

**Non-Goals:**
- `did:lango` format extension (Phase 3)
- ML-DSA/PQC algorithms (Phase 5)
- KEM provider (Phase 4)
- SchemeRegistry singleton (constants + injection are sufficient)
- Ed25519 production wiring (test-only for framework validation)

## Decisions

### D1: SignatureScheme is a canonical descriptor, not a registry

`SignatureScheme` struct holds algorithm metadata (ID, Verify, sizes). Actual dispatch is handled by each consumer's injected verifier map. No shared registry singleton.

**Alternative:** SchemeRegistry → risk of registration state drift between consumers, but since there are currently only 2 consumers (handshake, provenance) each using different verifier signatures, a unified registry would actually increase adapter complexity.

### D2: Ed25519 verifier is a wiring closure

`identity.ParseDIDPublicKey(didStr) → security.VerifyEd25519(pubkey, msg, sig)` closure assembled at app/cli wiring. Ed25519 verifier is not placed in the identity package — doing so would implicitly allow `did:lango:<ed25519-pubkey>`.

### D3: Add Algorithm() to Signer interface

handshake Signer: `SignMessage + PublicKey + Algorithm`. WalletProvider's implicit satisfaction is broken → `walletHandshakeSigner` wrapper (same pattern as walletBundleSigner).

### D4: SignatureAlgorithm field in Challenge/ChallengeResponse

`json:"signatureAlgorithm,omitempty"` — empty = legacy secp256k1 (backward compat). Previous peers don't send this field, so it deserializes as empty → secp256k1 default.

### D5: ResponseVerifyFunc → SignatureVerifyFunc rename

Same function type used for both challenge and response verification. "Response" prefix is misleading.

### D6: challengeSignPayload → challengeCanonicalPayload (bug fix)

Before: `Keccak256(canonical)` returned → wallet.SignMessage hashes again → double-hash signature / single-hash verification mismatch.
After: Return raw canonical bytes, verifier handles hashing. Signing/verification at the same depth.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Signer.Algorithm() addition → WalletProvider implicit satisfaction broken | walletHandshakeSigner wrapper (existing pattern) |
| Ed25519 verifier could get connected to production | verifier registration in tests only, wiring is secp256k1 only |
| Constant relocation (provenance → security) → call site changes | re-export for backward compat |
| challenge double-hash fix → old peer compatibility | Existing challenge verification was already broken, so no regression |
