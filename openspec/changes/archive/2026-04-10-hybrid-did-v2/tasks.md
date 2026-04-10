# Tasks: hybrid-did-v2

## Wave 1 — Foundation types (parallel)

### Unit 1: DID v2 types + IdentityBundle + ParseDID dispatcher
- [x] Add `DIDv2Prefix = "did:lango:v2:"` to `internal/types/identity.go`
- [x] Create `internal/p2p/identity/bundle.go`: IdentityBundle, PublicKeyEntry, BundleProofs, ComputeDIDv2, canonicalBundleBytes, BundleResolver interface, MemoryBundleCache, DIDAlias
- [x] Add `Version int` field to DID struct in `identity.go`
- [x] Refactor ParseDID → v1/v2 prefix dispatch
- [x] parseDIDv2: Version=2, PublicKey=nil, PeerID="" (hollow parse)
- [x] Update peerIDFromPublicKey: 33=secp256k1, 32=Ed25519
- [x] ParseDIDPublicKey: reject v2 DID with clear error
- [x] Tests for all new functions
- [x] Verify: `go build ./... && go test ./internal/p2p/identity/...`

### Unit 2: Identity key derivation from MK
- [x] Create `internal/security/derive_identity.go`: DeriveIdentityKey(mk, generation)
- [x] Create `internal/security/derive_identity_test.go`: determinism, generation salt
- [x] Verify: `go build ./internal/security/... && go test ./internal/security/...`

### Unit 3: Bundle file persistence
- [x] Create `internal/p2p/identity/bundle_file.go`: BundleFilePath, HasBundleFile, LoadBundleFile, StoreBundleFile
- [x] Create known-bundles persistence: StoreKnownBundle, LoadKnownBundle
- [x] Create `internal/p2p/identity/bundle_file_test.go`
- [x] Verify: `go test ./internal/p2p/identity/...`

## Wave 2 — BundleProvider

### Unit 4: BundleProvider + bundle creation
- [x] Create `internal/p2p/identity/bundle_provider.go`: BundleProvider struct
- [x] Implement LocalIdentityProvider: DID, PublicKey, SignMessage, Algorithm, DIDString, Bundle, LegacyDID
- [x] Implement bundle creation flow (Ed25519 + wallet pubkey → bundle → dual proofs → store)
- [x] Create `internal/p2p/identity/bundle_provider_test.go`
- [x] Verify: `go test ./internal/p2p/identity/...`

## Wave 3 — Protocol + Economy (parallel)

### Unit 5: Handshake v2 + Signer.DID() + LegacySigner + Bundle transport
- [x] Add `DID(ctx) (string, error)` to Signer interface in `handshake.go`
- [x] Add `LegacySigner Signer` to Config
- [x] Add `legacySigner` to Handshaker, `selectSigner(peerAlgo)` method
- [x] Add `Bundle *identity.IdentityBundle` to Challenge and ChallengeResponse (omitempty)
- [x] HandleIncoming: `identity.DIDFromPublicKey(pubkey)` → `signer.DID(ctx)`
- [x] HandleIncoming: include bundle if signer is BundleProvider
- [x] HandleIncoming: cache received challenge bundle in BundleResolver
- [x] Initiate: include bundle, cache received response bundle
- [x] Ed25519 in default verifier map
- [x] Update mockSigner: add DID() method
- [x] Update `walletHandshakeSigner`: add DID() method (returns v1 DID)
- [x] Update `wiring_p2p.go`: pass LegacySigner + BundleResolver
- [x] Tests: v2 handshake, v1 fallback, bundle exchange
- [x] Verify: `go build ./... && go test ./internal/p2p/handshake/... ./internal/app/...`

### Unit 6: Economy/Escrow ResolveAddress v2
- [x] Create `AddressResolver` interface in `address_resolver.go`
- [x] Create `DefaultAddressResolver` with v1 direct + v2 bundle lookup
- [x] Update `usdc_settler.go`: use AddressResolver interface (2 call sites)
- [x] Wire AddressResolver with BundleResolver in app wiring
- [x] Tests: v1 resolves, v2 resolves via bundle, v2 without bundle errors
- [x] Verify: `go build ./... && go test ./internal/economy/escrow/...`

## Wave 4 — Bootstrap + App wiring

### Unit 7: Bootstrap identity phase + app wiring
- [x] Add `phaseDeriveIdentityKey` to phases.go (after phaseInitCrypto)
- [x] Add `IdentityKey ed25519.PrivateKey` to State and Result in pipeline.go
- [x] Create BundleProvider in wiring_p2p.go (boot.IdentityKey + wallet pubkey)
- [x] Wire BundleProvider as Signer, walletHandshakeSigner as LegacySigner
- [x] Wire MemoryBundleCache as BundleResolver
- [x] Wire DIDAlias into session/reputation lookups
- [x] Change `p2pComponents.identity` type to LocalIdentityProvider interface
- [x] Update wiring_provenance.go to use BundleProvider DID
- [x] Update pipeline_test.go for 12 phases
- [x] Verify: `go build ./... && go test ./internal/bootstrap/... ./internal/app/...`

## Wave 5 — Surface layer (parallel)

### Unit 8: GossipCard + AgentCard v2 DID
- [x] Add optional `Bundle` field to GossipCard
- [x] Set AgentCard.DID from BundleProvider DID
- [x] Update gossip wiring to include bundle
- [x] Verify: `go build ./... && go test ./internal/p2p/discovery/...`

### Unit 9: CLI status + VerifyMessageSignature v2
- [x] Add identity bundle section to security status
- [x] Update VerifyMessageSignature: reject v2 DID with clear error
- [x] Update provenance verifier closures for v2 DID awareness (bundle lookup)
- [x] Verify: `go test ./internal/cli/security/... ./internal/p2p/identity/...`

### Unit 10: Migration + docs + OpenSpec sync
- [x] Auto-create IdentityBundle on first boot with MK
- [x] Preserve legacy v1 DID in bundle.LegacyDID
- [x] Update docs/security/encryption.md with identity bundle mention
- [x] OpenSpec verify → sync → archive
