# Proposal: PQ On-chain R&D

## Why

ZK proofs are P2P-only and PQ signatures exist off-chain but can't be verified on-chain (ML-DSA-65 lattice arithmetic is infeasible in EVM/ZK circuits). On-chain settlement and escrow lack ZK proof gates. This R&D phase bridges the gap with deployable ZK verifier contracts, a PQ attestation circuit, and a prototype ZK-gated escrow.

## What Changes

- Export gnark Groth16 verifiers as Solidity contracts for all 4 existing circuits + new PQ attestation circuit
- New `PQAttestationCircuit` (hash-based commitment, attestor-bound — NOT full ML-DSA verification)
- Standalone `LangoZKEscrow.sol` prototype with ZK proof gate, domain binding, and attestor allowlist
- `IZKVerifier` interface for on-chain ZK proof verification
- Feasibility report: ML-DSA on EVM, gas models, precompile proposals
- ZK verifier export CLI tool (`cmd/zkexport/`)

## Capabilities

### Modified Capabilities
- `zkp-core` — Groth16 verifier export, PQ attestation circuit
- `onchain-escrow` — ZK proof gate (prototype only, Hub V2 unmodified)

### New Capabilities
- `pq-onchain-research` — Feasibility report for PQ on-chain verification
