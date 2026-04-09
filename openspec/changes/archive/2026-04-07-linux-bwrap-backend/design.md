## Context

PR 1 (`linux-sandbox-hardening`, archived 2026-04-06) restored the Linux build
with a noop isolator and an honest "probe not yet implemented" reason. PR 2
(`sandbox-backend-registry`, archived 2026-04-06) added the `BackendMode`
enum, `SelectBackend` policy, and `PlatformBackendCandidates()` helper.

After PR 2, the Linux candidate list was `[bwrap stub, native stub]` and
both stubs returned `Available()=false`. `lango sandbox status` honestly
showed Linux as having no available backend, but no enforcement existed.

This design lands the first real Linux backend by wiring `bubblewrap` (a
small setuid helper that builds Linux user/PID/network namespaces and bind
mounts a filesystem view) into the existing `OSIsolator` interface, and
replaces the kernel capability probe stubs with real syscall probes. The
native (Landlock+seccomp) backend remains a noop and is deferred.

## Goals / Non-Goals

**Goals:**
- Linux users with `bubblewrap` installed get real filesystem and network
  isolation when running tools, MCP stdio servers, and skill scripts.
- `lango sandbox status` reports honest kernel capability data
  (Landlock ABI version, seccomp interface presence) via real syscalls.
- `lango sandbox test` deterministically verifies isolation on both
  macOS Seatbelt and Linux bwrap with no external tool dependencies.
- All advertised surfaces (README, docs, TUI, Go doc comments) match
  the actual state.

**Non-Goals:**
- Native Landlock+seccomp backend implementation (deferred — only the
  *probe* lands here, not the enforcement).
- File-level deny via `Policy.DenyPaths` (PR 3 ships directory-only deny;
  file-level masking via `--ro-bind /dev/null <file>` is deferred).
- Escape hardening (config files, skills directory, git metadata
  protection) — deferred to a follow-up.
- Exception policy / excluded command patterns / unsandboxed override.
- Workspace path normalization in `NormalizePaths` (separate concern).

## Decisions

### D1 — `BwrapIsolator` is build-tag-split, registry calls one factory

`BwrapIsolator` lives in `bwrap_linux.go` with `//go:build linux`. A no-op
stub with the same `NewBwrapIsolator()` factory name lives in `bwrap_other.go`
with `//go:build !linux`. `PlatformBackendCandidates()` always calls
`NewBwrapIsolator()` and lets the build tag pick the right implementation.

**Rationale:** macOS dev environments still need to cross-build for Linux
(`GOOS=linux GOARCH=amd64 go build ./...`) and run unit tests for
non-isolator code. Splitting on the constructor lets the registry stay
build-tag-free and removes the legacy `bwrapStub` type from `registry.go`.

**Alternatives considered:**
- Keep `bwrapStub` and add a separate `RealBwrapIsolator` only on Linux.
  Rejected: leaves dead code on the macOS path and forces the registry to
  conditionally pick between two factories.
- Build-tag the entire registry. Rejected: would require duplicating the
  `BackendMode` constants and `SelectBackend` logic per platform.

### D2 — `compileBwrapArgs` is platform-agnostic

The `Policy → []string` argv compiler is in `bwrap_args.go` with no build
tag, so it compiles and runs unit tests on every platform.

**Rationale:** The mapping logic is the highest-risk part of the change and
benefits most from being unit-testable on macOS dev. The actual `Apply()`
method in `bwrap_linux.go` is a thin wrapper around the compiler.

### D3 — Filesystem mapping is default-deny + explicit allow

| Policy field             | bwrap argv                          | Notes                          |
|--------------------------|-------------------------------------|--------------------------------|
| `ReadOnlyGlobal: true`   | `--ro-bind / /`                     | Mounts entire root read-only   |
| `ReadPaths` (when !global) | `--ro-bind <p> <p>` per path     | Only when global=false         |
| `WritePaths`             | `--bind <p> <p>` per path           | rw overlay on the ro root      |
| `DenyPaths`              | `--tmpfs <p>` per path (dir only)   | bwrap has no native deny       |
| Standard mounts          | `--proc /proc --dev /dev --tmpfs /run` | Always present              |

**Directory-only deny:** bwrap's `--tmpfs` cannot mount over a regular file.
The compiler runs `os.Stat` on each `DenyPath` and returns an error if the
path is missing or not a directory. The only current caller is
`StrictToolPolicy(workDir)` which adds `<workDir>/.git` (always a
directory), so this constraint is met in practice. File-level deny (e.g.
`--ro-bind /dev/null <file>`) is deferred.

### D4 — Network mapping uses `--unshare-net`

| `Policy.Network` | bwrap argv                             |
|------------------|----------------------------------------|
| `NetworkDeny`    | `--unshare-net`                        |
| `NetworkAllow`   | (no flag — host network passthrough)   |
| `NetworkUnixOnly`| `--unshare-net` (caveat below)         |

`NetworkUnixOnly` is treated as deny because bwrap has no AF_UNIX-only
filter. AF_UNIX sockets that are reachable via the bound filesystem still
work because the socket file is shared via `--bind`/`--ro-bind` mounts.
`AllowedNetworkIPs` is ignored on Linux (bwrap has no AF_INET filter); the
existing PR 1 status warning that flags `allowedNetworkIPs` as macOS-only
remains in place.

### D5 — Process namespace defaults

Always-on bwrap flags:
- `--die-with-parent` — parent exit kills the child
- `--unshare-pid` — PID namespace isolation (security-critical)
- `--unshare-ipc` — SysV IPC isolation
- `--unshare-uts` — hostname isolation
- `--unshare-cgroup-try` — best-effort, ignored on kernels < 4.6
- `--proc /proc`, `--dev /dev`, `--tmpfs /run` — standard mounts

**Deliberately NOT used:**
- `--new-session` — would break PTY pass-through that
  `exec.Tool.RunWithPTY` relies on.
- `--unshare-user` — distro AppArmor / sysctl
  (`unprivileged_userns_clone=0`) frequently blocks unprivileged user
  namespaces. We rely on the standard setuid `bwrap` binary instead.
- `--unshare-net` (unconditional) — only added when `Policy.Network` is
  Deny or UnixOnly.

### D6 — Real Linux kernel probes (informational only)

`probeLandlockKernel()`:
- Calls `unix.Syscall(SYS_LANDLOCK_CREATE_RULESET, 0, 0,
  LANDLOCK_CREATE_RULESET_VERSION)`.
- Returns `(true, abi, "Landlock ABI N")` on success (kernel ≥ 5.13).
- Returns `(false, 0, "Landlock not supported by this kernel ...")` on
  ENOSYS, or `"Landlock probe failed: <errno>"` on other errors.

`probeSeccompKernel()`:
- Calls `unix.PrctlRetInt(PR_GET_SECCOMP, ...)`.
- Returns `(true, "seccomp interface present (PR_GET_SECCOMP=N); BPF
  filter capability not directly verified")` on success.
- Augments the reason with `/proc/self/status:Seccomp` when readable.

The `HasSeccomp` field's doc comment explicitly states it is a presence
signal only and does NOT prove BPF filter installability. We deliberately
do not call `PR_SET_SECCOMP` because that would permanently install a
filter on the probing process.

These probes do not influence backend selection (bwrap depends on the
external `bwrap` binary, not on Landlock or seccomp). They populate
`PlatformCapabilities` so `lango sandbox status` can report what the
kernel supports — and so a future native backend has accurate input.

### D7 — Resolved bwrap path is captured at probe time

`NewBwrapIsolator()` calls `exec.LookPath("bwrap")`, then `filepath.Abs`,
then `bwrap --version`, and stores all three in `BwrapIsolator{
resolvedPath, version, available, reason }`. `Apply()` sets
`cmd.Path = b.resolvedPath` and `cmd.Args[0] = b.resolvedPath` instead
of writing the bare string `"bwrap"`.

**Rationale:** Probe and exec are separated in time. Between them, the
process's `PATH` could change (e.g. tool execution may flip the
environment) or `cwd` could change. Storing the absolute path
guarantees both refer to the same binary.

### D8 — Deterministic network deny test via hidden self-subcommand

`runNetworkDenyTest`:
1. Opens an ephemeral `net.Listen("tcp", "127.0.0.1:0")` in the parent
   so we have a known-reachable target.
2. Re-invokes `os.Executable() sandbox _probe-net <addr>` as a sandboxed
   child.
3. The hidden `_probe-net` cobra subcommand calls
   `net.DialTimeout("tcp", arg, 2*time.Second)` and exits 0 on success
   or non-zero on failure (cobra propagates the error).
4. Parent reports PASS only when the child failed to connect.

**Why hidden self-subcommand instead of `nc` / `curl` / `bash` / `/dev/tcp`:**
- Minimal Docker images (e.g. `golang-go` slim, busybox-based images)
  often lack `nc`, `curl`, and bash builtin `/dev/tcp` (dash is the
  default shell). The hidden subcommand uses only the Go stdlib.
- The lango binary is already present in the sandbox via `--ro-bind / /`,
  so no extra setup is needed.
- Both bwrap `--unshare-net` and Seatbelt `(deny network*)` block
  loopback, so the same test case is deterministic on both platforms.

### D9 — Smoke test reliability fix (root cause)

The legacy `runReadTest` and `runWriteTest` used shell `>/dev/null 2>&1`
redirection. Seatbelt's `(deny default)` blocks `/dev/null` open-for-write,
so the shell exited non-zero before the actual test command ran. The read
test had been silently false-negative on macOS the entire time.

Fix: replace shell redirects with parent-side `io.Discard` via a small
`discardOutput(*exec.Cmd)` helper. `exec.Cmd` creates pipes for non-
`*os.File` writers, and those pipe FDs are already open before the
sandbox takes effect, so the child never tries to open `/dev/null`
inside the sandbox. `runReadTest` now invokes `/bin/cat` directly;
`runWriteTest` and `runWorkspaceWriteTest` invoke `/usr/bin/touch`
directly.

`runWorkspaceWriteTest` additionally resolves `os.MkdirTemp` results
through `filepath.EvalSymlinks` because Seatbelt matches subpaths
against the realpath (`/private/var/folders/...`) while `os.MkdirTemp`
returns the user-visible `/var/folders/...` symlink. Without
`EvalSymlinks` the policy and the actual write path do not match.

### D10 — Documentation honesty contract

CLAUDE.md requires that `internal/` changes update README, docs, and TUI
labels in the same commit. Stage 5 of this change touches:
`README.md`, `docs/cli/sandbox.md`, `docs/configuration.md`,
`internal/cli/settings/forms_sandbox.go`,
`internal/config/types_sandbox.go`, and
`internal/sandbox/os/isolator.go`. Items that remain unimplemented
(native backend, file-level deny, escape hardening, NormalizePaths
expansion) are explicitly marked as planned, not silently dropped.

## Risks / Trade-offs

- **Risk:** `golang.org/x/sys` does not expose `SYS_LANDLOCK_CREATE_RULESET`
  on the version pinned in `go.mod`. → **Mitigation:** Verified before
  implementation that `x/sys v0.42.0` exposes both
  `SYS_LANDLOCK_CREATE_RULESET = 444` and
  `LANDLOCK_CREATE_RULESET_VERSION = 0x1` on `linux/amd64`. If a future
  arch or x/sys downgrade loses the constant, the probe falls back to a
  reason string instead of returning `false` silently.
- **Risk:** Distro AppArmor blocks unprivileged user namespaces, causing
  bwrap to EPERM on `--unshare-user`. → **Mitigation:** D5 deliberately
  does NOT use `--unshare-user`. We rely on the setuid `bwrap` binary
  shipped by every distro's `bubblewrap` package.
- **Risk:** `bubblewrap < 0.4.0` (Debian 10, Ubuntu 18.04) does not
  support `--unshare-cgroup-try`. → **Mitigation:** The `-try` suffix
  makes that flag a best-effort no-op. If a future user reports it as
  hard-fail, we can parse `bwrap --version` and skip the flag.
- **Risk:** PTY pass-through breakage. `exec.Tool.RunWithPTY` is used
  for interactive tools. bwrap passes stdin/stdout/stderr unchanged
  unless `--new-session` is set. → **Mitigation:** D5 deliberately
  does NOT use `--new-session`. The smoke tests cover the non-PTY
  path; existing `RunWithPTY` unit tests cover the PTY path.
- **Risk:** Smoke test `runNetworkDenyTest` fails because the host
  test environment cannot bind 127.0.0.1 (e.g. extreme network sandbox
  on the host). → **Mitigation:** The test returns `false` (FAIL) in
  that case rather than panicking. The error path is the same as a
  real isolation failure.
- **Trade-off:** Linux users without `bubblewrap` installed get no
  isolation. They see an honest reason string from `lango sandbox
  status`, and `failClosed: true` will reject execution. We accept
  this because installing `bubblewrap` is one `apt install` command
  and the alternative (refusing to ship any Linux backend until
  native lands) leaves the entire Linux user base unsandboxed.

## Migration Plan

No data migration. Pure runtime behavior change with build-tag isolation:

1. macOS users see no behavior change. `PlatformBackendCandidates()` still
   registers bwrap in the second slot, but on darwin it returns the no-op
   stub from `bwrap_other.go` with `Reason()="bwrap is Linux-only"`.
2. Linux users with `bubblewrap` installed automatically get isolation when
   `sandbox.enabled: true` (the default backend `auto` selects bwrap).
3. Linux users without `bubblewrap` see a clear reason from
   `lango sandbox status` and continue running unsandboxed unless
   `failClosed: true`.
4. Rollback: revert the change. The previous noop behavior is fully
   restored because the registry, integration, and config layers were
   already in place from PR 1/2.

## Open Questions

None at archive time. Items deferred to follow-up changes:
- File-level deny via `Policy.DenyPaths`.
- Escape hardening (config files, skills directory, git metadata).
- Exception policy / excluded command patterns.
- `NormalizePaths` expansion to `sandbox.workspacePath` and
  `sandbox.allowedWritePaths`.
- Native Landlock+seccomp backend implementation (the probe lands here;
  the enforcement does not).
