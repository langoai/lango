## Why

PR 1 restored the Linux build with a noop isolator and PR 2 added a backend
registry, but Linux still had no real isolation backend — `bwrap` and `native`
were both stubs reporting `Available()=false`. Documentation and TUI labels
correctly said "planned, not yet enforced", but that left every Linux user
without any sandbox enforcement at all. The kernel capability probes were
also stubs returning `(false, "probe not yet implemented")`, so `lango sandbox
status` could not honestly report what the kernel actually supports.

This change ships the first real Linux isolation backend (bubblewrap-based)
and replaces the kernel probe stubs with real syscalls, so that Linux users
get actual filesystem and network isolation as soon as the `bubblewrap`
package is installed, and `sandbox status` reports honest kernel capability
data.

## What Changes

- Add `BwrapIsolator` (Linux-only via `//go:build linux`) that wraps
  `exec.Cmd` with `bwrap` for filesystem bind mounts, `--unshare-net`,
  PID/IPC/UTS namespaces, and `--die-with-parent`. The absolute `bwrap`
  path is captured at probe time (`exec.LookPath` + `filepath.Abs`) and
  reused at `Apply()` so PATH/cwd changes between probe and exec cannot
  redirect to a different binary.
- Add a platform-agnostic `compileBwrapArgs(Policy) ([]string, error)`
  function (no build tag) that translates `Policy` into bwrap CLI flags.
  Reuses the existing `sanitizePath` injection guard. Directory-only deny
  via `--tmpfs <dir>` (file-level deny is planned for a follow-up).
- Add `bwrap_other.go` (`//go:build !linux`) with a no-op `BwrapIsolator`
  stub that reports `Reason()="bwrap is Linux-only"`. Remove the legacy
  `bwrapStub` type and `NewBwrapStub()` factory from `registry.go` since
  `PlatformBackendCandidates()` now always calls `NewBwrapIsolator()`.
- Replace the Landlock and seccomp probe stubs in `isolator_linux.go`:
  - Landlock: `landlock_create_ruleset(NULL, 0, LANDLOCK_CREATE_RULESET_VERSION)`
    via `unix.Syscall`, returning the ABI version on success and `ENOSYS`
    on kernels < 5.13.
  - seccomp: `unix.PrctlRetInt(PR_GET_SECCOMP, ...)` augmented with
    `/proc/self/status:Seccomp`. The `HasSeccomp` field doc and reason
    text explicitly state this is a presence signal only and does NOT
    prove BPF filter capability.
- Promote `golang.org/x/sys` from indirect to direct in `go.mod`.
- Expand `lango sandbox test` to four cases (write deny, read allow,
  workspace write, network deny) and add a hidden `lango sandbox
  _probe-net <addr>` self-subcommand used by the network test. The
  network test opens an ephemeral 127.0.0.1 listener in the parent
  and re-invokes the lango binary as a sandboxed child via
  `os.Executable()` so it depends on no external tools (`nc`/`curl`/
  `bash`/`/dev/tcp`).
- **Smoke test reliability fix**: replace shell `>/dev/null 2>&1` redirection
  with parent-side `io.Discard` in all four test cases. The previous
  shell-based read test had been silently failing on macOS Seatbelt the
  entire time because Seatbelt's `(deny default)` blocks `/dev/null`
  open-for-write, so the shell exited non-zero before `cat` ever ran.
  `runReadTest` now invokes `/bin/cat` directly; `runWriteTest` and
  `runWorkspaceWriteTest` invoke `/usr/bin/touch` directly.
  `runWorkspaceWriteTest` also resolves `os.MkdirTemp` results through
  `filepath.EvalSymlinks` because Seatbelt matches subpaths against the
  realpath (`/private/var/folders/...`) and not the user-visible
  `/var/folders/...` symlink.
- Synchronize all downstream advertised surfaces: README, `docs/cli/sandbox.md`,
  `docs/configuration.md`, `internal/cli/settings/forms_sandbox.go`,
  `internal/config/types_sandbox.go`, and `internal/sandbox/os/isolator.go`.
  Replaces every "Linux: planned, not yet enforced" string with the actual
  Linux state, while keeping items that remain unimplemented (native backend,
  file-level deny, escape hardening) honestly marked as planned.

## Capabilities

### New Capabilities
- `linux-bwrap-isolation`: Defines the `BwrapIsolator` contract,
  `compileBwrapArgs` Policy → argv mapping (filesystem, network,
  namespaces), absolute-path resolution at probe time, version
  capture, and the directory-only deny constraint.

### Modified Capabilities
- `os-sandbox-core`: `OSIsolator` interface doc reflects bwrap on Linux.
  `PlatformCapabilities.HasSeccomp` doc gains the "presence signal only"
  caveat. Adds the requirement that the Linux probe uses real syscalls
  via `golang.org/x/sys/unix` instead of stub returns.
- `sandbox-backend-registry`: Removes `NewBwrapStub` (the symbol no
  longer exists). `PlatformBackendCandidates()` on darwin and linux now
  registers `NewBwrapIsolator()` (real on linux, build-tag stub on
  non-linux).
- `os-sandbox-cli`: Adds the requirement that `lango sandbox test` runs
  four cases (write deny, read allow, workspace write, network deny),
  uses parent-side `io.Discard` instead of shell redirection so that
  Seatbelt's default-deny on `/dev/null` cannot cause false negatives,
  and provides the hidden `_probe-net` self-subcommand used by the
  network deny test.

## Impact

- **Code**: `internal/sandbox/os/{bwrap_args.go,bwrap_args_test.go,
  bwrap_linux.go,bwrap_other.go,bwrap_linux_test.go,registry.go,
  registry_test.go,isolator.go,isolator_linux.go,probe.go}`,
  `internal/cli/sandbox/sandbox.go`, `internal/config/types_sandbox.go`,
  `internal/cli/settings/forms_sandbox.go`.
- **Dependencies**: `golang.org/x/sys` promoted from indirect to direct
  in `go.mod`. No new third-party dependencies.
- **Runtime**: Linux users with `bubblewrap` installed now get real
  filesystem and network isolation when running tools, MCP stdio
  servers, and skill scripts. Linux users without `bubblewrap` see an
  honest "bwrap binary not found in PATH (install bubblewrap package)"
  reason from `lango sandbox status` and continue to run unsandboxed
  unless `failClosed: true`.
- **Docs**: README feature list, `docs/cli/sandbox.md`,
  `docs/configuration.md` config table, TUI form descriptions, and
  Go doc comments all updated.
- **Out of scope (deferred)**: native Landlock+seccomp backend,
  file-level deny via `DenyPaths`, escape hardening (config files,
  skills, git metadata), exception policy, and `NormalizePaths`
  expansion to `sandbox.workspacePath` / `sandbox.allowedWritePaths`.
