## MODIFIED Requirements

### Requirement: initOSSandbox logging
`initOSSandbox()` SHALL use `SandboxStatus` for structured logging. When the isolator is unavailable, log messages SHALL include the `Reason()` string and `PlatformCapabilities.Summary()`. Both fail-closed and fail-open paths SHALL log the reason.

#### Scenario: Fail-open logging with reason
- **WHEN** sandbox is enabled but isolator is unavailable with `failClosed=false`
- **THEN** log includes `reason` field with the isolator's `Reason()` value and `capabilities` field with `Summary()`

#### Scenario: Fail-closed logging with reason
- **WHEN** sandbox is enabled but isolator is unavailable with `failClosed=true`
- **THEN** log includes `reason` field with the isolator's `Reason()` value

### Requirement: Documentation accuracy
All code comments, doc comments, README, docs pages, and configuration references SHALL NOT claim Linux seccomp/Landlock enforcement when it is not implemented. Unimplemented features SHALL be marked as "planned" or "not yet enforced".

#### Scenario: Package doc comment
- **WHEN** reading the `sandbox/os` package documentation
- **THEN** it states Linux isolation is planned, not that it uses Landlock+seccomp

#### Scenario: Config field comments
- **WHEN** reading `SandboxConfig` field comments for Linux-specific behavior
- **THEN** they note Linux isolation is not yet enforced
