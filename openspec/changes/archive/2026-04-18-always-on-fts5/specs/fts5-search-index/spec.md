## ADDED Requirements

### Requirement: FTS5 is part of the default supported runtime
FTS5 search support MUST be documented and verified as part of the default supported runtime instead of a dedicated `fts5` build-tag variant.

#### Scenario: FTS5 tests describe default runtime expectation
- **WHEN** FTS5-related tests or diagnostics report unavailable support
- **THEN** they describe it as an unexpected runtime limitation
- **AND** they do not instruct the user to rebuild with `-tags "fts5"`
