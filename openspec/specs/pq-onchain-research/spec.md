## Purpose

Feasibility research for post-quantum cryptography on EVM. Documents gas cost models, precompile proposals, and cross-chain PQ attestation patterns.

## Requirements

### Requirement: PQ on-chain feasibility report

The system SHALL provide a research document at `docs/research/phase7-pq-onchain-feasibility.md` covering ML-DSA-65 on EVM, ZK-offload gas costs, and production recommendations.

#### Scenario: Report covers gas cost comparison
- **WHEN** the feasibility report is reviewed
- **THEN** it SHALL include gas cost estimates for: native Solidity ML-DSA-65 (infeasible), Groth16 ZK-offload (~200k gas), and hypothetical EVM precompile

#### Scenario: Report covers trust model
- **WHEN** the feasibility report discusses attestor-bound attestation
- **THEN** it SHALL document that on-chain verifies attestation validity (not PQ signature validity) and identify the attestor as a trusted oracle

#### Scenario: Report covers production requirements
- **WHEN** the feasibility report discusses deployment
- **THEN** it SHALL document trusted setup ceremony requirements, proof aggregation opportunities, and cross-chain attestation patterns
