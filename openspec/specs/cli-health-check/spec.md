## Purpose

Provide a built-in CLI health check command that eliminates the need for external tools like curl in Docker health checks.
## Requirements
### Requirement: CLI health check command
The system SHALL provide a `lango health` CLI command that checks the gateway health endpoint without external dependencies.

#### Scenario: Failed health check does not emit success payload
- **WHEN** `lango health` returns a non-200 status or times out
- **THEN** it SHALL return an error without emitting the `ok` success payload

#### Scenario: Failed health check does not duplicate error output through the command writer
- **WHEN** `lango health` returns a non-200 status or times out
- **THEN** the command output stream SHALL remain empty
- **AND** the returned error SHALL still describe the failure

### Requirement: Advanced feature hints in onboard flow
The onboard flow SHALL display hints about advanced features after initial setup is complete. The hints SHALL inform users about agent memory, hooks, librarian, and learning system features that can be configured via settings or CLI.

#### Scenario: Onboard completion hints
- **WHEN** user completes the onboard wizard successfully
- **THEN** system displays hints mentioning:
  - Agent memory configuration via `lango memory agents` or TUI settings
  - Hook system configuration via `lango agent hooks` or TUI settings
  - Librarian configuration via `lango librarian status`

### Requirement: Feature discovery in doctor output
The doctor command output SHALL include brief hints about new CLI commands when relevant checks pass or are skipped, to aid feature discovery.

#### Scenario: Graph check with hint
- **WHEN** GraphStoreCheck returns StatusSkip because graph is disabled
- **THEN** the check message SHALL mention that graph can be enabled and managed via `lango graph` commands

#### Scenario: Multi-agent check with hint
- **WHEN** MultiAgentCheck returns StatusSkip because multi-agent is disabled
- **THEN** the check message SHALL mention that multi-agent can be configured via settings

### Requirement: Existing onboard flow unaffected
The addition of feature hints SHALL NOT change the core onboard flow steps or validation logic. Hints are displayed only after successful completion.

#### Scenario: Onboard steps unchanged
- **WHEN** user runs `lango onboard`
- **THEN** all existing onboard steps (provider selection, API key, channel setup) function identically to before the hint additions

### Requirement: Doctor recent-failure output includes cause metadata
The doctor command SHALL show classified cause metadata for recent failed multi-agent traces.

#### Scenario: Multi-agent check shows classified failures
- **WHEN** recent failed traces exist
- **THEN** the `Multi-Agent` doctor check SHALL include `trace_id`, `outcome`, `error_code`, `cause_class`, and `summary`

#### Scenario: Doctor JSON preserves cause metadata
- **WHEN** `lango doctor --output json` reports recent failed multi-agent traces
- **THEN** the same classified fields SHALL be present in machine-readable output
