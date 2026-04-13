# Tasks: Security Review Fixes

## Gossip Card Verification (Round 1-2, 8-11)

- [x] Add `ClassicalVerify` field to `GossipConfig` and `GossipService`
- [x] Wire `security.VerifyEd25519` as `ClassicalVerify` in `wiring_p2p.go`
- [x] Call `VerifyCardSignature` in `handleMessage` before storing card
- [x] Set `SignatureAlgorithm` before `CanonicalCardPayload` in `signCard`
- [x] Attach `bp.Bundle()` JSON to `localCard.Bundle` in `wiring_p2p.go`
- [x] Reject signed cards with bundle but no valid signing key
- [x] Skip classical + PQ verification for bundle-less cards (backward compat)
- [x] Enforce `card.DID == ComputeDIDv2(bundle)` only (remove LegacyDID match)

## Handshake Identity Binding (Round 1-2, 4-7, 10-11)

- [x] Move bundle cache after signature verification in `HandleIncoming`
- [x] Move bundle cache after `verifyResponse` in `Initiate`
- [x] Add v1 DID↔pubkey consistency check (skip for v2 prefix)
- [x] Require bundle for v2 DID handshakes
- [x] Verify `ComputeDIDv2(bundle) == SenderDID` for v2 peers
- [x] Verify `bytes.Equal(PublicKey, Bundle.SigningKey.PublicKey)` for v2 peers
- [x] Move `registerAlias` to after approval (prevent forged LegacyDID bypass)
- [x] Auto-approve: use existing alias if registered, raw DID otherwise
- [x] Use `lookupDID` (canonical) for `pending.PeerDID` in approval callback
- [x] Restore `selectSigner("")` for outbound initiation (mixed-version compat)
- [x] Remove `bundle.LegacyDID` session lookup (unverifiable without Proofs.Legacy)

## Bootstrap Phase Order (Round 1-3)

- [x] Swap `phaseLoadSecurityState` before `phaseMigrateEnvelope` in `DefaultPhases`
- [x] Load salt/checksum when envelope has `PendingMigration` or `PendingRekey`

## Status + Credential Management (Round 1-4, 8-11)

- [x] Pass `keyring.DetectSecureProvider()` to `readDBStatusNonInteractive`
- [x] Load active config from DB when MK available (`configstore.Store.LoadActive`)
- [x] Keyfile fallback on stale keyring passphrase (envelope path)
- [x] Keyfile fallback on stale keyring passphrase (legacy DB open path)
- [x] `change-passphrase`: update keyfile if exists + always attempt keyring Set
- [x] `recovery restore`: same credential sync pattern

## Provenance DID Alignment (Round 5)

- [x] `provenanceSigner`: use wallet v1 DID (`DIDFromPublicKey(wp.PublicKey)`)
- [x] `wiring_provenance.go` Exporter: same wallet v1 DID pattern

## Economy Resolver (Round 4)

- [x] Add `AddressResolver` parameter to `selectSettler`
- [x] Pass `NewDefaultAddressResolver(nil)` when p2p components available

## ZK Escrow Verifier (Round 4)

- [x] Pin `zkVerifier` as `immutable` in `LangoZKEscrow` constructor
- [x] Remove `verifier` parameter from `releaseWithProof`
- [x] Update `LangoZKEscrow.t.sol` tests for pinned verifier

## Verification

- [x] `go build ./... && go vet ./...`
- [x] `go test ./internal/p2p/handshake/... ./internal/p2p/discovery/... ./internal/cli/security/... ./internal/bootstrap/... ./internal/app/...`
- [x] `cd contracts && forge build && forge test --match-contract LangoZKEscrow`
