## ADDED Requirements

### Requirement: Exclusion-aware MkDocs build
The project SHALL configure `exclude_docs` so hidden docs, superpowers planning artifacts, and withdrawn cockpit sub-guides do not ship in the built site.

#### Scenario: Hidden docs do not appear in the public site
- **WHEN** `python3 -m mkdocs build --strict` is run from the project root
- **THEN** the excluded files SHALL be omitted from the generated site and SHALL not produce not-in-nav warnings

### Requirement: Public MkDocs IA exposes the chosen public surfaces
The MkDocs navigation SHALL surface the selected public Security pages and the top-level Research page while keeping the rest of the documentation tree hidden from the public site.

#### Scenario: Public security and research pages remain navigable
- **WHEN** the documentation site is built
- **THEN** the nav SHALL include only the intended public Security entries and the Research entry
