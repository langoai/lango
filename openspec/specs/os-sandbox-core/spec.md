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

The policy helper functions SHALL accept a `dataRoot` parameter so they can deny the lango control-plane (typically `~/.lango`) on every sandboxed child. An empty `dataRoot` SHALL skip the control-plane mask so isolated unit tests can build a policy without fabricating a real directory.

`DefaultToolPolicy(workDir, dataRoot string) Policy` SHALL deny `<workDir>/.git` as a baseline (previously a `StrictToolPolicy`-only feature). When `dataRoot` is non-empty, it SHALL be resolved via `filepath.Abs` and added to `DenyPaths` so the lango data directory is masked from the sandboxed child.

`StrictToolPolicy(workDir, dataRoot string) Policy` SHALL currently return the same policy as `DefaultToolPolicy`. The function is preserved as a separate symbol so future strict-only options can branch later without another signature migration.

`MCPServerPolicy(dataRoot string) Policy` SHALL deny `dataRoot` (when non-empty) so MCP stdio server child processes cannot read or write the lango control-plane. It retains its other characteristics: read-global, write-`/tmp` only, network allowed.

#### Scenario: Default tool policy denies dataRoot and .git
- **WHEN** `DefaultToolPolicy("/home/user/project", "/home/user/.lango")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain `/home/user/project/.git` and `/home/user/.lango`
- **AND** `policy.Filesystem.WritePaths` SHALL contain `/home/user/project` and `/tmp`
- **AND** `policy.Network` SHALL equal `NetworkDeny`

#### Scenario: Default tool policy with empty dataRoot
- **WHEN** `DefaultToolPolicy("/home/user/project", "")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain only `/home/user/project/.git`
- **AND** the policy SHALL be safe to use in isolated unit tests that do not have a real control-plane directory

#### Scenario: Strict tool policy mirrors default
- **WHEN** `StrictToolPolicy(workDir, dataRoot)` is called with the same arguments as `DefaultToolPolicy(workDir, dataRoot)`
- **THEN** the returned policies SHALL be equal (including DenyPaths order)

#### Scenario: MCP server policy denies dataRoot
- **WHEN** `MCPServerPolicy("/home/user/.lango")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain `/home/user/.lango`
- **AND** `policy.Network` SHALL equal `NetworkAllow`

#### Scenario: MCP server policy with empty dataRoot
- **WHEN** `MCPServerPolicy("")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL be empty

### Requirement: Seatbelt profile generation
The system SHALL generate macOS Seatbelt `.sb` profiles from Policy via `text/template` with default-deny base, path sanitization against injection characters, and IP allowlist support.

When `Policy.Filesystem.DenyPaths` contains an entry, the generated profile SHALL include a `(deny file-write* (subpath "<path>"))` rule for that entry. The control-plane deny added by `DefaultToolPolicy(workDir, dataRoot)` SHALL therefore appear in the generated Seatbelt profile when `dataRoot` is non-empty.

#### Scenario: Profile blocks injection characters
- **WHEN** a path contains `"`, `(`, `)`, or `;`
- **THEN** `GenerateSeatbeltProfile()` SHALL return `ErrInvalidPolicy`

#### Scenario: Profile includes allowed IPs
- **WHEN** Policy has `AllowedNetworkIPs` with entries
- **THEN** the profile SHALL contain `(allow network-outbound (remote ip "..."))` rules

#### Scenario: Profile denies dataRoot when configured
- **WHEN** `GenerateSeatbeltProfile(DefaultToolPolicy("/tmp/work", "/home/user/.lango"))` is called
- **THEN** the profile SHALL contain `(deny file-write* (subpath "/home/user/.lango"))`

### Requirement: Platform capability detection
The system SHALL detect available OS sandbox primitives via `Probe()` returning `PlatformCapabilities` with `HasSeatbelt`, `SeatbeltReason`, `HasLandlock`, `LandlockABI`, `LandlockReason`, `HasSeccomp`, `SeccompReason`, `Platform`, `KernelVersion`. Probe functions SHALL NOT use concrete type-casts on isolator instances. On Linux, `probeLandlockKernel` and `probeSeccompKernel` SHALL use real syscalls via `golang.org/x/sys/unix` (not stub returns).

The `HasSeccomp` field doc comment SHALL state explicitly that a `true` value indicates only that the kernel exposes the seccomp prctl interface and does NOT prove that BPF filters are installable. The qualified description SHALL appear in `SeccompReason`.

#### Scenario: macOS probe detects seatbelt
- **WHEN** `Probe()` is called on macOS with sandbox-exec available
- **THEN** `HasSeatbelt` SHALL be true

#### Scenario: Linux probe without concrete cast
- **WHEN** `Probe()` is called on Linux
- **THEN** `probePlatform()` uses standalone probe functions without constructing isolator instances

#### Scenario: Reason fields populated
- **WHEN** `Probe()` is called on any platform
- **THEN** reason fields explain the probe result (e.g., `"sandbox-exec found"`, `"Landlock ABI 3"`, `"seccomp interface present (PR_GET_SECCOMP=N)"`, `"not on Linux"`)

#### Scenario: Linux Landlock probe uses real syscall
- **WHEN** `probeLandlockKernel()` runs on a Linux kernel ≥ 5.13 that supports Landlock
- **THEN** it SHALL return `(true, abi, "Landlock ABI N")` where `abi > 0`, having called `unix.Syscall(SYS_LANDLOCK_CREATE_RULESET, 0, 0, LANDLOCK_CREATE_RULESET_VERSION)`

#### Scenario: Linux Landlock probe handles ENOSYS
- **WHEN** `probeLandlockKernel()` runs on a Linux kernel that does not support Landlock
- **THEN** it SHALL return `(false, 0, "Landlock not supported by this kernel (requires 5.13+)")` after the syscall returns ENOSYS

#### Scenario: Linux seccomp probe captures presence only
- **WHEN** `probeSeccompKernel()` runs and `unix.PrctlRetInt(PR_GET_SECCOMP, ...)` returns successfully
- **THEN** it SHALL return `(true, reason)` where `reason` contains `"seccomp interface present"` and the substring `"BPF filter capability not directly verified"`

#### Scenario: Linux seccomp probe augments with /proc/self/status
- **WHEN** `/proc/self/status` is readable and contains a `Seccomp:` line
- **THEN** the seccomp reason SHALL include `"/proc/self/status Seccomp=<value>"`

#### Scenario: HasSeccomp doc comment carries the caveat
- **WHEN** a developer reads the `PlatformCapabilities.HasSeccomp` field doc comment
- **THEN** it SHALL state that `true` means only that the kernel exposes the prctl interface and does NOT prove BPF filter capability

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
