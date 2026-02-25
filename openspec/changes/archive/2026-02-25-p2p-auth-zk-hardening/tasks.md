# Tasks: P2P Auth & ZK Hardening

## Phase 1: Foundation

- [x] **1A**: NonceCache with TTL-based eviction (`nonce_cache.go` + tests)
- [x] **1B**: ZK Circuit Test Suite (15 circuit tests + 6 prover tests)
- [x] **1C**: Fix AgentCapability circuit binding (AgentTestBinding public field)

## Phase 2: Core Protocol

- [x] **2A**: Signed Challenge (ECDSA over canonical payload, timestamp validation, nonce replay, dual protocol v1.0/v1.1)
- [x] **2B**: Attestation Timestamp Freshness (MinTimestamp/MaxTimestamp range assertions)
- [x] **2C**: AttestationData Wire Format & Verification (AttestationData struct, AttestationResult, ZKAttestVerifyFunc)

## Phase 3: Integration

- [x] **3A**: Wiring & Config Integration (NonceCache, dual protocol, attestation callback, SRS/MaxCredentialAge)
- [x] **3B**: Credential Revocation in Gossip (revokedDIDs, maxCredentialAge validation)
- [x] **3C**: SRS Production Path (SRSMode "file" support)
- [x] **3D**: Security Roadmap Grade Update (Auth B-→A, ZK C→B+, P3 items)

## Verification

- [x] `go build ./...` — build success
- [x] `go test ./internal/p2p/handshake/...` — 29 tests pass
- [x] `go test ./internal/zkp/circuits/...` — 15 circuit tests pass (plonk + groth16)
- [x] `go test ./internal/zkp/...` — 6 prover tests pass
- [x] `go vet ./...` — static analysis clean
