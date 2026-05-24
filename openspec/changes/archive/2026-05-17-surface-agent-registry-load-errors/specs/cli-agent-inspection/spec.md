## MODIFIED Requirements

### Requirement: Agent list displays registry sources
The `lango agent list` command SHALL load agents from the dynamic agent registry (embedded + user-defined stores) instead of hardcoded lists. Each agent entry SHALL display its source: "builtin", "embedded", "user", or "remote". The command SHALL support `--output table|json` and `--check` flags. If a configured user-defined agent store contains a present `AGENT.md` file that cannot be read or parsed, the command SHALL return an actionable error instead of silently omitting user-defined agents.

#### Scenario: Agent list output uses the command writer
- **WHEN** `lango agent list` renders text or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: List shows embedded agents
- **WHEN** `lango agent list` is run with no user-defined agents
- **THEN** it SHALL display the 8 default agents with source "embedded"

#### Scenario: List shows user-defined agents
- **WHEN** user-defined agents exist in the configured agents directory
- **THEN** they SHALL appear in the list with source "user"

#### Scenario: Invalid user-defined agent file fails visibly
- **WHEN** `agent.agentsDir` points to a directory containing an invalid `AGENT.md`
- **THEN** `lango agent list` SHALL return an error
- **AND** the error SHALL identify the user agent load failure and file path
- **AND** the command SHALL NOT silently omit the user-defined agent definitions

### Requirement: Agent status shows registry info
The `lango agent status` command SHALL display registry information including builtin agent count, user agent count, active agent count, and agents directory path. If embedded or configured user-defined registry loading fails, the command SHALL return an actionable error instead of reporting misleading registry counts.

#### Scenario: Status includes registry counts
- **WHEN** `lango agent status` is run
- **THEN** it SHALL display "Builtin Agents", "User Agents", "Active Agents" counts

#### Scenario: JSON status includes registry
- **WHEN** `lango agent status --output json` is run
- **THEN** the output SHALL include a "registry" object with builtin, user, active counts

#### Scenario: Invalid user-defined agent file fails status visibly
- **WHEN** `agent.agentsDir` points to a directory containing an invalid `AGENT.md`
- **THEN** `lango agent status` SHALL return an error
- **AND** the error SHALL identify the user agent load failure and file path
- **AND** the command SHALL NOT report registry counts based on a partial load
