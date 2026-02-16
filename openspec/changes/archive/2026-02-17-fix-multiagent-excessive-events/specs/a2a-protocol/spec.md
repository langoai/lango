## MODIFIED Requirements

### Requirement: Remote Agent Loading Order
Remote A2A agents SHALL be loaded and assigned to `orchCfg.RemoteAgents` BEFORE calling `BuildAgentTree()`, ensuring they are included in the orchestrator's sub-agent list.

#### Scenario: A2A agents configured
- **WHEN** `cfg.A2A.Enabled` is true and remote agents are configured
- **THEN** remote agents SHALL be loaded and available in `orchCfg.RemoteAgents` before `BuildAgentTree()` is called

#### Scenario: A2A loading fails
- **WHEN** remote agent loading produces an error
- **THEN** the error SHALL be logged as a warning and the agent tree SHALL still be built without remote agents
