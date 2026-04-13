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

`DefaultToolPolicy(workDir, dataRoot string) Policy` SHALL route `workDir` through a `canonicalWorkDir` helper (`filepath.Abs` + `filepath.EvalSymlinks` with nonexistent fallback) so that `WritePaths[0]` is the canonical filesystem path — symlinked workspaces no longer leak their pre-resolve path into the writable set. It SHALL then call a private `findGitRoot` helper that walks upward from the canonical workDir looking for the first ancestor whose `.git` entry is a directory (standard repo) or a regular file (linked worktree pointer). The walk terminates when `filepath.Dir(cur) == cur` (filesystem root).

`findGitRoot` SHALL return a `gitRoot` struct with `pointerPath` and `gitdirPath` fields. For a standard `.git` directory, both fields equal the `.git` directory path. For a linked worktree `.git` file, `pointerPath` is the file itself and `gitdirPath` is the resolved target parsed from the file's `gitdir: <path>` line (relative targets resolved against the pointer file's parent directory, then flowed through `filepath.Abs` + `filepath.EvalSymlinks`). Malformed or unreadable pointer files SHALL degrade to `gitdirPath = ""` — callers still deny the pointer file itself via file-level deny.

`DefaultToolPolicy` and `MCPServerPolicy` SHALL share a `collectBaselineDeny(workDir, dataRoot)` helper that applies the two-deny strategy: the gitRoot `pointerPath` and (when non-empty and distinct) `gitdirPath` are both added to `DenyPaths`, followed by the resolved `dataRoot` (when non-empty and existing as a directory). This means:
- Standard repo → one DenyPaths entry (`.git` directory)
- Linked worktree → two DenyPaths entries (pointer file + gitdir target, which may lie outside the workspace — that is the normal worktree layout)
- Malformed worktree pointer → one DenyPaths entry (pointer file only, degraded)
- Non-repo workspace → zero entries from git; dataRoot added when present

`StrictToolPolicy(workDir, dataRoot string) Policy` SHALL currently return the same policy as `DefaultToolPolicy`.

`MCPServerPolicy(workDir, dataRoot string) Policy` SHALL apply the same `collectBaselineDeny` logic via the shared helper so MCP stdio server children get symmetric protection with exec/skill tools. Empty `workDir` or `dataRoot` SHALL silently skip the corresponding baseline deny.

The system SHALL also provide a `normalizePath(entry string) ([]string, error)` helper in `internal/sandbox/os/policy.go` that runs the canonical path normalization pipeline shared by all sandbox backends:

```
entry → sanitize → filepath.Abs → filepath.Glob → filepath.EvalSymlinks (with nonexistent fallback) → []string
```

Every policy path consumer (bwrap `compileBwrapArgs` ReadPaths/WritePaths/DenyPaths loops; Seatbelt `GenerateSeatbeltProfile` ReadPaths/WritePaths/DenyPaths loops; future native Linux backend) SHALL use `normalizePath` so bwrap and Seatbelt cannot drift. Zero-match globs silently skip; invalid glob patterns return an error; nonexistent paths fall back to the pre-resolve absolute path so downstream `os.Stat` catches the missing-path error with the existing error message format.

#### Scenario: Default tool policy denies existing .git directory and dataRoot
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND an ancestor `.git` directory exists (immediate or via walk-up) AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain both the discovered `.git` path AND `dataRoot`
- **AND** `policy.Filesystem.WritePaths` SHALL contain the canonical workDir path (symlink-resolved) and `/tmp` (symlink-resolved on macOS to `/private/tmp`)
- **AND** `policy.Network` SHALL equal `NetworkDeny`

#### Scenario: Default tool policy walks up to find ancestor .git
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `workDir` is a subdirectory of a git repository whose `.git` directory exists at an ancestor path
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain the absolute path of the discovered ancestor `.git` directory, not a fictional path under `workDir`

#### Scenario: No ancestor .git is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND no ancestor directory of `workDir` contains a `.git` directory (non-repo workspace)
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain any `.git` path
- **AND** the policy SHALL still be buildable

#### Scenario: Linked worktree denies both pointer and gitdir target
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `<workDir>/.git` exists as a regular file containing `gitdir: <target>` AND the target resolves to an existing directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain BOTH the `.git` pointer file AND the resolved gitdir target (two entries)
- **AND** the pointer file deny flows through file-level bwrap/Seatbelt enforcement (`--ro-bind /dev/null` for bwrap)

#### Scenario: Malformed worktree pointer degrades to pointer-only deny
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `<workDir>/.git` exists as a regular file that does NOT start with `gitdir:`
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain the `.git` pointer file (one entry)
- **AND** the policy SHALL still be buildable with degraded protection

#### Scenario: Symlinked workDir resolves to canonical path
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `workDir` is a symlink to a real directory
- **THEN** `policy.Filesystem.WritePaths[0]` SHALL equal the symlink-resolved real path, NOT the original symlink path
- **AND** `findGitRoot` SHALL walk up from the resolved real path

#### Scenario: Missing dataRoot is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND `dataRoot` is non-empty but does not exist as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain `dataRoot`
- **AND** the policy SHALL still be buildable

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

#### Scenario: normalizePath pipeline expands globs
- **WHEN** `normalizePath(pattern)` is called with a pattern containing `*`, `?`, or `[` AND the pattern matches one or more files
- **THEN** it SHALL return a slice with one entry per match, each run through `filepath.EvalSymlinks`

#### Scenario: normalizePath silently skips unmatched globs
- **WHEN** `normalizePath(pattern)` is called with a pattern that matches zero files
- **THEN** it SHALL return a nil slice with a nil error

#### Scenario: normalizePath rejects invalid glob patterns
- **WHEN** `normalizePath(entry)` is called with an entry containing an unclosed bracket
- **THEN** it SHALL return an error containing `"invalid glob pattern"`

#### Scenario: normalizePath falls back to Abs path for nonexistent entries
- **WHEN** `normalizePath(entry)` is called with a nonexistent path that has no glob characters
- **THEN** it SHALL return a slice containing the `filepath.Abs` result (not an error) so downstream `os.Stat` catches the missing-path error with the existing format

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

