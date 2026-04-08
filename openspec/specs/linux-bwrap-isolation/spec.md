## Purpose

Linux-specific sandbox isolator that wraps `exec.Cmd` with bubblewrap (`bwrap`) to provide process-level isolation: filesystem bind mounts (read-only root + writable workspace), network unshare, and PID/IPC/UTS/cgroup namespaces. Availability is validated by a two-phase kernel namespace smoke probe, and the isolator's `Apply()` method is the single integration point that exec/skill/MCP consumers call at runtime.
## Requirements
### Requirement: BwrapIsolator wraps exec.Cmd via bubblewrap on Linux
The system SHALL provide a `BwrapIsolator` type in `internal/sandbox/os/bwrap_linux.go` (`//go:build linux`) that implements `OSIsolator` and wraps an `exec.Cmd` so the child process runs inside a bubblewrap container.

Availability SHALL be determined by a **two-phase namespace smoke probe** in addition to the existing `bwrap --version` check. See `Requirement: BwrapIsolator validates namespace creation via two-phase smoke probe` for the probe contract. When `Available()` returns `true`, `Reason()` SHALL still return `""` — partial degradation (network isolation unavailable) is surfaced through the dedicated `NetworkIsolationAvailable()`/`NetworkIsolationReason()` methods rather than through `Reason()`, so the existing contract is preserved for consumers that only care about base availability.

#### Scenario: Linux build returns real isolator
- **WHEN** the package is built with `GOOS=linux`
- **THEN** `NewBwrapIsolator()` SHALL return a `*BwrapIsolator` value (not a stub)

#### Scenario: Available when bwrap is installed and base smoke probe succeeds
- **WHEN** `bwrap` is on `PATH` AND `bwrap --version` succeeds AND the base namespace smoke probe (NetworkAllow policy shape) succeeds
- **THEN** `NewBwrapIsolator().Available()` SHALL return `true` and `Reason()` SHALL return `""`

#### Scenario: Unavailable when bwrap binary is missing
- **WHEN** `bwrap` is not on `PATH`
- **THEN** `Available()` SHALL return `false` and `Reason()` SHALL contain `"bwrap binary not found in PATH (install bubblewrap package)"`

#### Scenario: Apply rejects when unavailable
- **WHEN** `Apply` is called on an unavailable `BwrapIsolator`
- **THEN** it SHALL return `ErrIsolatorUnavailable`

### Requirement: BwrapIsolator captures absolute path at probe time
The system SHALL resolve the bwrap binary's absolute path during `NewBwrapIsolator()` (via `exec.LookPath` followed by `filepath.Abs`) and store it on the `BwrapIsolator` struct. `Apply()` SHALL set both `cmd.Path` and `cmd.Args[0]` to that captured absolute path. This guarantees that the same binary is used at probe time and exec time even if `PATH` or `cwd` changes between them.

#### Scenario: Apply uses captured absolute path
- **WHEN** `Apply` rewrites a command on Linux with bwrap available
- **THEN** `cmd.Path` SHALL equal the absolute path that `NewBwrapIsolator()` resolved, and `cmd.Args[0]` SHALL equal that same absolute path (not the bare string `"bwrap"`)

#### Scenario: Original args appear after the separator
- **WHEN** `Apply` rewrites a command
- **THEN** `cmd.Args` SHALL contain a `"--"` separator after the bwrap flags, and the original command argv SHALL appear after the separator

### Requirement: BwrapIsolator captures version string
The system SHALL execute `bwrap --version` during `NewBwrapIsolator()` and store the trimmed output on the `BwrapIsolator` struct. The struct SHALL expose this via a `Version() string` method so optional `versioner` consumers (such as `lango sandbox test`) can display it.

#### Scenario: Available isolator has non-empty version
- **WHEN** `NewBwrapIsolator()` returns an available isolator on Linux
- **THEN** the underlying `*BwrapIsolator.Version()` SHALL return a non-empty string

### Requirement: BwrapIsolator is a no-op on non-Linux platforms
The system SHALL provide a `bwrap_other.go` file (`//go:build !linux`) with a `NewBwrapIsolator()` factory that returns an unavailable stub whose `Name()="bwrap"` and `Reason()="bwrap is Linux-only"`. Cross-build (`GOOS=linux GOARCH=amd64 go build`) on macOS SHALL succeed without additional toolchain.

#### Scenario: Non-Linux returns unavailable stub
- **WHEN** the package is built on darwin or windows
- **THEN** `NewBwrapIsolator().Available()` SHALL return `false` and `Reason()` SHALL equal `"bwrap is Linux-only"`

#### Scenario: Cross-build succeeds from macOS
- **WHEN** `GOOS=linux GOARCH=amd64 go build ./...` runs on darwin
- **THEN** the build SHALL succeed with no errors

### Requirement: compileBwrapArgs translates Policy to bwrap argv
The system SHALL provide a platform-agnostic `compileBwrapArgs(policy Policy) ([]string, error)` function in `internal/sandbox/os/bwrap_args.go` (no build tag) that converts a `Policy` into the bwrap CLI arguments to prepend before the original command. The function SHALL reuse `sanitizePath()` for injection safety on every path it inserts.

#### Scenario: Standard namespace flags are always present
- **WHEN** `compileBwrapArgs` is called with any policy
- **THEN** the returned slice SHALL contain `--die-with-parent`, `--unshare-pid`, `--unshare-ipc`, `--unshare-uts`, `--unshare-cgroup-try`, `--proc /proc`, `--dev /dev`, and `--tmpfs /run`

#### Scenario: ReadOnlyGlobal mounts root read-only
- **WHEN** `policy.Filesystem.ReadOnlyGlobal` is true
- **THEN** the returned slice SHALL contain the triple `--ro-bind / /`

#### Scenario: ReadPaths used when not global
- **WHEN** `policy.Filesystem.ReadOnlyGlobal` is false and `ReadPaths` has entries
- **THEN** each `ReadPath` SHALL appear as a `--ro-bind <path> <path>` triple and the global `--ro-bind / /` SHALL NOT be present

#### Scenario: WritePaths use --bind
- **WHEN** `policy.Filesystem.WritePaths` has entries
- **THEN** each entry SHALL appear as a `--bind <path> <path>` triple

#### Scenario: NetworkDeny adds --unshare-net
- **WHEN** `policy.Network` is `NetworkDeny`
- **THEN** the returned slice SHALL contain `--unshare-net`

#### Scenario: NetworkAllow omits --unshare-net
- **WHEN** `policy.Network` is `NetworkAllow`
- **THEN** the returned slice SHALL NOT contain `--unshare-net`

#### Scenario: NetworkUnixOnly is treated as deny
- **WHEN** `policy.Network` is `NetworkUnixOnly`
- **THEN** the returned slice SHALL contain `--unshare-net` (bwrap has no AF_UNIX-only filter)

#### Scenario: Path injection is rejected
- **WHEN** `compileBwrapArgs` is called with a path containing `;`, `\n`, `"`, `(`, or `)`
- **THEN** it SHALL return an error wrapping `ErrInvalidPolicy`

### Requirement: DenyPaths must be existing directories
The `compileBwrapArgs` function SHALL stat each `DenyPath` and return a clear error when the path does not exist or is not a directory. Each valid `DenyPath` SHALL be added to the argv as `--tmpfs <path>`. File-level deny is NOT supported in this version (planned for a follow-up).

#### Scenario: Directory deny path is masked with tmpfs
- **WHEN** `policy.Filesystem.DenyPaths` contains an existing directory
- **THEN** the returned slice SHALL contain `--tmpfs <path>`

#### Scenario: File deny path is rejected
- **WHEN** `policy.Filesystem.DenyPaths` contains the path of an existing regular file
- **THEN** `compileBwrapArgs` SHALL return an error containing `"must be a directory"`

#### Scenario: Missing deny path is rejected
- **WHEN** `policy.Filesystem.DenyPaths` contains a path that does not exist
- **THEN** `compileBwrapArgs` SHALL return an error referencing the missing path

### Requirement: BwrapIsolator validates namespace creation via two-phase smoke probe
`NewBwrapIsolator()` SHALL, after the `bwrap --version` integrity check, execute two smoke probes in sequence. Each probe SHALL generate its argv by calling `compileBwrapArgs(probePolicy)` and appending `"--", "/bin/true"` — argv generation MUST NOT be hand-maintained in the probe, so probe argv and runtime argv cannot drift as `compileBwrapArgs` evolves.

1. **Base probe** uses `Policy{Filesystem:{ReadOnlyGlobal:true}, Network:NetworkAllow, Process:{AllowFork:true}}` — the minimum every lango consumer needs, matching `MCPServerPolicy`'s network model exactly. Failure (non-zero exit or timeout) SHALL make `Available()` return `false` and `Reason()` SHALL contain an actionable diagnostic referencing the probe failure and common root causes (`kernel.unprivileged_userns_clone=0`, AppArmor lockdown, or missing setuid-root).
2. **Network probe** uses `Policy{Filesystem:{ReadOnlyGlobal:true}, Network:NetworkDeny, Process:{AllowFork:true}}` — additionally validates that `--unshare-net` is permitted. Failure SHALL NOT change `Available()`; instead, the isolator SHALL expose `NetworkIsolationAvailable() bool` (returning `false`) and `NetworkIsolationReason() string` (non-empty, containing the probe error). `Apply()` SHALL return an `ErrIsolatorUnavailable`-wrapped error for policies whose `Network` is `NetworkDeny` or `NetworkUnixOnly`.

Each probe SHALL be bounded by a 2-second timeout via `context.WithTimeout`. Base probe failure SHALL short-circuit — the network probe SHALL NOT run if the base probe failed.

The Apply-time network gate SHALL reject policies BEFORE mutating `cmd.Path` or `cmd.Args`, so a rejected command can be retried or fallen back to an alternative isolator without leaving the caller in an inconsistent state.

#### Scenario: Both probes use compileBwrapArgs for argv generation
- **WHEN** either smoke probe runs
- **THEN** its argv SHALL be exactly `append(compileBwrapArgs(probePolicy), "--", "/bin/true")` so probe and runtime share the same flag generator

#### Scenario: Base probe failure marks isolator unavailable
- **WHEN** the base probe (NetworkAllow) exits non-zero or times out (2-second timeout)
- **THEN** `Available()` SHALL return `false` AND `Reason()` SHALL contain the probe error AND the network probe SHALL NOT run

#### Scenario: Network probe failure downgrades to base-only
- **WHEN** the base probe succeeds AND the network probe (NetworkDeny) exits non-zero or times out (2-second timeout)
- **THEN** `Available()` SHALL return `true` AND `Reason()` SHALL return `""` AND `NetworkIsolationAvailable()` SHALL return `false` AND `NetworkIsolationReason()` SHALL be non-empty

#### Scenario: Apply rejects NetworkDeny when network isolation unavailable
- **WHEN** `NetworkIsolationAvailable()==false` AND `Apply()` is called with a policy whose `Network` is `NetworkDeny` or `NetworkUnixOnly`
- **THEN** `Apply()` SHALL return an error wrapping `ErrIsolatorUnavailable` and referencing the network isolation diagnostic, without mutating `cmd.Path` or `cmd.Args`

#### Scenario: Apply permits NetworkAllow when network isolation unavailable
- **WHEN** `NetworkIsolationAvailable()==false` AND `Apply()` is called with a policy whose `Network` is `NetworkAllow` (e.g. `MCPServerPolicy`)
- **THEN** `Apply()` SHALL succeed and rewrite `cmd` normally — partial degradation does not affect `NetworkAllow` consumers

#### Scenario: NetworkIsolationReason empty when network isolation available
- **WHEN** both the base probe and the network probe succeed
- **THEN** `NetworkIsolationAvailable()` SHALL return `true` AND `NetworkIsolationReason()` SHALL return `""`

