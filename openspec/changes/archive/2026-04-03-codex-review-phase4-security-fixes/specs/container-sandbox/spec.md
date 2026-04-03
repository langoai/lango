## MODIFIED Requirements

### Requirement: P2P sandbox executor respects toolIsolation.enabled
The P2P sandbox executor SHALL only be wired when `cfg.P2P.ToolIsolation.Enabled` is true. When P2P is enabled but `toolIsolation.enabled` is false, the system SHALL log a startup warning explaining that inbound `tool_invoke` requests will be rejected, and the handler SHALL reject such requests with `ErrNoSandboxExecutor`.

#### Scenario: P2P enabled, toolIsolation disabled (default)
- **WHEN** `p2p.enabled=true` and `p2p.toolIsolation.enabled=false`
- **THEN** no sandbox executor SHALL be attached and a startup warning SHALL be logged

#### Scenario: P2P enabled, toolIsolation enabled
- **WHEN** `p2p.enabled=true` and `p2p.toolIsolation.enabled=true`
- **THEN** a sandbox executor SHALL be attached (container or subprocess based on config)
