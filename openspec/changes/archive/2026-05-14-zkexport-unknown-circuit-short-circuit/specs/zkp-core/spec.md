## MODIFIED Requirements

### Requirement: Groth16 Solidity verifier export
The system SHALL provide a `cmd/zkexport` CLI tool that compiles gnark circuits and exports Groth16 verifying keys as Solidity contracts. The tool SHALL use unsafe SRS for R&D and support all registered circuit IDs.

#### Scenario: Unknown circuit short-circuits before prover service setup
- **WHEN** `zkexport` is invoked with an unknown circuit ID
- **THEN** it SHALL fail before attempting prover service setup
