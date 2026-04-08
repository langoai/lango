## Purpose

OS-level sandbox core: defines the `OSIsolator` interface, `Policy` types, platform-agnostic Seatbelt profile generation, and the backend registry. Policy helpers (`DefaultToolPolicy`, `StrictToolPolicy`, `MCPServerPolicy`) produce consistent baseline protection for exec/skill/MCP consumers, including walk-up `.git` discovery and lango control-plane masking.
## Requirements
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

`DefaultToolPolicy(workDir, dataRoot string) Policy` SHALL walk up from `workDir` via a private `findGitRoot` helper to discover the first ancestor `.git` directory and deny it as a baseline. `findGitRoot` SHALL terminate when `filepath.Dir(cur) == cur` (filesystem root) and SHALL return an empty string if no ancestor `.git` directory is found or if the only ancestor `.git` entry is a regular file (linked worktree pointer). When `dataRoot` is non-empty, it SHALL be resolved via `filepath.Abs` and added to `DenyPaths` only when the resolved path exists as a directory. Missing or non-directory entries SHALL be silently skipped so the policy remains buildable in non-repo workspaces, linked worktrees (where `.git` is a file and walk-up finds no ancestor directory), and minimal environments where the control-plane directory has not been created yet. This gate exists because `compileBwrapArgs` requires every deny path to exist as a directory and would otherwise fail the entire sandbox apply.

`StrictToolPolicy(workDir, dataRoot string) Policy` SHALL currently return the same policy as `DefaultToolPolicy`. The function is preserved as a separate symbol so future strict-only options can branch later without another signature migration.

`MCPServerPolicy(workDir, dataRoot string) Policy` SHALL apply the same walk-up `.git` baseline deny as `DefaultToolPolicy` via `findGitRoot(workDir)` when `workDir` is non-empty. It SHALL also deny `dataRoot` (when non-empty and when the resolved path exists as a directory) so MCP stdio server child processes cannot read or write the lango control-plane. Empty `workDir` or `dataRoot` SHALL silently skip the corresponding baseline deny. Other characteristics are retained: read-global, write-`/tmp` only, network allowed.

#### Scenario: Default tool policy denies existing .git and dataRoot
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND an ancestor `.git` directory exists (immediate or via walk-up) AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain both the discovered `.git` path AND `dataRoot`
- **AND** `policy.Filesystem.WritePaths` SHALL contain `workDir` and `/tmp`
- **AND** `policy.Network` SHALL equal `NetworkDeny`

#### Scenario: Default tool policy walks up to find ancestor .git
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `workDir` is a subdirectory of a git repository whose `.git` directory exists at an ancestor path (e.g. `workDir=/repo/cmd/lango`, ancestor `.git=/repo/.git`)
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain the absolute path of the discovered ancestor `.git` directory, not a fictional path under `workDir`

#### Scenario: No ancestor .git is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND no ancestor directory of `workDir` contains a `.git` directory (non-repo workspace)
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain any `.git` path
- **AND** the policy SHALL still be buildable so that `compileBwrapArgs` does not reject non-repo workspaces

#### Scenario: .git file (worktree) causes walk-up to continue past it
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `<workDir>/.git` exists as a regular file (linked worktree pointer) AND no ancestor directory above contains a `.git` directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain the `.git` file path
- **AND** the policy SHALL still be buildable
- **AND** walk-up SHALL NOT return the worktree file as a match (this gap closes in PR 5c with file-level deny semantics)

#### Scenario: Missing dataRoot is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND `dataRoot` is non-empty but does not exist as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain `dataRoot`
- **AND** the policy SHALL still be buildable

#### Scenario: Default tool policy with empty dataRoot
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND an ancestor `.git` directory exists
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain only the discovered ancestor `.git` path
- **AND** the policy SHALL be safe to use in isolated unit tests that do not have a real control-plane directory

#### Scenario: Strict tool policy mirrors default
- **WHEN** `StrictToolPolicy(workDir, dataRoot)` is called with the same arguments as `DefaultToolPolicy(workDir, dataRoot)`
- **THEN** the returned policies SHALL be equal (including DenyPaths order)

#### Scenario: MCP server policy denies existing dataRoot
- **WHEN** `MCPServerPolicy("", dataRoot)` is called AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain `dataRoot`
- **AND** `policy.Network` SHALL equal `NetworkAllow`

#### Scenario: MCP server policy with empty workDir and empty dataRoot
- **WHEN** `MCPServerPolicy("", "")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL be empty
- **AND** `policy.Network` SHALL equal `NetworkAllow`

#### Scenario: MCP server policy with missing dataRoot
- **WHEN** `MCPServerPolicy("", dataRoot)` is called AND `dataRoot` is non-empty but does not exist as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL be empty

#### Scenario: MCP server policy denies workspace .git via walk-up
- **WHEN** `MCPServerPolicy(workDir, "")` is called AND `workDir` is non-empty AND an ancestor `.git` directory exists via walk-up
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain the discovered ancestor `.git` path (symmetric with `DefaultToolPolicy`)

#### Scenario: MCP server policy denies both workspace .git and dataRoot
- **WHEN** `MCPServerPolicy(workDir, dataRoot)` is called AND an ancestor `.git` exists AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain both the discovered `.git` path AND `dataRoot`

### Requirement: Seatbelt profile generation
The system SHALL generate macOS Seatbelt `.sb` profiles from Policy via `text/template` with default-deny base, path sanitization against injection characters, and IP allowlist support.

When `Policy.Filesystem.DenyPaths` contains an entry, the generated profile SHALL include BOTH a `(deny file-read* (subpath "<path>"))` rule AND a `(deny file-write* (subpath "<path>"))` rule for that entry. The read-deny rule is required because `ReadOnlyGlobal=true` emits `(allow file-read*)` globally, and a write-only deny would leave sensitive files (such as `~/.lango/lango.db`, `.git/config`, or encrypted config tokens) fully readable by the sandboxed child. The control-plane deny added by `DefaultToolPolicy(workDir, dataRoot)` SHALL therefore appear in the generated Seatbelt profile as both read-deny and write-deny rules when `dataRoot` is non-empty and exists.

#### Scenario: Profile blocks injection characters
- **WHEN** a path contains `"`, `(`, `)`, or `;`
- **THEN** `GenerateSeatbeltProfile()` SHALL return `ErrInvalidPolicy`

#### Scenario: Profile includes allowed IPs
- **WHEN** Policy has `AllowedNetworkIPs` with entries
- **THEN** the profile SHALL contain `(allow network-outbound (remote ip "..."))` rules

#### Scenario: DenyPaths entries deny both read and write
- **WHEN** `Policy.Filesystem.DenyPaths` contains `/home/user/.lango` and `ReadOnlyGlobal=true`
- **THEN** the generated profile SHALL contain BOTH `(deny file-read* (subpath "/home/user/.lango"))` AND `(deny file-write* (subpath "/home/user/.lango"))`
- **AND** the sandboxed child SHALL NOT be able to read `/home/user/.lango/lango.db` or any other file under the control-plane directory

#### Scenario: Profile denies dataRoot for read and write when configured
- **WHEN** `GenerateSeatbeltProfile` is called on a Policy whose DenyPaths contains `/home/user/.lango`
- **THEN** the profile SHALL contain `(deny file-read* (subpath "/home/user/.lango"))` AND `(deny file-write* (subpath "/home/user/.lango"))`

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

### Requirement: bwrap mount ordering
`compileBwrapArgs` SHALL emit the root-level mount (`--ro-bind / /` when `ReadOnlyGlobal=true`, or `--ro-bind <p> <p>` entries for explicit `ReadPaths`) BEFORE the specialised mounts `--proc /proc`, `--dev /dev`, and `--tmpfs /run`. bubblewrap processes options left-to-right, and a later root bind would shadow any earlier mounts nested under the sandbox root, leaking the host's `/proc` and `/dev` into the sandboxed child and weakening PID namespace and device isolation. The specialised mounts must therefore be layered on top of the root bind, not underneath it.

#### Scenario: Root bind precedes --proc
- **WHEN** `compileBwrapArgs` is called with `ReadOnlyGlobal=true`
- **THEN** the index of `--ro-bind / /` in the returned argv slice SHALL be less than the index of `--proc /proc`
- **AND** the index of `--ro-bind / /` SHALL be less than the indices of `--dev /dev` and `--tmpfs /run`

#### Scenario: Specialised mounts still present
- **WHEN** `compileBwrapArgs` is called with any valid Policy
- **THEN** the returned argv slice SHALL contain `--proc /proc`, `--dev /dev`, and `--tmpfs /run` as three-token pairs, unconditionally

### Requirement: Sandbox path validation against DataRoot overlap
`config.Validate` SHALL reject configurations where `sandbox.workspacePath` or any entry of `sandbox.allowedWritePaths` resolves to `cfg.DataRoot` itself or to a subtree of `cfg.DataRoot`. This check is necessary because `DefaultToolPolicy` adds `cfg.DataRoot` to `DenyPaths`, and the resulting `--tmpfs cfg.DataRoot` mount (bwrap) or `(deny file* (subpath ...))` rule (Seatbelt) would cover the workspace and make it silently unreachable at runtime. The validation check SHALL fire AFTER `NormalizePaths` so it catches both relative paths that were resolved under `DataRoot` and absolute paths the user explicitly wrote inside the control-plane.

The validation error message SHALL name the colliding path, state that it is inside `cfg.DataRoot`, and direct the user to use an absolute path outside the control-plane.

#### Scenario: workspacePath nested under DataRoot rejected
- **WHEN** `cfg.DataRoot = "/tmp/lango"` and `cfg.Sandbox.WorkspacePath = "/tmp/lango/repo"`
- **THEN** `Validate(cfg)` SHALL return an error
- **AND** the error message SHALL contain `"sandbox.workspacePath"` and `"inside cfg.DataRoot"`

#### Scenario: workspacePath equal to DataRoot rejected
- **WHEN** `cfg.DataRoot = "/tmp/lango"` and `cfg.Sandbox.WorkspacePath = "/tmp/lango"`
- **THEN** `Validate(cfg)` SHALL return an error mentioning `sandbox.workspacePath`

#### Scenario: workspacePath outside DataRoot accepted
- **WHEN** `cfg.DataRoot = "/tmp/lango"` and `cfg.Sandbox.WorkspacePath = "/tmp/some-other-dir"`
- **THEN** `Validate(cfg)` SHALL return nil

#### Scenario: allowedWritePaths entry nested under DataRoot rejected
- **WHEN** `cfg.DataRoot = "/tmp/lango"` and `cfg.Sandbox.AllowedWritePaths` contains `"/tmp/lango/scratch"`
- **THEN** `Validate(cfg)` SHALL return an error naming `sandbox.allowedWritePaths` and the offending entry

#### Scenario: Empty workspacePath accepted
- **WHEN** `cfg.Sandbox.WorkspacePath = ""`
- **THEN** `Validate(cfg)` SHALL NOT error on the workspace path check (the supervisor falls back to `os.Getwd()` at runtime)

### Requirement: Sandbox status graceful degradation
`lango sandbox status` SHALL render the Sandbox Configuration, Active Isolation, Platform Capabilities, and Backend Availability sections in degraded modes (signed-out, locked DB, non-interactive environments, or missing BootLoader) by falling back to a config-only loader. The Recent Sandbox Decisions section SHALL be silently skipped when the audit DB is unreachable, but the rest of the command SHALL NOT error out — these diagnostic sections do not depend on the audit database.

`newStatusCmd` SHALL try the BootLoader first so that one bootstrap pass serves both the config rendering and the Recent Decisions audit query (preserving the no-double-passphrase contract). On nil BootLoader OR a BootLoader error, `newStatusCmd` SHALL fall back to the cfgLoader to load the config independently. Recent Decisions SHALL only render when the BootLoader returned a non-nil result with a non-nil DBClient.

#### Scenario: Nil BootLoader still renders config sections
- **WHEN** `lango sandbox status` is invoked with cfgLoader wired and BootLoader nil
- **THEN** the command SHALL exit successfully
- **AND** the output SHALL contain the `Sandbox Configuration:`, `Active Isolation:`, and `Backend Availability:` headers
- **AND** the output SHALL NOT contain a `Recent Sandbox Decisions` header

#### Scenario: BootLoader error falls back to cfgLoader
- **WHEN** `lango sandbox status` is invoked with cfgLoader wired and BootLoader returning an error
- **THEN** the command SHALL exit successfully via the cfgLoader fallback
- **AND** the non-audit sections SHALL render
- **AND** the `Recent Sandbox Decisions` section SHALL be silently skipped

#### Scenario: Healthy BootLoader runs only one bootstrap
- **WHEN** `lango sandbox status` is invoked with both loaders wired and BootLoader succeeding
- **THEN** cfgLoader SHALL NOT be called (the Recent Decisions path uses `boot.Config` directly)
- **AND** the user SHALL be prompted for the encryption passphrase at most once per invocation

### Requirement: Sandbox decision row formatter
The `Recent Sandbox Decisions` row formatter SHALL display `-` in the backend column whenever the decision is NOT `"applied"` OR the stored backend value is empty. Only `"applied"` decisions actually ran inside a sandbox backend; `excluded`, `skipped`, and `rejected` verdicts ran unsandboxed (or were blocked entirely), so echoing the published `Backend` value for those rows would falsely suggest the command was sandboxed under that backend.

The publish sites (`exec`, `skill`, `mcp`) SHALL continue to stamp the `Backend` field uniformly from the wired isolator's `Name()` regardless of decision; the verdict-specific formatting is the display layer's responsibility, not the publisher's.

#### Scenario: Applied decision shows backend
- **WHEN** an audit row has `decision="applied"` and `backend="bwrap"`
- **THEN** the rendered row SHALL show `bwrap` in the backend column

#### Scenario: Excluded decision shows dash
- **WHEN** an audit row has `decision="excluded"` and `backend="bwrap"`
- **THEN** the rendered row SHALL show `-` in the backend column
- **AND** the rendered row SHALL NOT contain the substring `bwrap`

#### Scenario: Skipped decision shows dash
- **WHEN** an audit row has `decision="skipped"` and `backend="seatbelt"`
- **THEN** the rendered row SHALL show `-` in the backend column
- **AND** the rendered row SHALL NOT contain the substring `seatbelt`

#### Scenario: Rejected decision shows dash
- **WHEN** an audit row has `decision="rejected"` and `backend="bwrap"`
- **THEN** the rendered row SHALL show `-` in the backend column

#### Scenario: Empty backend shows dash
- **WHEN** an audit row has `decision="applied"` and `backend=""`
- **THEN** the rendered row SHALL show `-` in the backend column

