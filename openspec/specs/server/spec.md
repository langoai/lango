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

### Requirement: Serve command shutdown is deadline-bounded
The `lango serve` command SHALL attempt graceful shutdown on the first `SIGINT` or `SIGTERM`, and the shutdown sequence SHALL remain bounded by the configured serve shutdown deadline even when an individual lifecycle component blocks during stop.

#### Scenario: First interrupt starts graceful shutdown
- **WHEN** `lango serve` receives the first `SIGINT` or `SIGTERM`
- **THEN** it SHALL log that shutdown has started
- **AND** it SHALL call `application.Stop()` with the existing 10-second shutdown deadline

#### Scenario: Blocked component does not stall all shutdown forever
- **WHEN** a lifecycle-managed component blocks during `Stop(ctx)`
- **AND** the shutdown context reaches its deadline
- **THEN** the shutdown coordinator SHALL stop waiting on that component
- **AND** it SHALL continue attempting to stop the remaining started components
- **AND** it SHALL return control to the serve command without hanging indefinitely

#### Scenario: Shutdown progress is observable
- **WHEN** a lifecycle-managed component enters shutdown
- **THEN** the system SHALL log that the component is stopping
- **AND** it SHALL log whether the component stopped successfully or timed out

### Requirement: Second interrupt forces serve process exit
The `lango serve` command SHALL treat a second interrupt received during graceful shutdown as an explicit force-exit request.

#### Scenario: Second Ctrl+C forces exit
- **WHEN** graceful shutdown is already in progress for `lango serve`
- **AND** the process receives another `SIGINT`
- **THEN** the process SHALL terminate immediately with exit code `130`

#### Scenario: Graceful shutdown completes without force exit
- **WHEN** graceful shutdown finishes before a second interrupt is received
- **THEN** `lango serve` SHALL exit normally without using the force-exit path

### Requirement: Serve-connected shutdown handlers honor shutdown context
Shutdown handlers on the `lango serve` path that currently wait on internal goroutines SHALL support context-aware termination so they can return when the shutdown deadline is reached.

#### Scenario: Background manager shutdown respects deadline
- **WHEN** the background manager is asked to shut down with a context
- **AND** one or more internal task goroutines do not finish before the context deadline
- **THEN** the shutdown call SHALL return an error derived from the context instead of waiting indefinitely

#### Scenario: Workflow engine shutdown respects deadline
- **WHEN** the workflow engine is asked to shut down with a context
- **AND** one or more workflow goroutines do not finish before the context deadline
- **THEN** the shutdown call SHALL return an error derived from the context instead of waiting indefinitely

#### Scenario: Telegram stop interrupts update polling before waiting
- **WHEN** the Telegram channel is asked to stop during serve shutdown
- **THEN** it SHALL stop receiving updates before waiting for worker goroutines to exit
- **AND** the stop path SHALL return when the shutdown context is done instead of waiting indefinitely
