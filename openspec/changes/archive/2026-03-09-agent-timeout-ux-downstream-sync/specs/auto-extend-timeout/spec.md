## ADDED Requirements

### Requirement: Auto-extend timeout config documented in README
The README.md config table SHALL include `agent.autoExtendTimeout` (bool, default `false`) and `agent.maxRequestTimeout` (duration, default 3× requestTimeout) rows after the `agent.agentsDir` row.

#### Scenario: User reads README config table
- **WHEN** a user views the README.md Agent configuration table
- **THEN** `agent.autoExtendTimeout` and `agent.maxRequestTimeout` rows are present with correct types and descriptions

### Requirement: Auto-extend timeout config documented in docs/configuration.md
The docs/configuration.md Agent section SHALL include both fields in the JSON example and the config table.

#### Scenario: JSON example includes new fields
- **WHEN** a user views the Agent JSON example in docs/configuration.md
- **THEN** `autoExtendTimeout` and `maxRequestTimeout` keys are present in the agent object

#### Scenario: Config table includes new fields
- **WHEN** a user views the Agent config table in docs/configuration.md
- **THEN** `agent.autoExtendTimeout` and `agent.maxRequestTimeout` rows are present after `agent.agentsDir`

### Requirement: TUI settings form includes auto-extend timeout fields
The Agent configuration form SHALL include an `auto_extend_timeout` boolean field and a `max_request_timeout` text field after the `tool_timeout` field.

#### Scenario: Agent form shows auto-extend fields
- **WHEN** user opens `lango settings` → Agent
- **THEN** "Auto-Extend Timeout" (bool) and "Max Request Timeout" (text) fields are displayed

### Requirement: TUI state update handles auto-extend timeout fields
The ConfigState.UpdateConfigFromForm SHALL handle `auto_extend_timeout` and `max_request_timeout` field keys, updating `Agent.AutoExtendTimeout` and `Agent.MaxRequestTimeout` respectively.

#### Scenario: State update processes auto_extend_timeout
- **WHEN** form field `auto_extend_timeout` has value `"true"`
- **THEN** `Agent.AutoExtendTimeout` is set to `true`

#### Scenario: State update processes max_request_timeout
- **WHEN** form field `max_request_timeout` has value `"15m"`
- **THEN** `Agent.MaxRequestTimeout` is set to 15 minutes

### Requirement: WebSocket events documented
The docs/gateway/websocket.md events table SHALL include `agent.progress`, `agent.warning`, and `agent.error` events.

#### Scenario: User views WebSocket events
- **WHEN** a user views the WebSocket events table
- **THEN** `agent.progress`, `agent.warning`, and `agent.error` events are listed with payload descriptions
