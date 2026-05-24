## MODIFIED Requirements

### Requirement: Groth16 Solidity verifier export
The system SHALL provide a `cmd/zkexport` CLI tool that compiles gnark circuits and exports Groth16 verifying keys as Solidity contracts. The tool SHALL use unsafe SRS for R&D and support all registered circuit IDs.

#### Scenario: Main wrapper forwards stderr and exit code through seams
- **WHEN** the `zkexport` main wrapper encounters a usage failure
- **THEN** it SHALL forward the usage text through the configured stderr seam
- **AND** it SHALL forward the non-zero exit code through the configured exit seam
