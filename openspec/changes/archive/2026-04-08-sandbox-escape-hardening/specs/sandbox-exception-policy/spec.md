## ADDED Requirements

### Requirement: ExcludedCommands bypass via first-token basename match
The system SHALL support a `Sandbox.ExcludedCommands []string` config field. When the basename of the user command's first whitespace-separated token matches an entry, the exec tool SHALL skip applying the OS isolator and run the command unsandboxed. Matching SHALL operate on the user command string passed to `exec.Tool.Run` / `RunWithPTY` / `StartBackground` BEFORE secret token resolution and BEFORE the `sh -c` wrapping that those methods apply internally.

The matcher SHALL NOT use `cmd.Args[0]`. Because every exec.Tool execution path wraps the user command in `exec.CommandContext(ctx, "sh", "-c", resolved)`, `cmd.Args[0]` is always `"sh"` and would either match nothing or, worse, falsely bypass every command if `"sh"` were ever excluded.

The matcher SHALL be conservative with shell chains: only the FIRST whitespace-separated token's basename is checked. Chained commands such as `cd /tmp && git status` SHALL NOT match `ExcludedCommands=["git"]` because the first token is `cd`. This is the safe direction — explicit invocations bypass, indirect invocations stay sandboxed.

The matcher SHALL recognise paths inside the first token: `/usr/bin/git push` with `ExcludedCommands=["git"]` SHALL match because `filepath.Base("/usr/bin/git")` returns `"git"`.

`ExcludedCommands` SHALL be wired in `exec.Tool` only. Skill executor and MCP transport SHALL NOT consume `ExcludedCommands`.

#### Scenario: Direct invocation matches
- **WHEN** `Config.ExcludedCommands = ["git"]` and the user command is `"git status"`
- **THEN** `OSIsolator.Apply` SHALL NOT be called and a `SandboxDecisionEvent{Decision:"excluded", Pattern:"git"}` SHALL be published

#### Scenario: Absolute path invocation matches by basename
- **WHEN** `Config.ExcludedCommands = ["git"]` and the user command is `"/usr/bin/git push"`
- **THEN** the bypass SHALL fire because `filepath.Base("/usr/bin/git")` is `"git"`

#### Scenario: Chained commands do not match
- **WHEN** `Config.ExcludedCommands = ["git"]` and the user command is `"cd /tmp && git status"`
- **THEN** the bypass SHALL NOT fire because the first token is `"cd"` (safe direction)

#### Scenario: Pipe in user command does not break first-token match
- **WHEN** `Config.ExcludedCommands = ["git"]` and the user command is `"git status | grep foo"`
- **THEN** the bypass SHALL fire because the first token is still `"git"`

#### Scenario: Empty user command does not match
- **WHEN** the user command is empty or whitespace-only
- **THEN** the matcher SHALL return no match and the sandbox SHALL be applied normally

#### Scenario: Excluded does not match sh wrapper (regression guard)
- **WHEN** `Config.ExcludedCommands = ["sh"]` and the user command is `"echo hello"`
- **THEN** `OSIsolator.Apply` SHALL be called normally — the matcher MUST NOT consume `cmd.Args[0]`, which would be `"sh"`

### Requirement: SandboxDecisionEvent schema and helper
The system SHALL provide an `eventbus.SandboxDecisionEvent` struct with the following fields:

- `SessionKey string` — derived from `session.SessionKeyFromContext(ctx)` at publish time; may be empty for MCP startup events.
- `Source string` — one of `"exec"`, `"skill"`, `"mcp"` identifying which subsystem made the decision.
- `Command string` — user-facing command, skill name, or MCP server name.
- `Decision string` — one of `"applied"`, `"skipped"`, `"rejected"`, `"excluded"`.
- `Backend string` — `"bwrap"`, `"seatbelt"`, `"noop"`, or `""`.
- `Reason string` — empty for `"applied"`, populated otherwise (error message, "no isolator configured", etc.).
- `Pattern string` — populated only for `Decision="excluded"` (the matched ExcludedCommands entry).
- `Timestamp time.Time` — set automatically by `PublishSandboxDecision` if zero.

The system SHALL provide a helper `eventbus.PublishSandboxDecision(bus *Bus, evt SandboxDecisionEvent)` that:

1. Returns immediately when `bus == nil` so call sites in standalone tests can omit bus wiring without panicking.
2. Sets `evt.Timestamp = time.Now()` when the field is zero.
3. Calls `bus.Publish(evt)`.

#### Scenario: Helper is nil-safe
- **WHEN** `PublishSandboxDecision(nil, evt)` is called
- **THEN** the call SHALL return without panicking and without dispatching to any handler

#### Scenario: Helper sets timestamp
- **WHEN** `PublishSandboxDecision(bus, SandboxDecisionEvent{...})` is called with a zero `Timestamp`
- **THEN** the published event SHALL have `Timestamp` set to a non-zero value before dispatch

#### Scenario: EventName matches constant
- **WHEN** `SandboxDecisionEvent{}.EventName()` is called
- **THEN** it SHALL return the value of `EventSandboxDecision` (`"sandbox.decision"`)

### Requirement: Audit recorder subscribes to SandboxDecisionEvent
`audit.Recorder.Subscribe(bus)` SHALL register a typed handler for `eventbus.SandboxDecisionEvent`. The handler SHALL create one `AuditLog` row per event with:

- `Action = auditlog.ActionSandboxDecision` (`"sandbox_decision"` enum value, added to the ent schema in this change)
- `Actor = evt.Source` (one of `"exec"`, `"skill"`, `"mcp"`) — never empty so the schema's `actor.NotEmpty()` validator passes
- `Target = evt.Command`
- `Details` map containing `decision`, `source`, `backend`, plus optional `reason` and `pattern` keys
- `SessionKey` set ONLY when `evt.SessionKey != ""` — empty session keys (MCP startup) are persisted as NULL/empty so the row remains valid

The ent schema `audit_log.action` enum SHALL include the value `"sandbox_decision"` (added via `go generate ./internal/ent`).

#### Scenario: All four decision values recorded
- **WHEN** `SandboxDecisionEvent` is published with `Decision` in `{"applied", "skipped", "rejected", "excluded"}`
- **THEN** the audit recorder SHALL persist one `AuditLog` row per event with `Action="sandbox_decision"` and the appropriate details

#### Scenario: Empty SessionKey is not set on the row
- **WHEN** an MCP transport publishes a decision event with `SessionKey=""`
- **THEN** the audit row SHALL be created without calling `SetSessionKey` so the schema validator does not reject the row

#### Scenario: Source becomes the audit actor
- **WHEN** an event has `Source="exec"`
- **THEN** the persisted row's `actor` field SHALL equal `"exec"` (and similarly for `"skill"` and `"mcp"`)

### Requirement: Fail-open fallback emits one-shot stderr warning
When `exec.Tool.applySandbox` runs in fail-open mode (`Config.FailClosed=false`) and the isolator either is `nil` or returns an error from `Apply`, the tool SHALL print a single stderr line of the form:

```
lango: WARNING — sandbox fallback active (reason: <reason>); commands run unsandboxed
```

The warning SHALL be guarded by a `sync.Once` (per `Tool` instance) so subsequent fallbacks in the same process do not duplicate the message. The full per-command audit trail SHALL be available via `SandboxDecisionEvent{Decision:"skipped"}` and `lango sandbox status`.

#### Scenario: First fallback emits warning
- **WHEN** the first sandbox fallback occurs in a process
- **THEN** a single line matching the warning format SHALL be written to stderr

#### Scenario: Subsequent fallbacks do not duplicate
- **WHEN** a second sandbox fallback occurs in the same `Tool` instance
- **THEN** stderr SHALL NOT receive a second warning line

### Requirement: Three publish sites must record decisions
The system SHALL publish a `SandboxDecisionEvent` from every code path that calls `OSIsolator.Apply` (or short-circuits the apply due to ExcludedCommands or fail-closed-without-isolator). The three current sites are:

- `internal/tools/exec/exec.go` — `exec.Tool.applySandbox` (Source: `"exec"`)
- `internal/skill/executor.go` — `skill.Executor.executeScript` (Source: `"skill"`)
- `internal/mcp/connection.go` — `mcp.ServerConnection.createTransport` (Source: `"mcp"`)

Future code that adds a new sandbox apply site SHALL also publish a corresponding `SandboxDecisionEvent`. Inventory: `Grep "isolator.Apply"` at the start AND end of any sandbox-touching change to verify every match has a corresponding `PublishSandboxDecision` call.

#### Scenario: Skill apply publishes
- **WHEN** `skill.Executor.executeScript` calls `OSIsolator.Apply` and it returns nil
- **THEN** a `SandboxDecisionEvent{Source:"skill", Decision:"applied", Command:skill.Name}` SHALL be published

#### Scenario: Skill fail-closed publishes
- **WHEN** `skill.Executor.executeScript` calls `OSIsolator.Apply`, it returns an error, and `failClosed=true`
- **THEN** a `SandboxDecisionEvent{Source:"skill", Decision:"rejected"}` SHALL be published before `executeScript` returns the error

#### Scenario: MCP transport publishes
- **WHEN** `mcp.ServerConnection.createTransport` calls `OSIsolator.Apply` and it returns nil
- **THEN** a `SandboxDecisionEvent{Source:"mcp", Decision:"applied", Command:sc.name}` SHALL be published with empty SessionKey

### Requirement: Configuration field for excluded commands
The system SHALL provide a `SandboxConfig.ExcludedCommands []string` field. The field SHALL be wired through the supervisor's exec tool config. Its doc comment SHALL warn that excluded commands run unsandboxed and that matching is performed on the user command's first whitespace-separated token (NOT on `cmd.Args[0]`, which is always `"sh"`).

The TUI OS Sandbox settings form SHALL include a corresponding `os_sandbox_excluded_commands` field of type `InputText` accepting a comma-separated list. The TUI state-update layer SHALL split the value on commas, trim whitespace, and store the result in `cfg.Sandbox.ExcludedCommands`.

#### Scenario: Config field round-trips through supervisor
- **WHEN** `cfg.Sandbox.ExcludedCommands = ["git"]` and `supervisor.New(cfg)` is called
- **THEN** the constructed `exec.Tool.Config.ExcludedCommands` SHALL equal `["git"]`

#### Scenario: TUI form maps to config
- **WHEN** the user enters `git, docker , kubectl` in the `os_sandbox_excluded_commands` field
- **THEN** the state-update layer SHALL set `cfg.Sandbox.ExcludedCommands = ["git", "docker", "kubectl"]` (whitespace trimmed)
