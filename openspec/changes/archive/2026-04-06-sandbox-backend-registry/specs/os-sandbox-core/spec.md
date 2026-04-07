## ADDED Requirements

### Requirement: Backend registry symbols
The `internal/sandbox/os` package SHALL export `BackendMode`, `BackendCandidate`, `BackendInfo`, `ParseBackendMode`, `SelectBackend`, `ListBackends`, `PlatformBackendCandidates`, `NewBwrapStub`, and `NewNativeStub` as the primary backend selection API. The `OSIsolator` interface SHALL remain unchanged.

#### Scenario: Symbols importable from sandboxos
- **WHEN** consumer code imports `sandboxos "github.com/langoai/lango/internal/sandbox/os"`
- **THEN** all backend registry symbols are accessible via the `sandboxos` package qualifier
