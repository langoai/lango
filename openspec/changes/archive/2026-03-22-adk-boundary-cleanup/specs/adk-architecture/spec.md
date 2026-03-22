## MODIFIED Requirements

### Requirement: Ent State Adapter
The system SHALL adapt the Ent-based session store to the ADK State interface. FunctionCall/FunctionResponse conversion logic SHALL be consolidated into shared converter functions rather than duplicated across save and restore paths.

#### Scenario: Load Session
- **WHEN** ADK requests state for a session ID
- **THEN** the adapter SHALL retrieve the session and messages from Ent
- **AND** map them to ADK's expected message format using the shared converter

#### Scenario: Save Session
- **WHEN** ADK persists state updates
- **THEN** the adapter SHALL save new messages and state to Ent using the shared converter
- **AND** the in-memory session history SHALL be updated to reflect the persisted message

#### Scenario: FunctionCall round-trip fidelity
- **WHEN** a FunctionCall event is saved via `eventToMessage()` and restored via `EventsAdapter.All()`
- **THEN** the resulting genai.FunctionCall SHALL have identical ID, Name, Args, Thought, and ThoughtSignature as the original

#### Scenario: FunctionResponse round-trip fidelity
- **WHEN** a FunctionResponse event is saved and restored
- **THEN** the resulting genai.FunctionResponse SHALL have identical ID, Name, and Response as the original
- **AND** the role SHALL be "function" (not "user") per bug fix #1

#### Scenario: Get auto-create for new sessions
- **WHEN** `SessionServiceAdapter.Get()` is called for a session ID that does not exist
- **THEN** the adapter SHALL auto-create the session
- **AND** a comment SHALL document this deviation from ADK's `session.Service.Get()` contract which expects an error for missing sessions

#### Scenario: Get auto-renew for expired sessions
- **WHEN** `SessionServiceAdapter.Get()` is called for a session that has expired
- **THEN** the adapter SHALL delete the expired session and create a new one
- **AND** a comment SHALL document this deviation from ADK's standard Get behavior

## ADDED Requirements

### Requirement: Runner PluginConfig pass-through
The agent creation functions in `internal/adk/agent.go` SHALL accept an optional `PluginConfig` and forward it to `runner.Config.PluginConfig`.

#### Scenario: Agent created with plugins
- **WHEN** `NewAgent()` or `NewAgentStreaming()` is called with ADK plugin options
- **THEN** the `runner.Config.PluginConfig.Plugins` field SHALL contain the provided plugins
- **AND** the runner SHALL invoke plugin callbacks during lifecycle

#### Scenario: Agent created without plugins (default)
- **WHEN** `NewAgent()` or `NewAgentStreaming()` is called without plugin options
- **THEN** `runner.Config.PluginConfig` SHALL be zero value
- **AND** behavior SHALL be identical to current implementation
