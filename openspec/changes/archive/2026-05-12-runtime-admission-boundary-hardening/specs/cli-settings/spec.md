## ADDED Requirements

### Requirement: Runtime admission settings
The settings surface SHALL expose runtime admission configuration under `ontology.governance.admissionMode`, `ontology.governance.learningDefaultConfidence`, and `ontology.governance.librarianDefaultConfidence`, with values `off` and `observe` for admission mode plus fallback confidence defaults of `0.60` for the learning producer group and `0.50` for the librarian producer group.

These fields SHALL use the existing `ontology.governance.*` config namespace for storage compatibility, but they SHALL always remain directly visible on the runtime admission settings surface rather than inheriting governance-enabled gating semantics.

#### Scenario: Runtime admission config is editable
- **WHEN** an operator edits runtime admission settings
- **THEN** the runtime admission mode SHALL be configurable as `off` or `observe`
- **AND** the learning producer group and librarian producer group SHALL each expose a fallback confidence default

#### Scenario: Runtime admission settings are not hidden behind governance-only gating
- **WHEN** the runtime admission settings surface is rendered
- **THEN** the runtime admission mode and both fallback confidence defaults SHALL remain directly editable within that settings surface
- **AND** the runtime SHALL NOT require a separate governance-enabled toggle before showing those fields

#### Scenario: No extra producer groups are implied
- **WHEN** the runtime admission settings surface is rendered
- **THEN** it SHALL scope fallback confidence defaults only to the learning producer group and the librarian producer group
- **AND** it SHALL NOT imply additional first-slice producer groups
