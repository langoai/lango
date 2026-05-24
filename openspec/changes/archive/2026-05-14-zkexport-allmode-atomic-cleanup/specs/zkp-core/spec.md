## MODIFIED Requirements

### Requirement: Groth16 Solidity verifier export
The system SHALL provide a `cmd/zkexport` CLI tool that compiles gnark circuits and exports Groth16 verifying keys as Solidity contracts. The tool SHALL use unsafe SRS for R&D and support all registered circuit IDs.

#### Scenario: All-mode failure removes earlier outputs from the same run
- **WHEN** `zkexport --all` successfully writes one or more verifier files and a later circuit export fails
- **THEN** the files created earlier in that same `--all` run SHALL be removed
