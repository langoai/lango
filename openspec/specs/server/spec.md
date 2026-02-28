# Server Spec

## Goal
Define requirements for the `lango serve` command and server capabilities.

## Requirements

### Requirement: Encryption Support
The application SHALL be able to open an encrypted SQLite database if the correct passphrase is provided.

#### Scenario: Server Startup with Passphrase
- **GIVEN** an encrypted session database exists at configured path
- **AND** `LANGO_PASSPHRASE` environment variable is set to the correct key
- **WHEN** `lango serve` is executed
- **THEN** the application starts successfully
- **AND** the session store is accessible (no "out of memory" or "not a database" errors)

### Requirement: Passphrase Configuration
The application SHALL prioritize the passphrase from environment variables over configuration files (standard security practice).

### Requirement: Path Expansion
The application SHALL verify that configuration paths using `~` are correctly expanded to the user's home directory.

#### Scenario: Tilde Expansion
- **GIVEN** `databasePath` is configured as `~/.lango/lango.db`
- **WHEN** the application initializes storage
- **THEN** it expands `~` to the current user's home directory
- **AND** successfully locates the file/directory

### Requirement: Shutdown cleanup errors logged at Warn level
During application shutdown, resource cleanup errors (gateway shutdown, browser close, session store close, graph store close) SHALL be logged at Warn level instead of Error level, since they occur at process exit and are non-actionable.

#### Scenario: Gateway shutdown error during stop
- **WHEN** `app.Stop()` is called and `Gateway.Shutdown()` returns an error
- **THEN** it SHALL log the error at Warn level (not Error level)

#### Scenario: Resource cleanup error during stop
- **WHEN** `app.Stop()` is called and browser close, session store close, or graph store close returns an error
- **THEN** each error SHALL be logged at Warn level (not Error level)

#### Scenario: Main shutdown handler error
- **WHEN** the main shutdown handler calls `application.Stop()` and it returns an error
- **THEN** it SHALL log at Warn level (not Error level)
