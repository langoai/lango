# Design: PQ On-chain R&D

## Approach

Actionable R&D: deployable code + research documentation. ML-DSA-65 direct verification is infeasible on EVM, so we use an attestor-bound ZK attestation pattern: off-chain PQ signature verification → ZK proof of attestation → on-chain proof verification.

## Key Decisions

### D1: Groth16 for on-chain (not PlonK)
~200k gas vs ~350k gas. Groth16 requires trusted setup (unsafe SRS for R&D).

### D2: Attestor-bound trust model
On-chain verifies attestor-bound attestation validity, NOT PQ signature validity directly.
Security property: non-repudiation of attestation. Attestor = trusted oracle.

### D3: Domain-separated public inputs
`DealID`, `ChainID`, `ContractAddress` in circuit public inputs prevent replay.

### D4: Separate prototype (LangoZKEscrow.sol)
Do NOT modify LangoEscrowHubV2. Standalone contract isolates experiment from production.

### D5: Groth16 native ABI throughout
`(a, b, c, publicInputs)` format. No `bytes` wrapper.

## Architecture

```
Off-chain (attestor):
  1. Verify ML-DSA-65 signature (VerifyMLDSA65)
  2. Generate ZK proof (PQAttestationCircuit + Groth16 prover)
  3. Submit proof to on-chain escrow

On-chain (LangoZKEscrow):
  1. Check domain binding (dealId, chainId, contractAddress)
  2. Check attestor is trusted (allowlist)
  3. Verify ZK proof via IZKVerifier (Groth16 pairing check)
  4. Release funds on success
```

## Waves

1. **Wave 1**: ZK verifier export tool + 4 Solidity verifier contracts
2. **Wave 2**: PQ attestation circuit + 5th verifier
3. **Wave 3**: LangoZKEscrow prototype + IZKVerifier interface
4. **Wave 4**: Feasibility report + delta specs
