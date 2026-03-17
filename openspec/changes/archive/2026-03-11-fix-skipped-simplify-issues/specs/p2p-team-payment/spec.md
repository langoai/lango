## MODIFIED Requirements

### Requirement: Team budget bridge goroutine lifecycle
The `wireTeamBudgetBridge` function SHALL accept a `context.Context` parameter. Budget reservation timeout goroutines SHALL use a `select` on both the timer channel and `ctx.Done()` to ensure cleanup on application shutdown.

#### Scenario: Normal timeout releases reservation
- **WHEN** a budget reservation is made and the 5-minute timer expires before shutdown
- **THEN** the goroutine SHALL call `releaseFn()` via the timer path

#### Scenario: Shutdown cancels pending reservations
- **WHEN** the application context is cancelled (shutdown) before the 5-minute timer expires
- **THEN** the goroutine SHALL call `releaseFn()` via the `ctx.Done()` path, preventing access to a closed store
