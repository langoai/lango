## MODIFIED Requirements

### Requirement: Agent list displays registry sources
The `lango agent list` command SHALL load agents from the dynamic agent registry (embedded + user-defined stores) instead of hardcoded lists. Each agent entry SHALL display its source: "builtin", "embedded", "user", or "remote".

#### Scenario: List shows embedded agents
- **WHEN** `lango agent list` is run with no user-defined agents
- **THEN** it SHALL display the 7 default agents with source "embedded"

#### Scenario: List shows user-defined agents
- **WHEN** user-defined agents exist in the configured agents directory
- **THEN** they SHALL appear in the list with source "user"

#### Scenario: List shows remote A2A agents
- **WHEN** A2A remote agents are configured
- **THEN** they SHALL appear in a separate table with source "a2a" and URL

#### Scenario: JSON output includes source
- **WHEN** `lango agent list --json` is run
- **THEN** each entry SHALL include "type" ("local" or "remote") and "source" fields

### Requirement: Agent status shows registry info
The `lango agent status` command SHALL display registry information including builtin agent count, user agent count, active agent count, and agents directory path.

#### Scenario: Status includes registry counts
- **WHEN** `lango agent status` is run
- **THEN** it SHALL display "Builtin Agents", "User Agents", "Active Agents" counts

#### Scenario: Status shows P2P and hooks status
- **WHEN** `lango agent status` is run
- **THEN** it SHALL display P2P enabled status and Hooks enabled status

#### Scenario: JSON status includes registry
- **WHEN** `lango agent status --json` is run
- **THEN** the output SHALL include a "registry" object with builtin, user, active counts
