## MODIFIED Requirements

### Requirement: App wiring
When `p2p.toolIsolation.container.enabled` is true, the app MUST attempt to create a `ContainerExecutor`. On failure, it MUST check `requireContainer`: if true, it MUST NOT fall back to `SubprocessExecutor` and MUST leave the sandbox executor nil (causing the handler to reject P2P tool calls). If `requireContainer` is false, it MUST fall back to `SubprocessExecutor` with a warning log.

#### Scenario: Container required but unavailable
- **WHEN** `requireContainer` is true and `NewContainerExecutor` fails
- **THEN** the sandbox executor SHALL remain nil
- **AND** the handler SHALL reject all P2P tool invocations with `ErrNoSandboxExecutor`

#### Scenario: Container optional and unavailable
- **WHEN** `requireContainer` is false and `NewContainerExecutor` fails
- **THEN** the system SHALL fall back to `SubprocessExecutor` with a warning log
