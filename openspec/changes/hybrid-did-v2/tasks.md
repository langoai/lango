# Tasks: hybrid-did-v2

## Wave 1 — Foundation types (parallel)

### Unit 1: DID v2 types + IdentityBundle + ParseDID dispatcher
- [ ] Add `DIDv2Prefix = "did:lango:v2:"` to `internal/types/identity.go`
- [ ] Create `internal/p2p/identity/bundle.go`: IdentityBundle, PublicKeyEntry, BundleProofs, ComputeDIDv2, canonicalBundleBytes, BundleResolver interface, MemoryBundleCache, DIDAlias
- [ ] Add `Version int` field to DID struct in `identity.go`
- [ ] Refactor ParseDID → v1/v2 prefix dispatch
- [ ] parseDIDv2: Version=2, PublicKey=nil, PeerID="" (hollow parse)
- [ ] Update peerIDFromPublicKey: 33=secp256k1, 32=Ed25519
- [ ] ParseDIDPublicKey: reject v2 DID with clear error
- [ ] Tests for all new functions
- [ ] Verify: `go build ./... && go test ./internal/p2p/identity/...`

### Unit 2: Identity key derivation from MK
- [ ] Create `internal/security/derive_identity.go`: DeriveIdentityKey(mk, generation)
- [ ] Create `internal/security/derive_identity_test.go`: determinism, generation salt
- [ ] Verify: `go build ./internal/security/... && go test ./internal/security/...`

### Unit 3: Bundle file persistence
- [ ] Create `internal/p2p/identity/bundle_file.go`: BundleFilePath, HasBundleFile, LoadBundleFile, StoreBundleFile
- [ ] Create known-bundles persistence: StoreKnownBundle, LoadKnownBundle
- [ ] Create `internal/p2p/identity/bundle_file_test.go`
- [ ] Verify: `go test ./internal/p2p/identity/...`

## Wave 2 — BundleProvider

### Unit 4: BundleProvider + bundle creation
- [ ] Create `internal/p2p/identity/bundle_provider.go`: BundleProvider struct
- [ ] Implement LocalIdentityProvider: DID, PublicKey, SignMessage, Algorithm, DIDString, Bundle, LegacyDID
- [ ] Implement bundle creation flow (Ed25519 + wallet pubkey → bundle → dual proofs → store)
- [ ] Create `internal/p2p/identity/bundle_provider_test.go`
- [ ] Verify: `go test ./internal/p2p/identity/...`

## Wave 3 — Protocol + Economy (parallel)

### Unit 5: Handshake v2 + Signer.DID() + LegacySigner + Bundle transport
- [ ] Add `DID(ctx) (string, error)` to Signer interface in `handshake.go`
- [ ] Add `LegacySigner Signer` to Config
- [ ] Add `legacySigner` to Handshaker, `selectSigner(peerAlgo)` method
- [ ] Add `Bundle *identity.IdentityBundle` to Challenge and ChallengeResponse (omitempty)
- [ ] HandleIncoming: `identity.DIDFromPublicKey(pubkey)` → `signer.DID(ctx)`
- [ ] HandleIncoming: include bundle if signer is BundleProvider
- [ ] HandleIncoming: cache received challenge bundle in BundleResolver
- [ ] Initiate: include bundle, cache received response bundle
- [ ] Ed25519 in default verifier map
- [ ] Update mockSigner: add DID() method
- [ ] Update `walletHandshakeSigner`: add DID() method (returns v1 DID)
- [ ] Update `wiring_p2p.go`: pass LegacySigner + BundleResolver
- [ ] Tests: v2 handshake, v1 fallback, bundle exchange
- [ ] Verify: `go build ./... && go test ./internal/p2p/handshake/... ./internal/app/...`

### Unit 6: Economy/Escrow ResolveAddress v2
- [ ] Create `AddressResolver` interface in `address_resolver.go`
- [ ] Create `DefaultAddressResolver` with v1 direct + v2 bundle lookup
- [ ] Update `usdc_settler.go`: use AddressResolver interface (2 call sites)
- [ ] Wire AddressResolver with BundleResolver in app wiring
- [ ] Tests: v1 resolves, v2 resolves via bundle, v2 without bundle errors
- [ ] Verify: `go build ./... && go test ./internal/economy/escrow/...`

## Wave 4 — Bootstrap + App wiring

### Unit 7: Bootstrap identity phase + app wiring
- [ ] Add `phaseDeriveIdentityKey` to phases.go (after phaseInitCrypto)
- [ ] Add `IdentityKey ed25519.PrivateKey` to State and Result in pipeline.go
- [ ] Create BundleProvider in wiring_p2p.go (boot.IdentityKey + wallet pubkey)
- [ ] Wire BundleProvider as Signer, walletHandshakeSigner as LegacySigner
- [ ] Wire MemoryBundleCache as BundleResolver
- [ ] Wire DIDAlias into session/reputation lookups
- [ ] Change `p2pComponents.identity` type to LocalIdentityProvider interface
- [ ] Update wiring_provenance.go to use BundleProvider DID
- [ ] Update pipeline_test.go for 12 phases
- [ ] Verify: `go build ./... && go test ./internal/bootstrap/... ./internal/app/...`

## Wave 5 — Surface layer (parallel)

### Unit 8: GossipCard + AgentCard v2 DID
- [ ] Add optional `Bundle` field to GossipCard
- [ ] Set AgentCard.DID from BundleProvider DID
- [ ] Update gossip wiring to include bundle
- [ ] Verify: `go build ./... && go test ./internal/p2p/discovery/...`

### Unit 9: CLI status + VerifyMessageSignature v2
- [ ] Add identity bundle section to security status
- [ ] Update VerifyMessageSignature: reject v2 DID with clear error
- [ ] Update provenance verifier closures for v2 DID awareness (bundle lookup)
- [ ] Verify: `go test ./internal/cli/security/... ./internal/p2p/identity/...`

### Unit 10: Migration + docs + OpenSpec sync
- [ ] Auto-create IdentityBundle on first boot with MK
- [ ] Preserve legacy v1 DID in bundle.LegacyDID
- [ ] Update docs/security/encryption.md with identity bundle mention
- [ ] OpenSpec verify → sync → archive
