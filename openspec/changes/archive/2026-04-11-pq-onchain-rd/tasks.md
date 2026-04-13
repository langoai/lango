# Tasks: PQ On-chain R&D

## Wave 1 — ZK Verifier Export Tool + Solidity Contracts

- [x] Add `ExportGroth16Verifier(circuitID string, circuit, w io.Writer) error` to ProverService
- [x] Create `cmd/zkexport/main.go` CLI tool (compile circuit → Groth16 setup → export Solidity)
- [x] Export OwnershipVerifier.sol
- [x] Export AttestationVerifier.sol
- [x] Export BalanceVerifier.sol
- [x] Export CapabilityVerifier.sol
- [x] Test: Go-side verifier export for all 4 circuits (verified via `go run cmd/zkexport --all`)
- [x] Verify: `go build ./cmd/zkexport/... && go test ./internal/p2p/zkp/...`

## Wave 2 — PQ Attestation Circuit

- [x] Create `internal/p2p/zkp/circuits/pq_attestation.go` with domain-binding public inputs
- [x] Implement Define() with MiMC constraints + freshness + domain separators
- [x] Test: `pq_attestation_test.go` — valid proof, wrong attestor, expired timestamp, domain binding
- [x] Export `PQAttestationVerifier.sol` via zkexport tool
- [x] Register `pq_attestation` circuit in `cmd/zkexport/main.go`
- [x] Verify: `go test ./internal/p2p/zkp/circuits/... -run TestPQAttestation -v`

## Wave 3 — Escrow + ZK Proof Gate (Prototype)

- [x] Create `contracts/src/interfaces/IZKVerifier.sol` (Groth16 native ABI)
- [x] Create `contracts/src/prototype/LangoZKEscrow.sol` with domain binding + attestor allowlist
- [x] Verify: `cd contracts && forge build`

## Wave 4 — Research Report + Specs

- [x] Create `docs/research/phase7-pq-onchain-feasibility.md`
- [x] Document: ML-DSA-65 on EVM gas cost model (>50M gas native, infeasible)
- [x] Document: ZK-offload approach (~200k gas Groth16)
- [x] Document: Trust model (attestor-bound attestation, not PQ signature validity)
- [x] Document: Production requirements (trusted setup, proof aggregation, cross-chain)

## Final Verification

- [x] Full Go build: `go build ./... && go vet ./...`
- [x] Full Go test: `go test ./internal/p2p/zkp/... -count=1`
- [x] Full Foundry: `cd contracts && forge build`
