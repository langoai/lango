## Context

When a Docker container receives SIGTERM, the application performs graceful shutdown by calling `httpServer.Shutdown()`. Go's standard library then causes the blocking `ListenAndServe()` call to return `http.ErrServerClosed`. Currently this is propagated up and logged at Error level, creating false alarm noise in production logs.

Additionally, resource cleanup errors during shutdown (browser, session store, graph store) are logged at Error level even though they occur at process exit and are non-actionable.

## Goals / Non-Goals

**Goals:**
- Eliminate false-positive error logs during normal graceful shutdown
- Distinguish between expected shutdown signals and genuine server errors
- Use appropriate log levels for shutdown-phase cleanup failures

**Non-Goals:**
- Changing shutdown behavior or timing
- Adding new shutdown hooks or cleanup logic
- Modifying the graceful shutdown flow itself

## Decisions

### 1. Filter `http.ErrServerClosed` at the gateway layer

The `gateway.Server.Start()` method will check for `http.ErrServerClosed` using `errors.Is()` and return `nil` instead. This keeps the responsibility in the correct package — the gateway knows what constitutes a normal vs abnormal server exit.

**Alternative considered**: Filtering in `app.go` at the call site. Rejected because it leaks HTTP-specific knowledge into the application layer.

### 2. Downgrade cleanup error logs to Warn

Shutdown-phase cleanup errors (gateway shutdown, browser close, session store close, graph store close) are downgraded from `Errorw` to `Warnw`. These errors occur at process exit and cannot be retried or acted upon.

**Alternative considered**: Suppressing these logs entirely. Rejected because cleanup failures may still be useful for post-mortem diagnosis.

## Risks / Trade-offs

- [Minimal risk] Genuine `ListenAndServe` errors that happen to coincide with `http.ErrServerClosed` could theoretically be masked → Mitigation: `http.ErrServerClosed` is only returned after `Shutdown()` is called, so this scenario cannot occur in practice.
- [Trade-off] Warn-level cleanup errors may be filtered out by some log aggregation setups → Mitigation: Any monitoring that needs these should include Warn level, which is standard practice.
