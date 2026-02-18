## ADDED Requirements

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
