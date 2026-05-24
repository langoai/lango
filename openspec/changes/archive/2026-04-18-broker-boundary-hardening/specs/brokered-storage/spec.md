## ADDED Requirements

### Requirement: Production code uses capability-specific storage access
Production app and CLI code MUST consume storage through capability-specific facade methods instead of generic Ent/SQL handle access.

#### Scenario: CLI storage readers do not use generic ent accessors
- **WHEN** learning history, librarian inquiry inspection, workflow state, payment setup, or reputation inspection runs through the CLI
- **THEN** those code paths use storage-provided readers or factories
- **AND** they do not call generic `EntClient()` accessors from production code

#### Scenario: App wiring uses facade dependency bundles
- **WHEN** app initialization wires ontology, observability alerts, workflow state, or P2P reputation/settlement components
- **THEN** it resolves those dependencies from facade capability methods
- **AND** it does not reconstruct them from generic production ent/sql handles
