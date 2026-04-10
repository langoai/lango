# Tasks: algorithm-agility

## Wave 1 — Foundation (parallel)

### Unit 1: SignatureScheme + algorithm constants + Verify implementations
- [ ] Create `internal/security/signature_scheme.go` with SignatureScheme struct + algorithm constants
- [ ] Create `internal/security/scheme_secp256k1.go` with VerifySecp256k1Keccak256 + Secp256k1Keccak256Scheme
- [ ] Create `internal/security/scheme_ed25519.go` with VerifyEd25519 + Ed25519Scheme
- [ ] Create `internal/security/signature_scheme_test.go` with secp256k1 + Ed25519 verify tests
- [ ] Verify: `go build ./internal/security/... && go test ./internal/security/...`

### Unit 2: Challenge double-hash bug fix
- [ ] Rename `challengeSignPayload` → `challengeCanonicalPayload` in `verify.go`, remove Keccak256
- [ ] Update `verifyChallengeSignature`: hash canonical payload with Keccak256 before SigToPub
- [ ] Update `handshake.go` Initiate: `challengeSignPayload` → `challengeCanonicalPayload`
- [ ] Add `TestVerifyChallengeSignature_Roundtrip` test
- [ ] Verify: `go build ./internal/p2p/handshake/... && go test ./internal/p2p/handshake/...`

### Unit 3: Identity ParseDIDPublicKey
- [ ] Add `ParseDIDPublicKey` to `identity.go`
- [ ] Add tests in `identity_test.go`
- [ ] Verify: `go build ./internal/p2p/identity/... && go test ./internal/p2p/identity/...`

## Wave 2 — Handshake algorithm awareness

### Unit 4: Signer.Algorithm() + verifier map + protocol fields
- [ ] Add `Algorithm() string` to Signer interface in `handshake.go`
- [ ] Rename `ResponseVerifyFunc` → `SignatureVerifyFunc` in `verify.go`
- [ ] Add `SignatureAlgorithm` field to Challenge and ChallengeResponse
- [ ] Replace `Config.ResponseVerifier` with `Config.Verifiers map[string]SignatureVerifyFunc`
- [ ] Replace `Handshaker.responseVerifier` with `verifiers map[string]SignatureVerifyFunc`
- [ ] Update `NewHandshaker`: default verifier map if nil
- [ ] Update Initiate/HandleIncoming: set SignatureAlgorithm from signer
- [ ] Update verifyResponse: dispatch by algorithm with secp256k1 default
- [ ] Convert verifyChallengeSignature to Handshaker method, dispatch by algorithm
- [ ] Update HandleIncoming: `verifyChallengeSignature` → `h.verifyChallengeSignature`
- [ ] Add `Algorithm()` to mockSigner in test
- [ ] Create `walletHandshakeSigner` wrapper in `wiring_p2p.go`
- [ ] Update Config construction: `Signer: &walletHandshakeSigner{wp: wp}`
- [ ] Add `TestVerifyResponse_Ed25519` integration test
- [ ] Verify: `go build ./... && go test ./internal/p2p/handshake/... ./internal/app/...`

## Wave 3 — Wiring + enforcement

### Unit 5: Provenance Ed25519 registration + constant migration
- [ ] Update `provenance/bundle.go`: re-export AlgorithmSecp256k1Keccak256 from security
- [ ] Update `app/modules_provenance.go`: use security constants + add Ed25519 wiring closure
- [ ] Update `app/wiring_provenance.go`: use security constants
- [ ] Update `cli/provenance/common.go`: use security constants + add Ed25519 wiring closure
- [ ] Update `app/p2p_routes_test.go`: verifier map with security constants
- [ ] Add `TestBundleService_ExportVerify_Ed25519` integration test in `bundle_test.go`
- [ ] Verify: `go build ./... && go test ./internal/provenance/... ./internal/app/... ./internal/cli/provenance/...`

### Unit 6: depguard update + downstream audit
- [ ] Add handshake/identity to depguard p2p-infra-no-economy files
- [ ] Remove NOTE comment about handshake wallet coupling
- [ ] Run `golangci-lint run` to verify
- [ ] Downstream audit: docs impact check, README no change needed
