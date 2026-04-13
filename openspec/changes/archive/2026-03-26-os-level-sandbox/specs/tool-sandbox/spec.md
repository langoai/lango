## ADDED Requirements

### Requirement: OS-level sandbox config
The system SHALL provide a `SandboxConfig` at `config.Sandbox` with `Enabled`, `FailClosed`, `WorkspacePath`, `NetworkMode`, `AllowedNetworkIPs`, `AllowedWritePaths`, `TimeoutPerTool`, and `OS` (SeccompProfile, SeatbeltCustomProfile) fields, independent of `p2p.toolIsolation`.

#### Scenario: Default config values
- **WHEN** no sandbox config is provided
- **THEN** defaults SHALL be: enabled=false, failClosed=false, networkMode="deny", timeoutPerTool=30s, seccompProfile="moderate", allowedWritePaths=["/tmp"]
