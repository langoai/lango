## ADDED Requirements

### Requirement: Container sandbox configuration
The system MUST support a `p2p.toolIsolation.container` configuration block with `enabled`, `runtime`, `image`, `networkMode`, `readOnlyRootfs`, `cpuQuotaUs`, `poolSize`, and `poolIdleTimeout` fields.

#### Scenario: Default configuration
- **WHEN** no container config is specified
- **THEN** defaults are: `runtime: "auto"`, `image: "lango-sandbox:latest"`, `networkMode: "none"`, `readOnlyRootfs: true`, `poolSize: 0`, `poolIdleTimeout: 5m`

### Requirement: ContainerRuntime interface
The system MUST define a `ContainerRuntime` interface with `Run(ctx, ContainerConfig)`, `Cleanup(ctx, id)`, `IsAvailable(ctx)`, and `Name()` methods.

### Requirement: Error types
The system MUST define sentinel errors: `ErrRuntimeUnavailable`, `ErrContainerTimeout`, `ErrContainerOOM`.

#### Scenario: OOM kill
- **WHEN** a container exits with code 137 (SIGKILL)
- **THEN** `ErrContainerOOM` is returned

#### Scenario: Timeout
- **WHEN** container execution exceeds the configured timeout
- **THEN** `ErrContainerTimeout` is returned

### Requirement: DockerRuntime
The system MUST implement `ContainerRuntime` using Docker Go SDK with container create, attach, start, stdin write, stdout read, wait, and force-remove lifecycle.

#### Scenario: Container creation
- **WHEN** `Run` is called
- **THEN** a container is created with the configured image, `--sandbox-worker` command, labels `lango.sandbox=true` and `lango.tool=<name>`, resource limits, network mode, read-only rootfs, and tmpfs `/tmp`

#### Scenario: Docker unavailable
- **WHEN** `IsAvailable()` is called and Docker daemon is not reachable
- **THEN** returns `false`

#### Scenario: Orphan cleanup
- **WHEN** `Cleanup` is called
- **THEN** all containers with label `lango.sandbox=true` are force-removed

### Requirement: NativeRuntime fallback
The system MUST provide a `NativeRuntime` that wraps `SubprocessExecutor` as a `ContainerRuntime` implementation. It MUST always report `IsAvailable() = true`.

### Requirement: GVisorRuntime stub
The system MUST provide a `GVisorRuntime` stub that always reports `IsAvailable() = false` and returns `ErrRuntimeUnavailable` on `Run`.

### Requirement: ContainerExecutor runtime probe
`NewContainerExecutor` MUST probe runtimes in order: Docker → gVisor → Native. The first available runtime is used.

#### Scenario: Auto mode with Docker available
- **WHEN** runtime is "auto" and Docker is available
- **THEN** Docker runtime is selected

#### Scenario: Auto mode without Docker
- **WHEN** runtime is "auto" and Docker is unavailable
- **THEN** Native runtime is selected as fallback

#### Scenario: Explicit runtime requested but unavailable
- **WHEN** runtime is "docker" but Docker is unavailable
- **THEN** an error wrapping `ErrRuntimeUnavailable` is returned

### Requirement: Protocol version
`ExecutionRequest` MUST include an optional `version` field (default 0) for forward compatibility.

### Requirement: App wiring
The P2P sandbox executor SHALL only be wired when `cfg.P2P.ToolIsolation.Enabled` is true. When P2P is enabled but `toolIsolation.enabled` is false, the system SHALL log a startup warning explaining that inbound `tool_invoke` requests will be rejected, and the handler SHALL reject such requests with `ErrNoSandboxExecutor`. When `toolIsolation.enabled` is true and `container.enabled` is true, the app MUST attempt to create a `ContainerExecutor`. On failure, it MUST check `requireContainer`: if true, it MUST NOT fall back to `SubprocessExecutor` and MUST leave the sandbox executor nil. If `requireContainer` is false, it MUST fall back to `SubprocessExecutor` with a warning log.

#### Scenario: P2P enabled, toolIsolation disabled (default)
- **WHEN** `p2p.enabled=true` and `p2p.toolIsolation.enabled=false`
- **THEN** no sandbox executor SHALL be attached
- **AND** a startup warning SHALL be logged

#### Scenario: P2P enabled, toolIsolation enabled
- **WHEN** `p2p.enabled=true` and `p2p.toolIsolation.enabled=true`
- **THEN** a sandbox executor SHALL be attached (container or subprocess based on config)

#### Scenario: Container required but unavailable
- **WHEN** `requireContainer` is true and `NewContainerExecutor` fails
- **THEN** the sandbox executor SHALL remain nil
- **AND** the handler SHALL reject all P2P tool invocations with `ErrNoSandboxExecutor`

#### Scenario: Container optional and unavailable
- **WHEN** `requireContainer` is false and `NewContainerExecutor` fails
- **THEN** the system SHALL fall back to `SubprocessExecutor` with a warning log

### Requirement: Container pool
When `poolSize > 0`, the system MUST maintain a pool of pre-warmed containers with `Acquire`/`Release` lifecycle and idle timeout cleanup.

### Requirement: CLI sandbox commands
The system MUST provide `lango p2p sandbox status`, `lango p2p sandbox test`, and `lango p2p sandbox cleanup` commands.

#### Scenario: Sandbox status
- **WHEN** `lango p2p sandbox status` is run
- **THEN** it displays tool isolation config, container mode status, active runtime name, and pool info

#### Scenario: Sandbox test
- **WHEN** `lango p2p sandbox test` is run
- **THEN** it executes an echo tool through the sandbox and reports success/failure

#### Scenario: Sandbox cleanup
- **WHEN** `lango p2p sandbox cleanup` is run
- **THEN** orphaned containers with label `lango.sandbox=true` are removed

### Requirement: Sandbox Docker image
A `build/sandbox/Dockerfile` MUST define a minimal Debian-based image with the lango binary, running as non-root `sandbox` user with `--sandbox-worker` entrypoint.


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
