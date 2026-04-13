## Why

`internal/p2p/handshake/` and `internal/p2p/identity/` directly import `internal/wallet`, depending on `WalletProvider` (5 methods) designed for settlement. In practice, only `PublicKey()` and `SignMessage()` (2 methods) are used. `internal/provenance/` also directly imports `internal/p2p/identity`, creating a hardcoded dependency on `VerifyMessageSignature`. This coupling is a structural barrier blocking Phase 2 (Algorithm Agility) and Phase 3 (Hybrid DID v2).

## What Changes

- `internal/p2p/identity/`: Replace `wallet.WalletProvider` dependency with a 1-method `KeyProvider` interface. Remove wallet import.
- `internal/p2p/handshake/`: Replace `wallet.WalletProvider` dependency with a 2-method `Signer` interface. Replace inline DID assembly with `identity.DIDFromPublicKey`. Remove wallet import.
- `internal/p2p/handshake/`: Extract response verification logic into an injectable `ResponseVerifyFunc` to prepare for Phase 2 algorithm agility.
- `internal/provenance/`: Replace `BundleSignFunc` callback with `BundleSigner` interface. The signer provides the signature algorithm. Inject verifier map at wiring. Remove `internal/p2p/identity` import.
- `internal/p2p/identity/signature.go`: Fix `string()` comparison to use `bytes.Equal` (bug fix).
- `internal/archtest/`: Fix `forbiddenForP2P` trailing slash bug. Add `p2p/identity` to networking prefixes. Add provenance → p2p/identity prohibition rule.

## Capabilities

### New Capabilities

(None — this change cleans up internal boundaries of existing capabilities)

### Modified Capabilities

- `p2p-identity`: `WalletDIDProvider` changed to accept consumer-local `KeyProvider` interface instead of `wallet.WalletProvider`. Bug fix for byte comparison in `VerifyMessageSignature`.
- `p2p-handshake`: `Handshaker` changed to accept consumer-local `Signer` interface instead of `wallet.WalletProvider`. Replaced inline DID assembly with `identity.DIDFromPublicKey`. Extracted response verification into injectable `ResponseVerifyFunc`.
- `session-provenance`: `BundleService.Export` changed to accept `BundleSigner` interface instead of `BundleSignFunc`. `BundleService.Verify` uses injected verifier map. Removed `p2p/identity` import from provenance package (dependency inversion).
- `architecture-boundary-enforcement`: Added p2p/identity to archtest, added provenance → p2p/identity prohibition rule, fixed trailing slash matching bug.

## Impact

- **Code:** `internal/p2p/identity/`, `internal/p2p/handshake/`, `internal/provenance/`, `internal/app/`, `internal/cli/provenance/`, `internal/archtest/`
- **API:** `NewBundleService`, `Export`, `provenanceSigner` signature changes (internal only, no external API impact)
- **Protocol:** P2P handshake Challenge/ChallengeResponse structure unchanged. DID format unchanged.
- **Dependencies:** No new external dependencies.
- **Backward compatibility:** `wallet.WalletProvider` implicitly satisfies both `KeyProvider` and `Signer`, so wiring code works without type conversion.
