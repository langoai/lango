## MODIFIED Requirements

### Requirement: Groth16 Solidity verifier export
The system SHALL provide a `cmd/zkexport` CLI tool that compiles gnark circuits and exports Groth16 verifying keys as Solidity contracts. The tool SHALL use unsafe SRS for R&D and support all registered circuit IDs.

#### Scenario: Unknown circuit error includes deterministic available list
- **WHEN** `zkexport` is invoked with an unknown circuit ID
- **THEN** the failure message SHALL be written to stderr
- **AND** the available circuit list in that message SHALL use deterministic lexical order
