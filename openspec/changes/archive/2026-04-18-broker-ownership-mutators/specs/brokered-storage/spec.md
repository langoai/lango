## ADDED Requirements

### Requirement: Payment mutators use storage-facing transaction capabilities
Production payment and settlement mutator paths MUST obtain payment transaction persistence and spending-limit collaborators through explicit storage-facing capabilities instead of extracting raw parent-side Ent clients.

#### Scenario: Payment CLI mutator setup avoids raw Ent access
- **WHEN** a payment CLI command initializes send, balance, or info dependencies
- **THEN** it obtains transaction persistence and spending-limit collaborators through storage-facing capabilities
- **AND** it does not extract a raw `*ent.Client` from the session store

#### Scenario: App payment and settlement setup avoids raw Ent access
- **WHEN** app wiring initializes payment service or P2P settlement persistence
- **THEN** it uses explicit storage-facing transaction capabilities
- **AND** it does not reconstruct those dependencies from raw parent-side ORM handles
