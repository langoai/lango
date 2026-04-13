## MODIFIED Requirements

### Requirement: PlatformBackendCandidates is the single source of truth
The system SHALL provide `PlatformBackendCandidates() []BackendCandidate` that returns the candidate list for the current platform. Both wiring and CLI code paths SHALL use this helper to prevent drift. The bwrap slot SHALL always be populated by `NewBwrapIsolator()`, whose Linux build returns the real `BwrapIsolator` and whose non-Linux build returns an unavailable stub with `Reason()="bwrap is Linux-only"`.

#### Scenario: macOS candidates
- **WHEN** running on darwin
- **THEN** `PlatformBackendCandidates()` returns `[seatbelt, bwrap (non-Linux stub), native stub]`

#### Scenario: Linux candidates
- **WHEN** running on linux
- **THEN** `PlatformBackendCandidates()` returns `[bwrap (real BwrapIsolator), native stub]`

#### Scenario: bwrap slot uses NewBwrapIsolator on every platform
- **WHEN** the registry assembles the bwrap candidate
- **THEN** it SHALL call `NewBwrapIsolator()` (build-tag-split) and SHALL NOT call any function named `NewBwrapStub`

## REMOVED Requirements

### Requirement: Stub isolators for planned backends
**Reason**: The bwrap stub no longer exists. `NewBwrapIsolator()` (build-tag-split) replaces it: on Linux it returns the real `BwrapIsolator`, and on non-Linux it returns an unavailable stub with `Reason()="bwrap is Linux-only"`. The native stub remains and is now covered by its own requirement below.

**Migration**: Replace any use of `sandboxos.NewBwrapStub()` with `sandboxos.NewBwrapIsolator()`. The function signature and return type are identical (`OSIsolator`); only the underlying implementation changes.

## ADDED Requirements

### Requirement: Native stub remains a planned backend
The system SHALL provide `NewNativeStub() OSIsolator` returning an unavailable isolator with `Name()="native"` and `Reason()="native backend not yet implemented"`. This represents the planned Landlock+seccomp kernel-syscall backend whose implementation is deferred.

#### Scenario: Native stub reports planned status
- **WHEN** `NewNativeStub().Reason()` is called
- **THEN** it returns `"native backend not yet implemented"`

#### Scenario: Native stub Apply returns ErrIsolatorUnavailable
- **WHEN** `NewNativeStub().Apply(ctx, cmd, policy)` is called
- **THEN** it SHALL return `ErrIsolatorUnavailable`
