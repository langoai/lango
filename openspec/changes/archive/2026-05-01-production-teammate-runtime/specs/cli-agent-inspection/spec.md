## MODIFIED Requirements

### Requirement: Agent status command

The existing `lango agent status` command contract remains preserved unless explicitly changed by this requirement: current status fields continue to be shown in table and JSON output, and this change only adds teammate runtime reporting. The command SHALL expose `teammate_runtime` when multi-agent mode is enabled. For the production in-process teammate runtime defined by this change, `dynamic-v1` means the built-in teammate runtime path is configured and available for built-in teammates under multi-agent mode; it does not imply that legacy static fallback or remote A2A paths are disabled. If that built-in runtime path is not configured or not available, the command SHALL omit the `teammate_runtime` field rather than reporting `dynamic-v1`.

#### Scenario: Table output shows dynamic teammate runtime
- **WHEN** `lango agent status` is run with `agent.multiAgent: true`
- **AND** the built-in dynamic teammate runtime path is configured and available
- **THEN** the command SHALL display a `Teammate Runtime` field with value `dynamic-v1`

#### Scenario: JSON output shows dynamic teammate runtime
- **WHEN** `lango agent status --json` is run with `agent.multiAgent: true`
- **AND** the built-in dynamic teammate runtime path is configured and available
- **THEN** the output SHALL include `"teammate_runtime": "dynamic-v1"`

#### Scenario: Single-agent output omits teammate runtime
- **WHEN** `lango agent status` is run with `agent.multiAgent: false`
- **THEN** the command SHALL omit the teammate runtime field

#### Scenario: Multi-agent without built-in dynamic runtime omits teammate runtime
- **WHEN** `lango agent status` is run with `agent.multiAgent: true`
- **AND** the built-in dynamic teammate runtime path is not configured or not available
- **THEN** the command SHALL omit the teammate runtime field
