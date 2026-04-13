## ADDED Requirements

### Requirement: ZK-gated escrow prototype

The system SHALL provide a standalone `LangoZKEscrow.sol` contract (separate from `LangoEscrowHubV2`) that gates fund release on ZK proof verification. The contract uses Groth16 native ABI (`a/b/c/publicInputs`) and does NOT use `bytes` proof wrappers.

#### Scenario: Release with valid ZK proof
- **WHEN** a seller submits a valid ZK proof via `releaseWithProof` with matching domain binding
- **THEN** the contract verifies the proof via `IZKVerifier`, checks domain inputs (dealId, chainId, contractAddress), validates the attestor is trusted, and releases funds

#### Scenario: Domain binding mismatch rejected
- **WHEN** a ZK proof is submitted with `publicInputs[chainIdIdx] != block.chainid`
- **THEN** the contract SHALL revert

#### Scenario: Untrusted attestor rejected
- **WHEN** a ZK proof is submitted with an attestor DID hash not in `trustedAttestors`
- **THEN** the contract SHALL revert

#### Scenario: Invalid ZK proof rejected
- **WHEN** `IZKVerifier.verifyProof` returns false
- **THEN** the contract SHALL revert with `InvalidZKProof()`

### Requirement: IZKVerifier interface

The system SHALL define an `IZKVerifier` interface with `verifyProof(a, b, c, publicInputs) returns (bool)` using Groth16 native ABI encoding.

#### Scenario: Verifier interface matches gnark export
- **WHEN** a gnark Groth16 verifier is exported as Solidity
- **THEN** it SHALL implement `IZKVerifier` interface
