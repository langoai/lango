## ADDED Requirements

### Requirement: Agent Card endpoint
The system SHALL serve an Agent Card at `GET /.well-known/agent.json` when A2A is enabled, containing the agent's name, description, URL, and skills.

#### Scenario: Agent card served
- **WHEN** a GET request is made to `/.well-known/agent.json`
- **THEN** the response SHALL be JSON with `name`, `description`, `url`, and `skills` fields

#### Scenario: Skills derived from agent tree
- **WHEN** the agent has sub-agents (multi-agent mode)
- **THEN** each sub-agent SHALL appear as a skill in the Agent Card

### Requirement: A2A server route mounting
The A2A server SHALL mount its routes on the gateway's chi.Router when `a2a.enabled` and `agent.multiAgent` are both true.

#### Scenario: Routes mounted on gateway
- **WHEN** the application starts with both A2A and multi-agent enabled
- **THEN** the A2A server's RegisterRoutes SHALL be called with the gateway's Router

#### Scenario: A2A disabled
- **WHEN** `a2a.enabled` is false
- **THEN** no A2A routes SHALL be mounted

### Requirement: Gateway Router accessor
The gateway Server SHALL expose a `Router() chi.Router` method for external route mounting.

#### Scenario: Router method returns chi.Router
- **WHEN** `Router()` is called on the gateway server
- **THEN** it SHALL return the internal chi.Router instance

### Requirement: ADK Agent accessor
The adk.Agent SHALL expose an `ADKAgent()` method returning the underlying `adk_agent.Agent` for use by A2A server.

#### Scenario: ADKAgent returns underlying agent
- **WHEN** `ADKAgent()` is called on an adk.Agent created via NewAgent or NewAgentFromADK
- **THEN** it SHALL return the stored adk_agent.Agent instance

### Requirement: Remote Agent Loading Order
Remote A2A agents SHALL be loaded and assigned to `orchCfg.RemoteAgents` BEFORE calling `BuildAgentTree()`, ensuring they are included in the orchestrator's sub-agent list.

#### Scenario: A2A agents configured
- **WHEN** `cfg.A2A.Enabled` is true and remote agents are configured
- **THEN** remote agents SHALL be loaded and available in `orchCfg.RemoteAgents` before `BuildAgentTree()` is called

#### Scenario: A2A loading fails
- **WHEN** remote agent loading produces an error
- **THEN** the error SHALL be logged as a warning and the agent tree SHALL still be built without remote agents
