## ADDED Requirements

### Requirement: Builder Clone method
The prompt Builder SHALL provide a `Clone()` method that returns a deep copy of the builder, allowing independent modification without affecting the original.

#### Scenario: Clone produces independent copy
- **WHEN** a builder is cloned and the clone is modified (Add/Remove)
- **THEN** the original builder's sections SHALL remain unchanged

#### Scenario: Clone of empty builder
- **WHEN** an empty builder is cloned
- **THEN** the clone SHALL be a valid empty builder that can accept new sections

### Requirement: SectionAgentIdentity constant
The prompt package SHALL define a `SectionAgentIdentity` section ID with the value `"agent_identity"`, used for sub-agent role descriptions at priority 150.

#### Scenario: SectionAgentIdentity is distinct from SectionIdentity
- **WHEN** both SectionIdentity and SectionAgentIdentity are added to a builder
- **THEN** they SHALL coexist as separate sections (not replace each other)

### Requirement: LoadAgentFromDir function
The prompt package SHALL provide a `LoadAgentFromDir(base *Builder, dir string, logger)` function that overlays per-agent prompt overrides from a directory onto a cloned base builder.

#### Scenario: Known file mapping
- **WHEN** `IDENTITY.md` exists in the agent directory
- **THEN** it SHALL be loaded as SectionAgentIdentity with priority 150

#### Scenario: Empty or missing files ignored
- **WHEN** a `.md` file in the agent directory is empty or whitespace-only
- **THEN** it SHALL be ignored and the base section SHALL remain

#### Scenario: Non-existent directory returns base
- **WHEN** the agent directory does not exist
- **THEN** the original base builder SHALL be returned unmodified
