## ADDED Requirements

### Requirement: Shell command execution
The system SHALL execute shell commands in a controlled environment with configurable timeouts.

#### Scenario: Synchronous command execution
- **WHEN** a command is executed with a timeout
- **THEN** the system SHALL run the command and return stdout, stderr, and exit code

#### Scenario: Command timeout
- **WHEN** a command exceeds its timeout duration
- **THEN** the process SHALL be terminated and a timeout error returned

### Requirement: PTY support
The system SHALL support pseudo-terminal (PTY) mode for interactive commands.

#### Scenario: PTY command execution
- **WHEN** a command requires PTY (e.g., interactive prompts)
- **THEN** the system SHALL allocate a PTY and capture output

#### Scenario: ANSI escape handling
- **WHEN** PTY output contains ANSI escape codes
- **THEN** the codes SHALL be preserved for rendering or stripped as configured

### Requirement: Background process management
The system SHALL support running commands in the background with process tracking.

#### Scenario: Background execution
- **WHEN** a command is started in background mode
- **THEN** a session ID SHALL be returned for later status checks

#### Scenario: Background process status
- **WHEN** status is requested for a background process
- **THEN** current output and execution state SHALL be returned

### Requirement: Working directory control
The system SHALL execute commands in a specified working directory.

#### Scenario: Custom working directory
- **WHEN** a working directory is specified
- **THEN** the command SHALL execute relative to that directory

#### Scenario: Invalid working directory
- **WHEN** the specified directory does not exist
- **THEN** an error SHALL be returned before execution

### Requirement: Environment variable handling
The system SHALL control environment variables passed to child processes.

#### Scenario: Custom environment
- **WHEN** custom environment variables are specified
- **THEN** they SHALL be merged with or replace the base environment

#### Scenario: Dangerous variable filtering
- **WHEN** dangerous environment variables (LD_PRELOAD, etc.) are present
- **THEN** they SHALL be filtered out for security

#### Scenario: LANGO_PASSPHRASE filtered
- **WHEN** an agent executes a command and `LANGO_PASSPHRASE` is set in the parent environment
- **THEN** `LANGO_PASSPHRASE` is not passed to the child process

### Requirement: Enhanced execution feedback
The system SHALL provide more descriptive feedback when commands fail or time out.

#### Scenario: Detailed failure message
- **WHEN** a command fails with a non-zero exit code
- **THEN** the system SHALL return both stdout and stderr to the agent for debugging

### Requirement: Reference token resolution in exec
The exec tool SHALL resolve secret reference tokens in command strings immediately before execution. Resolved values SHALL never be logged or returned to the agent.

#### Scenario: Command with secret reference
- **WHEN** exec is called with command `curl -H "Auth: {{secret:api-key}}" https://api.example.com`
- **AND** the RefStore contains a value for `{{secret:api-key}}`
- **THEN** the token SHALL be replaced with the actual secret value before shell execution
- **AND** the log entry SHALL contain the original command with the unresolved token
- **AND** the BackgroundProcess.Command field SHALL contain the original command with the unresolved token

#### Scenario: Command with decrypt reference
- **WHEN** exec is called with command `echo {{decrypt:uuid-123}}`
- **AND** the RefStore contains a value for `{{decrypt:uuid-123}}`
- **THEN** the token SHALL be replaced with the actual decrypted value before shell execution

#### Scenario: Command with unknown reference
- **WHEN** exec is called with command `echo {{secret:unknown}}`
- **AND** the RefStore does NOT contain a value for `{{secret:unknown}}`
- **THEN** the literal string `{{secret:unknown}}` SHALL be passed to the shell unchanged

#### Scenario: Command without references
- **WHEN** exec is called with a command containing no reference tokens
- **THEN** the command SHALL be executed unchanged

#### Scenario: Reference resolution in PTY mode
- **WHEN** RunWithPTY is called with a command containing reference tokens
- **THEN** tokens SHALL be resolved identically to synchronous execution

#### Scenario: Reference resolution in background mode
- **WHEN** StartBackground is called with a command containing reference tokens
- **THEN** tokens SHALL be resolved identically to synchronous execution
