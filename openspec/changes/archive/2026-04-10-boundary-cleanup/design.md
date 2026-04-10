## Context

In Phase 1 (master-key-envelope), the storage encryption layer was separated into MK/KEK. To proceed with Phase 2 (Algorithm Agility) and Phase 3 (Hybrid DID v2), the coupling where the wallet secp256k1 key directly penetrates areas beyond settlement must first be resolved.

Current coupling structure:
- `p2p/handshake` → `wallet` (imports WalletProvider for signing)
- `p2p/identity` → `wallet` (imports WalletProvider for public key)
- `provenance` → `p2p/identity` (imports VerifyMessageSignature)
- `archtest`'s boundary rule misses wallet imports due to a trailing slash bug

## Goals / Non-Goals

**Goals:**
- Remove `internal/wallet` import from `internal/p2p/handshake/`
- Remove `internal/wallet` import from `internal/p2p/identity/`
- Remove `internal/p2p/identity` import from `internal/provenance/` (dependency inversion)
- Extract handshake response verification into an injectable function
- Strengthen archtest boundary enforcement

**Non-Goals:**
- Changing the `wallet.WalletProvider` interface
- Changing P2P protocol format (Challenge/ChallengeResponse)
- Changing DID format (did:lango:v2 is Phase 3)
- Adding new signature algorithms (Phase 2)
- Changing `internal/p2p/settlement/` dependencies (legitimate wallet dependency)

## Decisions

### D1: Consumer-local interface (not shared package)

Go idiom: interfaces are defined at the consumer. `identity` defines `KeyProvider` (1 method) and `handshake` defines `Signer` (2 methods) separately. No shared interface package is created.

**Alternative:** `internal/security/signer/` shared package → unnecessary abstraction layer (violates the no-dead-abstraction-layer rule)

### D2: Handshake uses identity.DIDFromPublicKey

After Unit 1 is complete, the `identity` package is wallet-free. The `handshake → identity` import is safe and eliminates duplicate inline DID assembly (`"did:lango:" + fmt.Sprintf("%x", pubkey)`).

**Alternative:** `types.DIDPrefix + hex.EncodeToString` → duplicates DID assembly logic, risks sync failure in future DID v2

### D3: Provenance verifier is injected only at wiring

`BundleService` receives `verifiers map[string]SignatureVerifyFunc` in its constructor. No default verifier is placed inside the provenance package. If the map is empty, Verify returns an "unsupported algorithm" error.

**Core principle:** Ownership of verification implementation belongs to the `app/cli` integration layer. The provenance package only defines types.

**Alternative:** Default verifier inside provenance → `p2p/identity` import remains, causing separation failure

### D4: ResponseVerifyFunc is optional in Config

If `Config.ResponseVerifier` is nil, `VerifySecp256k1Signature` is used as default. This pattern is already used by `ZKProverFunc`/`ZKVerifierFunc`.

### D5: BundleSigner interface (replacing callback)

Replace `BundleSignFunc func(ctx, payload) ([]byte, error)` with `BundleSigner interface { Sign(); Algorithm() }`. Since the signer provides the algorithm as its own property, hardcoding is eliminated.

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| NewBundleService signature change (11 call sites) | Compile error if missed | Full call site list documented in the plan + go build verification |
| Incorrect assumption about implicit interface satisfaction | Runtime failure | Compile-time check `var _ Signer = (*wallet.LocalWallet)(nil)` is unnecessary — wiring produces compile errors directly |
| archtest trailing slash fix breaks existing tests | CI failure | Strengthen archtest only after wallet imports are already removed from all p2p packages (Wave 4) |
| provenance default verifier regression | Separation boundary collapse | Enforce `provenance → p2p/identity` prohibition with archtest rule |
