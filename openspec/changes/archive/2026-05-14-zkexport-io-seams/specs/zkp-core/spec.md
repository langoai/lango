## MODIFIED Requirements

### Requirement: Groth16 Solidity verifier export
The system SHALL provide a `cmd/zkexport` CLI tool that compiles gnark circuits and exports Groth16 verifying keys as Solidity contracts. The tool SHALL use unsafe SRS for R&D and support all registered circuit IDs.

#### Scenario: Export success writes progress to stdout
- **WHEN** `zkexport` successfully exports one or more verifier contracts
- **THEN** progress and summary output SHALL be written to stdout

#### Scenario: Usage and export failures write to stderr
- **WHEN** `zkexport` is invoked with missing required flags or export setup fails
- **THEN** the usage or error message SHALL be written to stderr
