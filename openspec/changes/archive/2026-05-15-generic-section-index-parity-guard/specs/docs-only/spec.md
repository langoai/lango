## ADDED Requirements

### Requirement: Every docs section index links its own dedicated pages
Each public docs section that ships a local `index.md` SHALL also link every dedicated Markdown page in that same section directory, so section-level navigation remains complete as the docs tree grows.

#### Scenario: Section indexes stay complete
- **WHEN** a maintainer updates a section landing page such as `docs/architecture/index.md`, `docs/cli/index.md`, `docs/features/index.md`, `docs/security/index.md`, `docs/automation/index.md`, `docs/payments/index.md`, `docs/gateway/index.md`, `docs/deployment/index.md`, `docs/development/index.md`, or `docs/getting-started/index.md`
- **THEN** that section index SHALL include links to every sibling `*.md` page in the same directory other than `index.md`
- **AND** this completeness rule SHALL apply generically across all public section indexes under `docs/`
