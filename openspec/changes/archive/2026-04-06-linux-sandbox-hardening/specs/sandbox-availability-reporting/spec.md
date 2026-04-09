## ADDED Requirements

### Requirement: OSIsolator reports unavailability reason
The `OSIsolator` interface SHALL include a `Reason() string` method that returns a human-readable explanation of why the isolator is unavailable. When `Available()` returns true, `Reason()` SHALL return an empty string.

#### Scenario: Seatbelt unavailable on macOS
- **WHEN** `sandbox-exec` is not found in PATH on macOS
- **THEN** `Reason()` returns `"sandbox-exec not found in PATH"`

#### Scenario: Noop isolator on Linux
- **WHEN** the platform is Linux with no isolation backend
- **THEN** `Reason()` returns `"Linux isolation backend not yet implemented"`

#### Scenario: Sandbox disabled by configuration
- **WHEN** `sandbox.enabled` is false
- **THEN** `disabledIsolator.Reason()` returns `"sandbox disabled by configuration"`

### Requirement: PlatformCapabilities includes probe reasons
`PlatformCapabilities` SHALL include `SeatbeltReason`, `LandlockReason`, and `SeccompReason` string fields that explain the probe result for each primitive.

#### Scenario: Linux with unimplemented probes
- **WHEN** running on Linux where kernel probes are not yet implemented
- **THEN** `LandlockReason` and `SeccompReason` are `"probe not yet implemented"` and `HasLandlock`/`HasSeccomp` are false

#### Scenario: macOS with sandbox-exec available
- **WHEN** running on macOS with `sandbox-exec` in PATH
- **THEN** `SeatbeltReason` is `"sandbox-exec found"` and `HasSeatbelt` is true

### Requirement: SandboxStatus combines config and runtime
The system SHALL provide a `SandboxStatus` struct that combines `Enabled`, `FailClosed`, the active `OSIsolator`, and `PlatformCapabilities`. The `OSIsolator` field SHALL never be nil — a `disabledIsolator` is substituted when sandbox is disabled.

#### Scenario: Sandbox disabled
- **WHEN** `NewSandboxStatus()` is called with a nil isolator
- **THEN** the returned status contains a `disabledIsolator` (not nil)

### Requirement: Summary reflects unknown probe state
`PlatformCapabilities.Summary()` SHALL return `"unknown (Landlock/seccomp probe not yet implemented)"` when the platform is Linux and probes have not been implemented.

#### Scenario: Linux summary with stub probes
- **WHEN** `Platform` is `"linux"` and `LandlockReason` is `"probe not yet implemented"`
- **THEN** `Summary()` returns a string containing `"unknown"`
