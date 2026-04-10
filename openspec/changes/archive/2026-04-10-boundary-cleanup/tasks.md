# Tasks: boundary-cleanup

## Wave 1 — Interface extraction (parallel)

### Unit 1: Identity KeyProvider + bytes.Equal
- [ ] Define `KeyProvider` interface in `internal/p2p/identity/identity.go`
- [ ] Change `WalletDIDProvider.wallet` field to `keys KeyProvider`
- [ ] Change `NewProvider` parameter from `wallet.WalletProvider` to `KeyProvider`
- [ ] Remove `internal/wallet` import from `identity.go`
- [ ] Fix `signature.go`: `string()` comparison → `bytes.Equal`
- [ ] Add `import "bytes"` to `signature.go`
- [ ] Update `identity_test.go`: `mockWalletProvider` → `mockKeyProvider` (1 method)
- [ ] Remove `internal/wallet` import from `identity_test.go`
- [ ] Verify: `go build ./internal/p2p/identity/... && go test ./internal/p2p/identity/...`

### Unit 2: Handshake Signer + DIDFromPublicKey
- [ ] Define `Signer` interface in `internal/p2p/handshake/handshake.go`
- [ ] Change `Handshaker.wallet` field to `signer Signer`
- [ ] Change `Config.Wallet` field to `Config.Signer`
- [ ] Replace all `h.wallet.X` → `h.signer.X` (6 call sites)
- [ ] Replace inline DID construction (line 273) with `identity.DIDFromPublicKey(pubkey)`
- [ ] Add `internal/p2p/identity` import, remove `internal/wallet` import
- [ ] Update `handshake_test.go`: `mockWallet` → `mockSigner` (2 methods)
- [ ] Remove `internal/wallet` import from `handshake_test.go`
- [ ] Update `app/wiring_p2p.go:131`: `Config{Wallet: wp}` → `Config{Signer: wp}`
- [ ] Verify: `go build ./... && go test ./internal/p2p/handshake/... ./internal/app/...`

## Wave 2 — Provenance separation

### Unit 3: Provenance BundleSigner + verifier injection
- [ ] Define `BundleSigner` interface in `provenance/bundle.go`
- [ ] Define `SignatureVerifyFunc` type in `provenance/bundle.go`
- [ ] Export algorithm constant: `AlgorithmSecp256k1Keccak256`
- [ ] Add `verifiers map[string]SignatureVerifyFunc` to `BundleService`
- [ ] Update `NewBundleService` signature (add verifiers parameter)
- [ ] Update `Export`: `signFn BundleSignFunc` → `signer BundleSigner`
- [ ] Update `Export`: `bundle.SignatureAlgorithm = signer.Algorithm()`
- [ ] Update `Verify`: dispatch to `s.verifiers[bundle.SignatureAlgorithm]`
- [ ] Remove `internal/p2p/identity` import from `bundle.go`
- [ ] Update `bundle_test.go`: `testSigner` → return `BundleSigner`, update all NewBundleService/Export calls
- [ ] Update `app/modules_provenance.go:71`: add verifiers param
- [ ] Update `app/wiring_provenance.go:102`: create `walletBundleSigner`, inject verifiers
- [ ] Update `app/p2p_routes.go:217,308`: `provenanceSigner` return type + Export call
- [ ] Update `app/p2p_routes_test.go:222,330,351`: NewBundleService + Export
- [ ] Update `cli/provenance/common.go:52`: NewBundleService + loadSigner return type
- [ ] Update `cli/provenance/bundle.go:49`: Export call
- [ ] Verify: `go build ./... && go test ./internal/provenance/... ./internal/app/... ./internal/cli/provenance/...`

## Wave 3 — Verifier extraction

### Unit 4: Handshake ResponseVerifier
- [ ] Define `ResponseVerifyFunc` type in `handshake.go`
- [ ] Add `ResponseVerifier` field to `Config`
- [ ] Add `responseVerifier` field to `Handshaker`
- [ ] In `NewHandshaker`: default to `VerifySecp256k1Signature` if nil
- [ ] Create `verify.go`: extract `VerifySecp256k1Signature` and `verifyChallengeSignature`
- [ ] In `verifyResponse`: replace inline logic with `h.responseVerifier(pubkey, nonce, sig)`
- [ ] Verify: `go build ./internal/p2p/handshake/... && go test ./internal/p2p/handshake/...`

## Wave 4 — Enforcement + audit

### Unit 5: archtest boundary enforcement
- [ ] Fix `forbiddenForP2P`: remove trailing slashes
- [ ] Fix `forbiddenMatch`: use `dep == prefix || strings.HasPrefix(dep, prefix+"/")` pattern
- [ ] Add `internal/p2p/identity` to `p2pNetworkingPrefixes`
- [ ] Add `provenance → p2p/identity` forbidden rule
- [ ] Update depguard: remove p2p/handshake wallet exemption
- [ ] Verify: `go test ./internal/archtest/...`

### Unit 6: Downstream artifact audit
- [ ] Check `docs/features/provenance.md` for BundleSigner impact
- [ ] Check `docs/features/p2p-network.md` for handshake impact
- [ ] Check `docs/architecture/dependency-graph.md` for graph update
- [ ] Confirm `README.md` needs no changes (refactoring only)
- [ ] Update affected docs if needed
