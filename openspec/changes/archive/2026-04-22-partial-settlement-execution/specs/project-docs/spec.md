## ADDED Requirements

### Requirement: Partial settlement execution page is published
The architecture docs SHALL include a dedicated `partial-settlement-execution.md` page for the first direct partial settlement execution slice.

#### Scenario: Partial settlement execution page exists
- **WHEN** a reader opens the architecture docs
- **THEN** they SHALL find the Partial Settlement Execution page

### Requirement: Architecture landing page links the partial settlement execution slice
The `docs/architecture/index.md` page SHALL include a quick link to `partial-settlement-execution.md` and a short summary that frames it as the first direct partial settlement execution slice for `knowledge exchange v1`.

#### Scenario: Partial settlement execution appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Partial Settlement Execution entry linking to `partial-settlement-execution.md`
- **AND** the entry SHALL describe the slice as direct partial settlement execution bounded by current implementation limits
