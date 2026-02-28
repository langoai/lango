## MODIFIED Requirements

### Requirement: Background process management
The system SHALL support running commands in the background with process tracking. Background process output SHALL be thread-safe for concurrent read/write access.

#### Scenario: Background execution
- **WHEN** a command is started in background mode
- **THEN** a session ID SHALL be returned for later status checks

#### Scenario: Background process status
- **WHEN** status is requested for a background process
- **THEN** current output and execution state SHALL be returned

#### Scenario: Concurrent output access
- **WHEN** a background process is writing output while status is being read
- **THEN** the output buffer SHALL be safely accessible without data races
