## ADDED Requirements

### Requirement: Docs home links every top-level section index
The public docs home page SHALL provide a section catalog that links every top-level docs section carrying its own `index.md`, so major documentation areas remain discoverable from the landing page.

#### Scenario: Top-level section indexes stay linked from docs home
- **WHEN** a maintainer updates `docs/index.md`
- **THEN** it SHALL include links to every top-level `docs/*/index.md` section index
- **AND** that catalog SHALL cover sections such as `getting-started/`, `architecture/`, `cli/`, `features/`, `security/`, `gateway/`, `payments/`, `automation/`, `deployment/`, and `development/`
