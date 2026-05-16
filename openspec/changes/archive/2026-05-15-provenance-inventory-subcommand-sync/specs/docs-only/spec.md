## ADDED Requirements

### Requirement: Provenance inventory docs use current subcommand slices
The architecture inventory and README internal tree SHALL describe the current provenance subcommand slices rather than only broad family names.

#### Scenario: Broad provenance shorthand stays removed
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** they SHALL describe the current checkpoint, session, attribution, and bundle subcommand slices
