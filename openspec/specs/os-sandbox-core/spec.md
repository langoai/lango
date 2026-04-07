## ADDED Requirements

### Requirement: OSIsolator interface
The `OSIsolator` interface SHALL provide `Apply`, `Available`, `Name`, and `Reason` methods. `Reason()` SHALL return a human-readable string explaining why the isolator is unavailable, or empty string when available.

#### Scenario: Interface contract
- **WHEN** any type implements `OSIsolator`
- **THEN** it provides `Apply(ctx, cmd, policy) error`, `Available() bool`, `Name() string`, and `Reason() string`

#### Scenario: Apply sandbox to command
- **WHEN** `OSIsolator.Apply()` is called with a valid `exec.Cmd` and `Policy`
- **THEN** the command SHALL be modified to run under OS-level isolation (e.g., wrapped with `sandbox-exec` on macOS)

#### Scenario: Isolator unavailable
- **WHEN** `OSIsolator.Apply()` is called on a platform without sandbox support
- **THEN** the method SHALL return `ErrIsolatorUnavailable`

### Requirement: Policy types
The system SHALL define `Policy` with `FilesystemPolicy` (ReadOnlyGlobal, ReadPaths, WritePaths, DenyPaths), `NetworkPolicy` (deny/allow/unix-only), `ProcessPolicy` (AllowFork, AllowSignals), and `AllowedNetworkIPs` (macOS only).

#### Scenario: Default tool policy
- **WHEN** `DefaultToolPolicy(workDir)` is called
- **THEN** the policy SHALL allow global read, write to workDir+/tmp, deny network, allow fork

#### Scenario: Strict tool policy
- **WHEN** `StrictToolPolicy(workDir)` is called
- **THEN** the policy SHALL additionally deny writes to `.git` under workDir

### Requirement: Seatbelt profile generation
The system SHALL generate macOS Seatbelt `.sb` profiles from Policy via `text/template` with default-deny base, path sanitization against injection characters, and IP allowlist support.

#### Scenario: Profile blocks injection characters
- **WHEN** a path contains `"`, `(`, `)`, or `;`
- **THEN** `GenerateSeatbeltProfile()` SHALL return `ErrInvalidPolicy`

#### Scenario: Profile includes allowed IPs
- **WHEN** Policy has `AllowedNetworkIPs` with entries
- **THEN** the profile SHALL contain `(allow network-outbound (remote ip "..."))` rules

### Requirement: Platform capability detection
The system SHALL detect available OS sandbox primitives via `Probe()` returning `PlatformCapabilities` with `HasSeatbelt`, `SeatbeltReason`, `HasLandlock`, `LandlockABI`, `LandlockReason`, `HasSeccomp`, `SeccompReason`, `Platform`, `KernelVersion`. Probe functions SHALL NOT use concrete type-casts on isolator instances.

#### Scenario: macOS probe detects seatbelt
- **WHEN** `Probe()` is called on macOS with sandbox-exec available
- **THEN** `HasSeatbelt` SHALL be true

#### Scenario: Linux probe without concrete cast
- **WHEN** `Probe()` is called on Linux
- **THEN** `probePlatform()` uses standalone probe functions without constructing isolator instances

#### Scenario: Reason fields populated
- **WHEN** `Probe()` is called on any platform
- **THEN** reason fields explain the probe result (e.g., `"sandbox-exec found"`, `"probe not yet implemented"`, `"not on Linux"`)

### Requirement: Cross-platform build tags
The system SHALL compile on darwin, linux, and other platforms using build-tag stubs that return `ErrIsolatorUnavailable` for unavailable primitives.

#### Scenario: Build on unsupported platform
- **WHEN** the project is built on a platform without sandbox support
- **THEN** `NewOSIsolator()` SHALL return a noop isolator with `Available() == false`

### Requirement: noopIsolator carries reason
The `noopIsolator` SHALL accept a `reason` string field and return it from `Reason()`. When reason is empty, it SHALL return `"no OS isolator available"`.

#### Scenario: Linux noop with reason
- **WHEN** `newPlatformIsolator()` returns a noop on Linux
- **THEN** `Reason()` returns `"Linux isolation backend not yet implemented"`

### Requirement: disabledIsolator for config-off
A `disabledIsolator` type SHALL exist that returns `Available()=false`, `Name()="disabled"`, `Reason()="sandbox disabled by configuration"`.

#### Scenario: Disabled sandbox
- **WHEN** sandbox is disabled by configuration and isolator is nil
- **THEN** `disabledIsolator` is used as a nil-safe substitute

### Requirement: Backend registry symbols
The `internal/sandbox/os` package SHALL export `BackendMode`, `BackendCandidate`, `BackendInfo`, `ParseBackendMode`, `SelectBackend`, `ListBackends`, `PlatformBackendCandidates`, `NewBwrapStub`, and `NewNativeStub` as the primary backend selection API. The `OSIsolator` interface SHALL remain unchanged.

#### Scenario: Symbols importable from sandboxos
- **WHEN** consumer code imports `sandboxos "github.com/langoai/lango/internal/sandbox/os"`
- **THEN** all backend registry symbols are accessible via the `sandboxos` package qualifier
