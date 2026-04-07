### Requirement: BackendMode enum identifies isolation backends
The system SHALL provide a `BackendMode` enum type with values `BackendAuto`, `BackendSeatbelt`, `BackendBwrap`, `BackendNative`, `BackendNone`. Each value SHALL have a `String()` method returning its config-string form (`"auto"`, `"seatbelt"`, etc.).

#### Scenario: String form for each mode
- **WHEN** `BackendBwrap.String()` is called
- **THEN** it returns `"bwrap"`

### Requirement: ParseBackendMode rejects unknown values
`ParseBackendMode(s string) (BackendMode, error)` SHALL map config strings to `BackendMode`. The empty string SHALL map to `BackendAuto`. Unknown strings SHALL return an error.

#### Scenario: Empty string maps to auto
- **WHEN** `ParseBackendMode("")` is called
- **THEN** it returns `(BackendAuto, nil)`

#### Scenario: Typo rejected
- **WHEN** `ParseBackendMode("seatbeltt")` is called
- **THEN** it returns an error containing `"unknown sandbox backend"`

### Requirement: BackendCandidate provides typed identity
The system SHALL provide a `BackendCandidate` struct pairing a `BackendMode` with an `OSIsolator`. Selection logic SHALL identify candidates by `Mode` (not by `Name()` string matching).

#### Scenario: Selection by mode
- **WHEN** `SelectBackend(BackendBwrap, candidates)` is called and one candidate has `Mode: BackendBwrap`
- **THEN** that candidate's isolator is returned regardless of its `Name()` string

### Requirement: SelectBackend handles auto, none, and explicit modes
`SelectBackend(mode BackendMode, candidates []BackendCandidate) (OSIsolator, BackendInfo)` SHALL return an isolator and `BackendInfo` describing the result. Behavior:
- `BackendAuto`: returns the first candidate where `Available()=true`. If none available, returns a noop with aggregated candidate reasons.
- `BackendNone`: always returns a noop with reason `"backend explicitly set to none"`.
- Explicit modes (seatbelt/bwrap/native): returns the matching candidate's isolator AS-IS even if `Available()=false`. If the mode is not in candidates, returns a noop with reason `"backend X not available on this platform"`.

#### Scenario: Auto picks first available
- **WHEN** candidates are `[seatbelt(available), bwrap(unavailable)]` and mode is `BackendAuto`
- **THEN** `SelectBackend` returns the seatbelt isolator

#### Scenario: Auto fallback aggregates reasons
- **WHEN** all candidates are unavailable and mode is `BackendAuto`
- **THEN** the returned noop's `Reason()` contains each candidate's name and reason joined by `"; "`

#### Scenario: Explicit unavailable preserves identity
- **WHEN** mode is `BackendBwrap` and the bwrap candidate is unavailable
- **THEN** `SelectBackend` returns the bwrap stub isolator (not a noop), preserving its `Name()` and `Reason()`

#### Scenario: None always returns noop
- **WHEN** mode is `BackendNone`
- **THEN** `SelectBackend` returns a noop with reason `"backend explicitly set to none"`

### Requirement: ListBackends reports all candidate states
`ListBackends(candidates []BackendCandidate) []BackendInfo` SHALL return one `BackendInfo` per candidate with `Name`, `Mode`, `Available`, `Reason` fields.

#### Scenario: All candidates listed
- **WHEN** `ListBackends([seatbelt, bwrap, native])` is called
- **THEN** the result contains 3 `BackendInfo` entries in input order

### Requirement: PlatformBackendCandidates is the single source of truth
The system SHALL provide `PlatformBackendCandidates() []BackendCandidate` that returns the candidate list for the current platform. Both wiring and CLI code paths SHALL use this helper to prevent drift.

#### Scenario: macOS candidates
- **WHEN** running on darwin
- **THEN** `PlatformBackendCandidates()` returns `[seatbelt, bwrap stub, native stub]`

#### Scenario: Linux candidates
- **WHEN** running on linux
- **THEN** `PlatformBackendCandidates()` returns `[bwrap stub, native stub]`

### Requirement: Stub isolators for planned backends
The system SHALL provide `NewBwrapStub()` and `NewNativeStub()` returning `OSIsolator` implementations with `Available()=false`, `Name()="bwrap"`/`"native"`, and `Reason()="bwrap backend not yet implemented"` / `"native backend not yet implemented"`.

#### Scenario: Bwrap stub reports planned status
- **WHEN** `NewBwrapStub().Reason()` is called
- **THEN** it returns `"bwrap backend not yet implemented"`
