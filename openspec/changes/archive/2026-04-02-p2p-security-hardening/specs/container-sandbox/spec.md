## ADDED Requirements

### Requirement: Fail-closed container enforcement
The `ContainerExecutor` MUST support a `requireContainer` mode. When enabled and the runtime resolves to `NativeRuntime` in auto mode, the executor MUST return an error wrapping `ErrRuntimeUnavailable` instead of silently falling back.

#### Scenario: RequireContainer with Docker available
- **WHEN** `requireContainer` is true and Docker is available
- **THEN** Docker runtime is selected normally

#### Scenario: RequireContainer without Docker
- **WHEN** `requireContainer` is true and no container runtime is available
- **THEN** `NewContainerExecutor` returns an error wrapping `ErrRuntimeUnavailable`
- **AND** `NativeRuntime` is NOT used as fallback

#### Scenario: RequireContainer disabled (backward compatible)
- **WHEN** `requireContainer` is false
- **THEN** fallback to `NativeRuntime` proceeds as before

## MODIFIED Requirements

### Requirement: ContainerExecutor runtime probe
`NewContainerExecutor` MUST check the `requireContainer` config field after the probe chain. If true and only `NativeRuntime` is available, it MUST return an error instead of proceeding.

### Requirement: Container sandbox configuration
The `p2p.toolIsolation.container` configuration block MUST include a `requireContainer` boolean field (default: `true` for new installations).
