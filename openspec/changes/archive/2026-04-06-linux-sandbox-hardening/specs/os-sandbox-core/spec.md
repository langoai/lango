## MODIFIED Requirements

### Requirement: OSIsolator interface
The `OSIsolator` interface SHALL provide `Apply`, `Available`, `Name`, and `Reason` methods. `Reason()` SHALL return a human-readable string explaining why the isolator is unavailable, or empty string when available.

#### Scenario: Interface contract
- **WHEN** any type implements `OSIsolator`
- **THEN** it provides `Apply(ctx, cmd, policy) error`, `Available() bool`, `Name() string`, and `Reason() string`

### Requirement: Platform capability detection
The system SHALL detect available OS sandbox primitives via `Probe()` returning `PlatformCapabilities` with `HasSeatbelt`, `SeatbeltReason`, `HasLandlock`, `LandlockABI`, `LandlockReason`, `HasSeccomp`, `SeccompReason`, `Platform`, `KernelVersion`. Probe functions SHALL NOT use concrete type-casts on isolator instances.

#### Scenario: Linux probe without concrete cast
- **WHEN** `Probe()` is called on Linux
- **THEN** `probePlatform()` uses standalone `probeLandlockKernel()` and `probeSeccompKernel()` functions without constructing isolator instances

#### Scenario: Reason fields populated
- **WHEN** `Probe()` is called on any platform
- **THEN** reason fields explain the probe result (e.g., `"sandbox-exec found"`, `"probe not yet implemented"`, `"not on Linux"`)

## ADDED Requirements

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
