## ADDED Requirements

### Requirement: OSIsolator interface
The system SHALL provide an `OSIsolator` interface with `Apply(ctx, cmd, policy) error`, `Available() bool`, and `Name() string` methods for applying OS-level kernel restrictions to `exec.Cmd` before process start.

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

### Requirement: Platform capability probe
The system SHALL detect available OS sandbox primitives via `Probe()` returning `PlatformCapabilities` with HasSeatbelt, HasLandlock, LandlockABI, HasSeccomp, Platform, KernelVersion.

#### Scenario: macOS probe detects seatbelt
- **WHEN** `Probe()` is called on macOS with sandbox-exec available
- **THEN** `HasSeatbelt` SHALL be true

### Requirement: Cross-platform build tags
The system SHALL compile on darwin, linux, and other platforms using build-tag stubs that return `ErrIsolatorUnavailable` for unavailable primitives.

#### Scenario: Build on unsupported platform
- **WHEN** the project is built on a platform without sandbox support
- **THEN** `NewOSIsolator()` SHALL return a noop isolator with `Available() == false`
