## Purpose

Capability spec for os-sandbox-cli. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: sandbox status command output
`lango sandbox status` SHALL display: Sandbox Configuration (enabled, fail-mode explanation when enabled and not opted out, backend label, network mode, workspace), Active Isolation (isolator name, available, reason if unavailable), Platform Capabilities (platform, kernel, primitives), and **Backend Availability** (one row per platform candidate with available/unavailable status and reason). The capability formatter SHALL distinguish between `"unknown (probe not yet implemented)"` and `"unavailable (reason)"`.

When `sandbox.enabled=true` and `sandbox.backend=none`, status SHALL display `"Backend: none (explicit opt-out — fail-closed not applied)"` and SHALL NOT print the `Fail-Closed` line, accurately reflecting that the runtime skips fail-closed for this configuration.

#### Scenario: Backend Availability section present
- **WHEN** `lango sandbox status` runs
- **THEN** output contains a `Backend Availability:` header followed by one row per platform candidate using `ListBackends(PlatformBackendCandidates())`

#### Scenario: Auto resolved label
- **WHEN** `sandbox.backend=auto` and seatbelt is selected
- **THEN** status shows `"Backend: auto (resolved: seatbelt)"`

#### Scenario: backend=none opt-out display
- **WHEN** `sandbox.enabled=true` and `sandbox.backend=none`
- **THEN** status shows `"Backend: none (explicit opt-out — fail-closed not applied)"` and omits the Fail-Closed line

#### Scenario: Linux status with noop isolator
- **WHEN** `lango sandbox status` runs on Linux with no isolation backend
- **THEN** output shows `Isolator: noop` and the noop's `Reason()` field aggregates each candidate's reason

#### Scenario: macOS status with seatbelt
- **WHEN** `lango sandbox status` runs on macOS with sandbox-exec available
- **THEN** output shows `Isolator: seatbelt`, `Available: true`, and `Seatbelt: available (sandbox-exec found)`

#### Scenario: Fail-mode display
- **WHEN** sandbox is enabled with `failClosed=false` and not opted out
- **THEN** status shows `Fail-Closed: fail-open (warning + unsandboxed execution)`

#### Scenario: Status shows allowedNetworkIPs warning on Linux
- **WHEN** `lango sandbox status` is run on Linux with `allowedNetworkIPs` configured
- **THEN** output SHALL include a warning that `allowedNetworkIPs` is macOS-only

### Requirement: TUI settings descriptions
The OS Sandbox settings form and menu descriptions SHALL accurately reflect Linux enforcement status. Descriptions SHALL NOT claim Linux Landlock/seccomp enforcement when it is not implemented.

#### Scenario: Form description accuracy
- **WHEN** user views OS Sandbox settings form
- **THEN** enabled field description says "Seatbelt on macOS; Linux: planned, not yet enforced"
- **AND** seccomp profile field description says "Linux only — not yet enforced"
- **AND** menu description says "OS-level tool isolation (macOS enforced, Linux planned)"

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

### Requirement: TUI settings backend selection
The OS Sandbox settings form SHALL include an `os_sandbox_backend` field of type `InputSelect` with options `["auto", "seatbelt", "bwrap", "native", "none"]`. The TUI state-update layer SHALL map this field to `cfg.Sandbox.Backend`.

#### Scenario: Backend select field present
- **WHEN** the OS Sandbox form is rendered
- **THEN** it contains a select field keyed `os_sandbox_backend` with the five backend options

### Requirement: sandbox status Recent Decisions section
`lango sandbox status` SHALL render a `Recent Sandbox Decisions` section showing the most recent N=10 audit rows with `action="sandbox_decision"`. Each row SHALL display the timestamp, an 8-character session-key prefix in brackets (or `--------` when the audit row has no session key), the decision verdict, the backend that produced it (or `-` for non-applied verdicts), and the command target. When a row has a non-empty `reason` or `pattern` detail, it SHALL appear in parentheses at the end of the line.

The section SHALL be rendered only when an optional `BootLoader` dependency is wired into `NewSandboxCmd`. When the loader is `nil`, returns an error (DB locked, signed-out, or missing), or returns a result with no `DBClient`, the section SHALL be silently omitted so the status command remains usable as a pure sandbox-layer diagnostic without depending on audit availability.

`lango sandbox status` SHALL accept a `--session <prefix>` flag. When provided, the audit query SHALL filter rows whose `SessionKey` has that prefix. When omitted, the query SHALL return the global last 10 decisions across all sessions.

The audit DB client returned by the `BootLoader` SHALL NOT be closed by the status command — the bootstrap result owns the client and the cobra root is responsible for the process lifecycle.

#### Scenario: Recent Decisions section uses global last 10 by default
- **WHEN** `lango sandbox status` runs without `--session`
- **THEN** the `Recent Sandbox Decisions` section header SHALL contain `"global, last 10"`
- **AND** the audit query SHALL NOT include a session filter

#### Scenario: --session flag filters by prefix
- **WHEN** `lango sandbox status --session a3f1` runs
- **THEN** the audit query SHALL include `auditlog.SessionKeyHasPrefix("a3f1")`
- **AND** the section header SHALL contain `"session=a3f1"`

#### Scenario: Section omitted when BootLoader is nil
- **WHEN** `NewSandboxCmd` was called with a nil `BootLoader`
- **THEN** `lango sandbox status` SHALL render the rest of the status without panicking and SHALL NOT print a `Recent Sandbox Decisions` header

#### Scenario: Section omitted when BootLoader returns error
- **WHEN** the wired `BootLoader` returns an error (DB locked, signed-out, missing)
- **THEN** the status command SHALL silently skip the section and continue rendering

#### Scenario: Empty session key renders as dashes
- **WHEN** an audit row has an empty `SessionKey` (e.g. an MCP server startup decision)
- **THEN** the row SHALL render with `[--------]` in the session-prefix column

#### Scenario: Long session keys are truncated to 8 characters
- **WHEN** an audit row has a session key longer than 8 characters
- **THEN** only the first 8 characters SHALL appear in the session-prefix column

### Requirement: TUI settings excluded commands field
The OS Sandbox settings form SHALL include an `os_sandbox_excluded_commands` field of type `InputText` whose value is a comma-separated list of command basenames. The field's description SHALL state that excluded commands run UNSANDBOXED and that they are recorded in audit. The TUI state-update layer SHALL split the value on commas, trim whitespace, and store the result in `cfg.Sandbox.ExcludedCommands`.

#### Scenario: Excluded commands field present
- **WHEN** the OS Sandbox form is rendered
- **THEN** it contains a text field keyed `os_sandbox_excluded_commands` whose description warns that the listed commands run unsandboxed

#### Scenario: State update parses comma-separated values
- **WHEN** the user enters `git, docker , kubectl` in the field
- **THEN** `cfg.Sandbox.ExcludedCommands` SHALL be `["git", "docker", "kubectl"]` (whitespace trimmed)
