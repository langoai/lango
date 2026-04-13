## MODIFIED Requirements

### Requirement: Command execution
The exec tool SHALL execute shell commands via `sh -c` with configurable timeout, environment filtering, and optional OS-level sandbox isolation applied before process start.

#### Scenario: Run with sandbox enabled
- **WHEN** a command is executed via `Run()` with `Config.OSIsolator` set
- **THEN** the child process SHALL run under OS-level kernel restrictions per `Config.SandboxPolicy`

#### Scenario: Run with sandbox disabled
- **WHEN** a command is executed via `Run()` with `Config.OSIsolator` nil
- **THEN** the child process SHALL run without OS-level restrictions (existing behavior)

#### Scenario: Background process with sandbox
- **WHEN** a background command is started via `StartBackground()` with `Config.OSIsolator` set
- **THEN** the child process SHALL run under OS-level kernel restrictions
