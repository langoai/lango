# Phase 7: PQ On-Chain Feasibility Report

## Executive Summary

This report evaluates approaches for verifying post-quantum (ML-DSA-65) signatures on EVM-compatible blockchains. Direct on-chain verification is infeasible due to the computational complexity of lattice-based cryptography. We implement an **attestor-bound ZK attestation** pattern as a pragmatic alternative, achieving ~200k gas per verification.

## 1. ML-DSA-65 On EVM: Gas Cost Analysis

### 1.1 Native Solidity Implementation (Infeasible)

ML-DSA-65 (FIPS 204, NIST Level 3) verification involves:
- **NTT (Number Theoretic Transform)**: 256-element polynomial multiplication over Z_q (q=8380417)
- **Matrix-vector multiplication**: 6×5 matrix of 256-degree polynomials
- **Rejection sampling**: Hash-based coefficient generation
- **Modular arithmetic**: ~2000 field multiplications per NTT

**Estimated gas cost**: >50M gas (exceeds Ethereum block gas limit of 30M)
- Each NTT: ~5M gas (256 modular multiplications + butterfly operations)
- Matrix operations: ~30M gas (30 NTTs)
- Hash operations: ~10M gas (SHAKE-128/256 in Solidity)
- Comparison: ECDSA ecrecover = ~3k gas

**Verdict**: Not feasible. Even on L2s with higher gas limits, the cost is prohibitive.

### 1.2 EVM Precompile (Future)

A native EVM precompile for ML-DSA-65 would reduce verification to ~10-50k gas (comparable to ecrecover). This requires:
- EIP draft proposing `ML_DSA_VERIFY` precompile at a specific address
- Consensus across Ethereum clients (geth, Prysm, etc.)
- NIST PQC standards finalization (FIPS 204 published August 2024)

**Timeline**: 2-3 years minimum. No EIP exists as of April 2026.

### 1.3 ZK-Offload (Implemented — This Phase)

Use ZK proofs to attest PQ signature verification off-chain:
- **Groth16 verification**: ~200k gas (1 pairing check on BN254)
- **Proof size**: 128 bytes (compressed) or 256 bytes (uncompressed)
- **Public inputs**: 8 field elements (~256 bytes)

**Gas comparison**:

| Approach | Gas Cost | Feasibility |
|----------|----------|-------------|
| Native Solidity ML-DSA-65 | >50M | Infeasible |
| Hypothetical EVM precompile | ~10-50k | Future (2-3 years) |
| Groth16 ZK attestation | ~200k | **Implemented** |
| PlonK ZK attestation | ~350k | Available (not chosen) |
| ECDSA ecrecover (reference) | ~3k | Existing |

## 2. Trust Model: Attestor-Bound Attestation

### 2.1 What On-Chain Verifies

On-chain does **NOT** verify PQ signature validity. It verifies **attestor-bound attestation validity**:

1. The attestor (trusted oracle) performed ML-DSA-65 verification off-chain
2. The attestor produced a ZK proof binding the verification to specific inputs
3. The on-chain contract verifies the ZK proof + domain binding + attestor trust

### 2.2 Security Properties

- **Non-repudiation**: Attestor cannot deny having made the attestation (ZK proof binds attestor secret)
- **Replay prevention**: Domain binding (dealId, chainId, contractAddress) prevents cross-context replay
- **Freshness**: Timestamp constraint prevents stale attestations
- **Trust boundary**: Attestor is a trusted oracle — must be allowlisted on-chain

### 2.3 Limitations

- **Not trustless**: Relies on attestor honesty for PQ signature verification
- **Attestor compromise**: If attestor secret is compromised, false attestations possible
- **No signature non-repudiation**: The signer's ML-DSA key is committed but not verified on-chain

### 2.4 When to Use

This pattern is appropriate for:
- High-value escrow releases where PQ attestation adds defense-in-depth
- Multi-party deals where attestor is a neutral third party
- Migration period until EVM precompiles become available

## 3. Production Requirements

### 3.1 Trusted Setup Ceremony

Groth16 requires a per-circuit trusted setup (toxic waste must be destroyed):
- **R&D**: Uses gnark's unsafe SRS (acceptable for testing)
- **Production**: Requires a multi-party computation (MPC) ceremony
- Tools: gnark supports `gnark-groth16-ceremony` for production setups
- Each circuit change requires a new ceremony

### 3.2 Proof Aggregation (Future)

Batch N attestations into a single on-chain proof:
- **Recursive proofs**: Prove N proofs inside 1 outer proof (gnark supports recursion)
- **Gas savings**: 1 × 200k gas instead of N × 200k gas
- **Complexity**: Recursive circuit adds ~500k constraints
- **Recommended when**: N ≥ 5 attestations per transaction

### 3.3 Cross-Chain PQ Attestation

Patterns for multi-chain PQ attestation:

1. **Optimistic bridge**: Attestation accepted with fraud proof window. Challenge period: 7 days. Gas efficient but slow finality.
2. **Validator quorum**: M-of-N validators verify PQ sig off-chain, submit collective attestation on-chain. Gas: M signatures × 3k gas. Fast finality.
3. **ZK relay**: ZK proof of attestation generated on source chain, verified on destination chain. Gas: 200k on each chain. Most trust-minimized.

### 3.4 Verifier Contract Management

- Each circuit has a dedicated verifier contract (generated by `cmd/zkexport`)
- Verifier contracts are immutable (verifying key baked in at generation time)
- Circuit changes require redeployment of the verifier contract
- Consider: upgradeable proxy for verifier address in escrow contract

## 4. Implementation Summary

### Delivered in Phase 7

| Component | Status | Files |
|-----------|--------|-------|
| ZK Verifier Export Tool | Done | `cmd/zkexport/main.go` |
| 5 Groth16 Verifier Contracts | Done | `contracts/src/verifiers/*.sol` |
| PQ Attestation Circuit | Done | `internal/p2p/zkp/circuits/pq_attestation.go` |
| IZKVerifier Interface | Done | `contracts/src/interfaces/IZKVerifier.sol` |
| LangoZKEscrow Prototype | Done | `contracts/src/prototype/LangoZKEscrow.sol` |
| Feasibility Report | Done | This document |

### Deferred

| Component | Reason |
|-----------|--------|
| Full ML-DSA-65 ZK circuit | Millions of constraints, impractical prover time |
| EVM precompile implementation | Requires Ethereum core protocol change |
| Production trusted setup | Requires MPC ceremony infrastructure |
| Proof aggregation | Research only — needs recursive circuit design |
| Cross-chain bridge contracts | Architecture design only |

## 5. Recommendations

1. **Deploy ZK verifiers on Base Sepolia** for integration testing with existing escrow
2. **Design attestor selection protocol** — who becomes a trusted attestor and how
3. **Monitor EIP landscape** for ML-DSA precompile proposals (EIP-XXXX)
4. **Evaluate recursive proof** feasibility when dealing volume justifies aggregation
5. **Consider Poseidon hash** migration for gas optimization (~15% reduction over MiMC in circuit)
