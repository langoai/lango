## ADDED Requirements

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
