## ADDED Requirements

### Requirement: Shell command execution
The system SHALL execute shell commands in a controlled environment with configurable timeouts.

#### Scenario: Synchronous command execution
- **WHEN** a command is executed with a timeout
- **THEN** the system SHALL run the command and return stdout, stderr, and exit code

#### Scenario: Command timeout
- **WHEN** a command exceeds its timeout duration
- **THEN** the process SHALL be terminated and a timeout error returned

### Requirement: PTY support
The system SHALL support pseudo-terminal (PTY) mode for interactive commands.

#### Scenario: PTY command execution
- **WHEN** a command requires PTY (e.g., interactive prompts)
- **THEN** the system SHALL allocate a PTY and capture output

#### Scenario: ANSI escape handling
- **WHEN** PTY output contains ANSI escape codes
- **THEN** the codes SHALL be preserved for rendering or stripped as configured

### Requirement: Background process management
The system SHALL support running commands in the background with process tracking. Background process output SHALL be thread-safe for concurrent read/write access.

#### Scenario: Background execution
- **WHEN** a command is started in background mode
- **THEN** a session ID SHALL be returned for later status checks

#### Scenario: Background process status
- **WHEN** status is requested for a background process
- **THEN** current output and execution state SHALL be returned

#### Scenario: Concurrent output access
- **WHEN** a background process is writing output while status is being read
- **THEN** the output buffer SHALL be safely accessible without data races

### Requirement: Working directory control
The system SHALL execute commands in a specified working directory.

#### Scenario: Custom working directory
- **WHEN** a working directory is specified
- **THEN** the command SHALL execute relative to that directory

#### Scenario: Invalid working directory
- **WHEN** the specified directory does not exist
- **THEN** an error SHALL be returned before execution

### Requirement: Environment variable handling
The system SHALL control environment variables passed to child processes.

#### Scenario: Custom environment
- **WHEN** custom environment variables are specified
- **THEN** they SHALL be merged with or replace the base environment

#### Scenario: Dangerous variable filtering
- **WHEN** dangerous environment variables (LD_PRELOAD, etc.) are present
- **THEN** they SHALL be filtered out for security

#### Scenario: LANGO_PASSPHRASE filtered
- **WHEN** an agent executes a command and `LANGO_PASSPHRASE` is set in the parent environment
- **THEN** `LANGO_PASSPHRASE` is not passed to the child process

### Requirement: Enhanced execution feedback
The system SHALL provide more descriptive feedback when commands fail or time out.

#### Scenario: Detailed failure message
- **WHEN** a command fails with a non-zero exit code
- **THEN** the system SHALL return both stdout and stderr to the agent for debugging

### Requirement: Reference token resolution in exec
The exec tool SHALL resolve secret reference tokens in command strings immediately before execution. Resolved values SHALL never be logged or returned to the agent.

#### Scenario: Command with secret reference
- **WHEN** exec is called with command `curl -H "Auth: {{secret:api-key}}" https://api.example.com`
- **AND** the RefStore contains a value for `{{secret:api-key}}`
- **THEN** the token SHALL be replaced with the actual secret value before shell execution
- **AND** the log entry SHALL contain the original command with the unresolved token
- **AND** the BackgroundProcess.Command field SHALL contain the original command with the unresolved token

#### Scenario: Command with decrypt reference
- **WHEN** exec is called with command `echo {{decrypt:uuid-123}}`
- **AND** the RefStore contains a value for `{{decrypt:uuid-123}}`
- **THEN** the token SHALL be replaced with the actual decrypted value before shell execution

#### Scenario: Command with unknown reference
- **WHEN** exec is called with command `echo {{secret:unknown}}`
- **AND** the RefStore does NOT contain a value for `{{secret:unknown}}`
- **THEN** the literal string `{{secret:unknown}}` SHALL be passed to the shell unchanged

#### Scenario: Command without references
- **WHEN** exec is called with a command containing no reference tokens
- **THEN** the command SHALL be executed unchanged

#### Scenario: Reference resolution in PTY mode
- **WHEN** RunWithPTY is called with a command containing reference tokens
- **THEN** tokens SHALL be resolved identically to synchronous execution

#### Scenario: Reference resolution in background mode
- **WHEN** StartBackground is called with a command containing reference tokens
- **THEN** tokens SHALL be resolved identically to synchronous execution

### Requirement: Lango CLI block message includes builtin_list hint
The `blockLangoExec` catch-all message for unrecognized `lango` subcommands SHALL include a hint to use `builtin_list` for tool discovery.

#### Scenario: Catch-all message with builtin_list hint
- **WHEN** an unrecognized `lango` subcommand is blocked by `blockLangoExec`
- **THEN** the returned message SHALL contain "builtin_list"
- **AND** SHALL suggest using built-in tools or asking the user to run the command directly

### Requirement: Block lango automation commands via exec
The exec and exec_bg tool handlers SHALL detect and block commands that attempt to invoke lango CLI automation subcommands (cron, bg, background, workflow).

#### Scenario: Block lango cron command
- **WHEN** an exec or exec_bg tool receives a command starting with "lango cron"
- **THEN** the tool SHALL return a structured response with blocked=true and a message guiding to use built-in cron tools instead, without executing the command

#### Scenario: Block lango bg command
- **WHEN** an exec or exec_bg tool receives a command starting with "lango bg" or "lango background"
- **THEN** the tool SHALL return a structured response with blocked=true and a message guiding to use built-in background tools instead

#### Scenario: Block lango workflow command
- **WHEN** an exec or exec_bg tool receives a command starting with "lango workflow"
- **THEN** the tool SHALL return a structured response with blocked=true and a message guiding to use built-in workflow tools instead

#### Scenario: Context-aware guidance when feature is disabled
- **WHEN** an exec tool blocks a lango automation command and the corresponding feature is not enabled
- **THEN** the guidance message SHALL instruct the user to enable the feature in Settings

#### Scenario: Allow non-lango commands
- **WHEN** an exec or exec_bg tool receives a command that does not start with "lango cron", "lango bg", "lango background", or "lango workflow"
- **THEN** the tool SHALL execute the command normally without blocking

### Requirement: Exec handlers return typed BlockedResult
The exec and exec_bg handlers SHALL return a `BlockedResult` struct instead of `map[string]interface{}` when a command is blocked. The struct SHALL have `Blocked bool` and `Message string` fields with JSON tags.

#### Scenario: Blocked command returns BlockedResult
- **WHEN** exec handler blocks a command via blockLangoExec or blockProtectedPaths
- **THEN** handler returns `&BlockedResult{Blocked: true, Message: reason}`

### Requirement: Exec handlers integrate CommandGuard
The exec and exec_bg handlers SHALL call `blockProtectedPaths` after `blockLangoExec`. The CommandGuard SHALL be constructed in `app.New()` with DataRoot and AdditionalProtectedPaths, then passed through `buildTools` → `buildExecTools`.

#### Scenario: Guard blocks protected path access
- **WHEN** agent executes `sqlite3 ~/.lango/lango.db` via exec tool
- **THEN** handler returns BlockedResult before reaching the Supervisor

#### Scenario: Guard allows normal commands
- **WHEN** agent executes `go build ./...` via exec tool
- **THEN** command passes all guards and executes normally

### Requirement: WithPolicy middleware enforces policy before approval
The `WithPolicy` middleware SHALL be the outermost middleware in the tool chain, applied after `WithApproval` in `app.go`. The execution order SHALL be: `WithPolicy → WithApproval → WithPrincipal → WithHooks → ... → Handler`. It evaluates only `exec` and `exec_bg` tools.

#### Scenario: Policy blocks before approval
- **WHEN** `WithPolicy` evaluates a command that results in a block verdict
- **THEN** it returns `BlockedResult{Blocked: true}` without invoking the approval middleware

#### Scenario: Non-exec tools pass through
- **WHEN** `WithPolicy` receives `exec_status` or `exec_stop`
- **THEN** it calls the next handler without policy evaluation

### Requirement: Handler guards serve as defense-in-depth
The existing handler guards in `BuildTools` (`langoGuard`, `pathGuard`) SHALL be preserved as deterministic-only fallback. They check a strict subset of what the middleware checks. New rules are added only to `PolicyEvaluator`.

#### Scenario: Handler guards remain active
- **WHEN** a command passes `WithPolicy` middleware
- **THEN** handler guards still run inside `checkGuards`

### Requirement: PolicyDecisionEvent published for policy decisions
A `PolicyDecisionEvent` SHALL be published to the event bus for observe and block verdicts. Allow verdicts are not published. Publishing respects `cfg.Hooks.EventPublishing` gate.

#### Scenario: Event published for block verdict
- **WHEN** PolicyEvaluator blocks a command and event publishing is enabled
- **THEN** a `PolicyDecisionEvent` with verdict="block" is published

#### Scenario: No event when publishing disabled
- **WHEN** PolicyEvaluator makes a decision and `cfg.Hooks.EventPublishing` is false
- **THEN** no event is published

## OS-Level Sandbox

### Requirement: Command execution with OS sandbox
The exec tool SHALL execute shell commands via `sh -c` with configurable timeout, environment filtering, and optional OS-level sandbox isolation applied before process start.

#### Scenario: Run with sandbox enabled
- **WHEN** a command is executed via `Run()` with `Config.OSIsolator` set
- **THEN** the child process SHALL run under OS-level kernel restrictions per `Config.SandboxPolicy`

#### Scenario: Run with sandbox disabled
- **WHEN** a command is executed via `Run()` with `Config.OSIsolator` nil
- **THEN** the child process SHALL run without OS-level restrictions (existing behavior)

#### Scenario: Background process with sandbox
- **WHEN** a background command is started via `StartBackground()` with `Config.OSIsolator` set
- **THEN** the child process SHALL run under OS-level kernel restrictions
