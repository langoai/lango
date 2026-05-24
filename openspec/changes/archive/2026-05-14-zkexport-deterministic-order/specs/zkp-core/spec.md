## MODIFIED Requirements

### Requirement: Groth16 Solidity verifier export
The system SHALL provide a `cmd/zkexport` CLI tool that compiles gnark circuits and exports Groth16 verifying keys as Solidity contracts. The tool SHALL use unsafe SRS for R&D and support all registered circuit IDs.

#### Scenario: Circuit listing is deterministic
- **WHEN** `zkexport` renders its available circuit list
- **THEN** the circuit IDs SHALL be listed in deterministic lexical order

#### Scenario: All-mode export progress is deterministic
- **WHEN** `zkexport --all` exports verifier contracts
- **THEN** the per-circuit progress output SHALL follow the same deterministic lexical circuit order
