## MODIFIED Requirements

### Requirement: Build Tag Parity
All Goreleaser build configurations SHALL include the same feature build tags as the Makefile and Dockerfile. Both `lango` and `lango-extended` builds MUST include the `fts5` and `vec` tags so that release binaries have full FTS5 and sqlite-vec support.

#### Scenario: Standard build includes vec tag
- **WHEN** Goreleaser builds the `lango` binary
- **THEN** the build tags MUST include both `fts5` and `vec`

#### Scenario: Extended build includes vec tag
- **WHEN** Goreleaser builds the `lango-extended` binary
- **THEN** the build tags MUST include `fts5`, `vec`, and `kms_all`

#### Scenario: Release binary supports vector operations
- **WHEN** a user installs a Goreleaser-built binary and enables embedding/RAG
- **THEN** `embedding.NewVectorStore()` SHALL return a functional `SQLiteVecStore` (not `ErrVecNotCompiled`)
