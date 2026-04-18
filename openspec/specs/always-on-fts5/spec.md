# always-on-fts5 Specification

## Purpose
TBD - created by archiving change always-on-fts5. Update Purpose after archive.
## Requirements
### Requirement: Default runtime includes FTS5 without a build tag
The standard Lango build and test workflow MUST include SQLite FTS5 support without requiring a dedicated `fts5` build tag.

#### Scenario: Default build uses FTS5-capable runtime
- **WHEN** a developer runs `go build ./...` or `go test ./...`
- **THEN** the resulting runtime includes the normal FTS5-capable SQLite configuration
- **AND** no `-tags "fts5"` argument is required

#### Scenario: Optional vec build remains separate
- **WHEN** a developer wants the legacy sqlite-vec integration
- **THEN** they enable it with the `vec` build tag
- **AND** FTS5 remains part of the default runtime rather than a separate tag

