## ADDED Requirements

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
