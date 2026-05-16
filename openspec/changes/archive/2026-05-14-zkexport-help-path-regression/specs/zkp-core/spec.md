## MODIFIED Requirements

### Requirement: Groth16 Solidity verifier export
The system SHALL provide a `cmd/zkexport` CLI tool that compiles gnark circuits and exports Groth16 verifying keys as Solidity contracts. The tool SHALL use unsafe SRS for R&D and support all registered circuit IDs.

#### Scenario: Help path returns success without prover setup
- **WHEN** `zkexport --help` is invoked
- **THEN** the command SHALL return success
- **AND** it SHALL print help text to stderr
- **AND** it SHALL NOT attempt prover service setup
