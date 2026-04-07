## MODIFIED Requirements

### Requirement: sandbox test command honors configured backend
`lango sandbox test` SHALL accept a `cfgLoader` callback and use `ParseBackendMode(cfg.Sandbox.Backend) + SelectBackend(mode, PlatformBackendCandidates())` instead of `NewOSIsolator()`. When the configured backend is `none`, test SHALL print a message indicating no isolation to test and exit successfully without running smoke tests. When the configured backend is unavailable, test SHALL print the backend name and reason and exit successfully.

When the configured backend is available, the command SHALL run four smoke test cases against the active isolator and print PASS/FAIL for each: write restriction (deny `/etc`), read permission (allow a system file), workspace write (allow a temp dir), and network deny (loopback unreachable). The command SHALL also print the backend name, the resolved mode, and (when the isolator implements an optional `Version() string` interface) the captured version string.

The four smoke test helpers SHALL silence child stdout/stderr via Go-side `io.Discard` (`exec.Cmd.Stdout = io.Discard; exec.Cmd.Stderr = io.Discard`) instead of shell `>/dev/null 2>&1` redirection. Reason: shell redirection causes the child shell to open `/dev/null` for write inside the sandbox, and Seatbelt's `(deny default)` blocks that open, producing false negatives. Parent-side `io.Discard` works because `exec.Cmd` creates pipe FDs in the parent before the sandbox takes effect.

#### Scenario: Test on platform with available backend
- **WHEN** `lango sandbox test` is run with an available backend
- **THEN** it SHALL run all four cases (write restriction, read permission, workspace write, network deny) and print PASS or FAIL for each

#### Scenario: backend=none short-circuits test
- **WHEN** `sandbox.backend=none` and `lango sandbox test` runs
- **THEN** output contains `"no isolation to test"` and the command exits successfully without running write/read tests

#### Scenario: Unavailable backend reports reason
- **WHEN** `sandbox.backend=bwrap` is configured but the bwrap binary is not installed
- **THEN** output contains `"Sandbox backend bwrap not available"` and the bwrap reason

#### Scenario: Test command uses io.Discard not shell redirection
- **WHEN** any of the four smoke test helpers (`runWriteTest`, `runReadTest`, `runWorkspaceWriteTest`, `runNetworkDenyTest`) is invoked
- **THEN** it SHALL set `cmd.Stdout = io.Discard` and `cmd.Stderr = io.Discard` and SHALL NOT pass `>/dev/null` or `2>/dev/null` to a shell

#### Scenario: Read test invokes binary directly
- **WHEN** `runReadTest` runs on macOS
- **THEN** it SHALL invoke `/bin/cat` directly (not `/bin/sh -c "cat ..."`)

#### Scenario: Version string displayed when available
- **WHEN** the active isolator implements `interface { Version() string }` and returns a non-empty value
- **THEN** the test command output SHALL contain a `Version: <value>` line

## ADDED Requirements

### Requirement: sandbox workspace write smoke test
`lango sandbox test` SHALL include a `runWorkspaceWriteTest` case that creates an `os.MkdirTemp` directory, resolves it through `filepath.EvalSymlinks` (so macOS Seatbelt's realpath matching works), adds the resolved path to a policy's `WritePaths`, and executes `/usr/bin/touch <dir>/probe.txt` under the sandbox. The case SHALL PASS only when the touched file exists after the command completes.

#### Scenario: Workspace write succeeds in tmp dir
- **WHEN** `runWorkspaceWriteTest` runs against an available isolator on a healthy host
- **THEN** the case SHALL return true and the probe file SHALL exist

#### Scenario: Realpath resolution applied
- **WHEN** `runWorkspaceWriteTest` runs on macOS where `os.MkdirTemp` returns `/var/folders/...`
- **THEN** the policy SHALL receive the path resolved by `filepath.EvalSymlinks` (e.g. `/private/var/folders/...`) so Seatbelt subpath matching succeeds

### Requirement: sandbox network deny smoke test via hidden self-subcommand
`lango sandbox test` SHALL include a `runNetworkDenyTest` case that opens an ephemeral `127.0.0.1:0` TCP listener in the parent process, then re-invokes the lango binary as a sandboxed child via `exec.Command(os.Executable(), "sandbox", "_probe-net", target)`. The case SHALL PASS only when the child fails to connect to the loopback target.

The hidden `lango sandbox _probe-net <addr>` cobra subcommand SHALL be registered with `Hidden: true`, accept exactly one address argument, call `net.DialTimeout("tcp", addr, 2*time.Second)`, and exit with status 0 on success or non-zero on failure (the cobra error path achieves this).

The network deny test SHALL NOT depend on any external tools (`nc`, `curl`, `bash`, `/dev/tcp`).

#### Scenario: Network deny test PASSes when sandbox blocks loopback
- **WHEN** `runNetworkDenyTest` runs against an available isolator with `NetworkDeny` policy
- **THEN** the sandboxed child SHALL fail to connect to the parent's 127.0.0.1 listener and the case SHALL return true

#### Scenario: Hidden subcommand exits non-zero on connect failure
- **WHEN** `lango sandbox _probe-net <unreachable-addr>` is invoked
- **THEN** the process SHALL exit with non-zero status

#### Scenario: Hidden subcommand exits zero on connect success
- **WHEN** `lango sandbox _probe-net <reachable-addr>` is invoked outside any sandbox
- **THEN** the process SHALL exit with status 0

#### Scenario: Network deny test uses no external tools
- **WHEN** `runNetworkDenyTest` runs in a minimal container without `nc`, `curl`, `bash`, or `/dev/tcp` support
- **THEN** the test SHALL still execute correctly because it depends only on the lango binary itself
