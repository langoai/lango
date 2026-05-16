## MODIFIED Requirements

### Requirement: Groth16 Solidity verifier export
The system SHALL provide a `cmd/zkexport` CLI tool that compiles gnark circuits and exports Groth16 verifying keys as Solidity contracts. The tool SHALL use unsafe SRS for R&D and support all registered circuit IDs.

#### Scenario: Failed export removes partial output
- **WHEN** verifier export fails after the output file has been created
- **THEN** `zkexport` SHALL remove the partial output file
- **AND** the failure message SHALL be written to stderr
