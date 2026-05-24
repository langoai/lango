## MODIFIED Requirements

### Requirement: Remote Agent Loading Order
Remote A2A agents SHALL be loaded and assigned to `orchCfg.RemoteAgents` BEFORE calling `BuildAgentTree()`, ensuring successfully loaded remotes are included in the orchestrator's sub-agent list. When one or more configured remotes cannot be loaded, the loader SHALL return the successfully loaded remotes plus an error describing skipped remotes so startup can log a degraded remote-agent warning instead of silently omitting configured capacity.

#### Scenario: A2A agents configured
- **WHEN** `cfg.A2A.Enabled` is true and remote agents are configured
- **THEN** valid remote agents SHALL be loaded and available in `orchCfg.RemoteAgents` before `BuildAgentTree()` is called

#### Scenario: A2A loading partially fails
- **WHEN** multiple remote agents are configured and one remote is missing `agentCardUrl`
- **THEN** the loader SHALL return the successfully loaded remote agents
- **AND** it SHALL return a non-nil error that identifies the skipped remote

#### Scenario: A2A proxy construction fails
- **WHEN** a configured remote agent has an `agentCardUrl` but ADK remote proxy construction fails
- **THEN** the loader SHALL skip that remote
- **AND** it SHALL return a non-nil error identifying the remote and construction failure

#### Scenario: A2A loading fails for all remotes
- **WHEN** every configured remote agent is invalid or cannot be constructed
- **THEN** the loader SHALL return no remote agents
- **AND** it SHALL return a non-nil error describing the skipped remotes
- **AND** app startup SHALL continue to build the local agent tree while logging the remote-load warning
