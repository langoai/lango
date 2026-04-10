## Why

Handshake challenge/response verification and identity signature verification are hardcoded to secp256k1+keccak256. Phase 0 removed the wallet dependency, but the algorithm itself is still fixed. To add new algorithms in Phase 3 (DID v2) and Phase 5 (PQ signatures), a framework for making signature algorithms swappable is needed first.

Additional discovery: A double-hash bug in challenge signatures causes signed challenge verification to always fail. This is fixed in the same change.

## What Changes

- Add `SignatureScheme` type + algorithm constants (`AlgorithmSecp256k1Keccak256`, `AlgorithmEd25519`) + Verify function implementations to `internal/security/`
- `internal/p2p/handshake/`: Add `Algorithm()` to `Signer` interface, add `SignatureAlgorithm` field to `Challenge`/`ChallengeResponse` (backward compat), introduce verifier map dispatch, fix challenge double-hash bug
- `internal/p2p/identity/`: Add `ParseDIDPublicKey` function (extract pubkey without peerID)
- `internal/provenance/` + wiring: Register Ed25519 verifier as wiring closure (for framework validation, not a production feature)
- `.golangci.yml`: Add handshake/identity to p2p-infra-no-economy rule
- Move `provenance.AlgorithmSecp256k1Keccak256` → `security.AlgorithmSecp256k1Keccak256` as canonical source

**`did:lango` format remains secp256k1-only. Ed25519 is framework validation through integration tests only.**

## Capabilities

### New Capabilities

- `signature-algorithms`: SignatureScheme type, algorithm constants, secp256k1-keccak256 and Ed25519 Verify functions

### Modified Capabilities

- `p2p-handshake`: Algorithm() added to Signer interface, SignatureAlgorithm field added to Challenge/ChallengeResponse, verifier map dispatch, challenge double-hash bug fix
- `p2p-identity`: ParseDIDPublicKey function added (DID format/peerID unchanged)
- `session-provenance`: Ed25519 verifier registered (wiring closure, for framework validation)
- `architecture-boundary-enforcement`: handshake/identity added to depguard, outdated NOTE removed

## Impact

- **Code:** `internal/security/`, `internal/p2p/handshake/`, `internal/p2p/identity/`, `internal/provenance/`, `internal/app/`, `internal/cli/provenance/`, `.golangci.yml`
- **Protocol:** omitempty field added to Challenge/ChallengeResponse (backward compat). Compatible with existing peers.
- **API:** Algorithm() added to Signer interface — existing WalletProvider's implicit satisfaction broken → wiring wrapper needed
- **Dependencies:** No new external dependencies (`crypto/ed25519` is Go stdlib)
