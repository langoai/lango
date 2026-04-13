## ADDED Requirements

### Requirement: Groth16 Solidity verifier export

The system SHALL provide a `cmd/zkexport` CLI tool that compiles gnark circuits and exports Groth16 verifying keys as Solidity contracts. The tool SHALL use unsafe SRS for R&D and support all registered circuit IDs.

#### Scenario: Export verifier for existing circuit
- **WHEN** `zkexport --circuit ownership --output contracts/src/verifiers/OwnershipVerifier.sol` is run
- **THEN** the tool compiles the circuit, runs Groth16 setup, and writes a Solidity verifier contract

#### Scenario: Exported verifier accepts valid proof
- **WHEN** a proof is generated off-chain for the ownership circuit and submitted to the exported verifier contract
- **THEN** the verifier returns true

### Requirement: PQ Attestation Circuit

The system SHALL provide a `PQAttestationCircuit` that proves an attestor observed specific PQ signature verification inputs and binds them to an on-chain action. The circuit uses MiMC hashes (consistent with existing circuits) and domain-separated public inputs for replay prevention.

#### Scenario: PQ attestation proof generated
- **WHEN** an attestor verifies an ML-DSA-65 signature off-chain and generates a PQ attestation proof
- **THEN** the proof SHALL bind attestor DID hash, message hash, PQ public key hash, timestamp, deal ID, chain ID, and contract address as public inputs

#### Scenario: Domain binding prevents replay
- **WHEN** a valid PQ attestation proof for deal D1 on chain C1 is submitted for deal D2 on chain C2
- **THEN** the on-chain verifier SHALL reject it because the public inputs do not match

#### Scenario: Attestor binding prevents impersonation
- **WHEN** a PQ attestation proof is generated
- **THEN** the circuit SHALL prove `MiMC(AttestorSecret) == AttestorDIDHash`
- **AND** only the holder of the attestor secret can produce a valid proof

## MODIFIED Requirements

### Requirement: ProverService

#### Scenario: Groth16 verifier export method
- **WHEN** `ExportGroth16Verifier(circuitID, w)` is called
- **THEN** the service compiles the circuit, runs Groth16 setup with unsafe SRS, and writes the Solidity verifier to `w`
